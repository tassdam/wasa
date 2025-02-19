package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
	"github.com/tassdam/wasa/service/database"
)

func (rt *_router) startConversation(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	var req struct {
		SenderID    string `json:"senderId"`
		RecipientID string `json:"recipientId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.SenderID == "" || req.RecipientID == "" {
		http.Error(w, "Missing senderId or recipientId", http.StatusBadRequest)
		return
	}
	conversationID, err := rt.db.GetDirectConversation(req.SenderID, req.RecipientID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to check conversation existence")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if conversationID == "" {
		conversationID, err = generateNewID()
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to generate conversation ID")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = rt.db.CreateDirectConversation(conversationID, req.SenderID, req.RecipientID)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to create new conversation")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"conversationId": conversationID,
	}); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode response")
	}
}

func (rt *_router) getConversation(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Missing conversationId", http.StatusBadRequest)
		return
	}
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isMember, err := rt.db.IsUserInConversation(conversationID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to check conversation membership")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
		return
	}
	if err := rt.db.MarkMessagesAsRead(conversationID, userID); err != nil {
		ctx.Logger.WithError(err).Error("Failed to mark messages as read")
	}
	conversation, err := rt.db.GetConversationDetails(conversationID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch conversation details")
		if errors.Is(err, database.ErrConversationDoesNotExist) {
			http.Error(w, "Conversation not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversation); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode conversation details")
	}
}

func (rt *_router) sendMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Missing conversationId", http.StatusBadRequest)
		return
	}
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	content := r.FormValue("content")
	replyTo := r.FormValue("replyTo")
	var attachment []byte
	file, header, err := r.FormFile("attachment")
	if err == nil {
		defer file.Close()
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
		}
		if !allowedTypes[header.Header.Get("Content-Type")] {
			http.Error(w, "Invalid file type. Only images and GIFs are allowed", http.StatusBadRequest)
			return
		}
		attachment, err = io.ReadAll(file)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to read attachment")
			http.Error(w, "Failed to process attachment", http.StatusInternalServerError)
			return
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		ctx.Logger.WithError(err).Error("Error retrieving file")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if content == "" && len(attachment) == 0 {
		http.Error(w, "Message content or attachment is required", http.StatusBadRequest)
		return
	}
	senderID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	messageID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate message ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	message, err := rt.db.SaveMessage(conversationID, senderID, messageID, content, attachment, replyTo)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to save message")
		if errors.Is(err, database.ErrConversationDoesNotExist) {
			http.Error(w, "Conversation does not exist", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	members, err := rt.db.GetConversationMembers(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch conversation members")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, memberID := range members {
		if memberID != senderID {
			if err := rt.db.InsertDeliveryReceipt(messageID, memberID, message.Timestamp); err != nil {
				ctx.Logger.WithError(err).Error("Failed to insert delivery receipt")
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(message); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode response")
	}
}

func (rt *_router) getMyConversations(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch user's conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversations); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode conversations")
	}
}

func (rt *_router) deleteMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if ok, err := rt.db.IsUserInConversation(conversationID, userID); !ok {
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to check conversation membership")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Forbidden: You are not a member of this conversation", http.StatusForbidden)
		return
	}

	err = rt.db.DeleteMessage(conversationID, messageID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to delete message")
		if errors.Is(err, database.ErrMessageDoesNotExist) {
			http.Error(w, "Message not found", http.StatusNotFound)
		} else if errors.Is(err, database.ErrUnauthorizedToDeleteMessage) {
			http.Error(w, "Forbidden: You are not the sender of this message", http.StatusForbidden)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rt *_router) forwardMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	messageID := ps.ByName("messageId")
	var req struct {
		TargetConversationID string `json:"targetConversationId"`
		ForwarderName        string `json:"forwarderName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	currentUserID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	originalMessage, err := rt.db.GetMessage(messageID, currentUserID)
	if err != nil {
		if errors.Is(err, database.ErrMessageDoesNotExist) {
			http.Error(w, "Message not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	newMessageID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate new message ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	newContent := "<strong>Forwarded from " + req.ForwarderName + ":</strong> " + originalMessage.Content
	newMessage := database.Message{
		Id:             newMessageID,
		ConversationId: req.TargetConversationID,
		SenderId:       currentUserID,
		SenderName:     req.ForwarderName,
		Content:        newContent,
		Timestamp:      time.Now().Format(time.RFC3339),
		Attachment:     originalMessage.Attachment,
	}
	if _, err := rt.db.SaveMessage(
		newMessage.ConversationId,
		newMessage.SenderId,
		newMessage.Id,
		newMessage.Content,
		newMessage.Attachment,
		"",
	); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save forwarded message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	members, err := rt.db.GetConversationMembers(req.TargetConversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch conversation members for forwarded message")
	} else {
		for _, memberID := range members {
			if memberID != currentUserID {
				if err := rt.db.InsertDeliveryReceipt(newMessageID, memberID, newMessage.Timestamp); err != nil {
					ctx.Logger.WithError(err).Error("Failed to insert delivery receipt for forwarded message")
				}
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}
