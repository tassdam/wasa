package api

import (
	"encoding/json"
	"errors"
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
	// Parse the request body to extract senderId and recipientId
	var req struct {
		SenderID    string `json:"senderId"`
		RecipientID string `json:"recipientId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.SenderID == "" || req.RecipientID == "" {
		http.Error(w, "Missing senderId or recipientId", http.StatusBadRequest)
		return
	}

	// Check if the conversation exists
	conversationID, err := rt.db.GetDirectConversation(req.SenderID, req.RecipientID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to check conversation existence")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If no conversation exists, create one
	if conversationID == "" {
		conversationID, err = generateNewID() // Use your existing UUID generation function
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

	// Respond with the conversation ID
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

	conversation, err := rt.db.GetConversationDetails(conversationID)
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
	// Extract conversationId from the route parameters
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Missing conversationId", http.StatusBadRequest)
		return
	}

	// Parse the request body
	var req struct {
		Content          string  `json:"content"`
		ForwardedMessage *string `json:"forwardedMessageId,omitempty"`
		Attachment       []byte  `json:"attachment,omitempty"` // Optional for photos/GIFs
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Content == "" && req.Attachment == nil {
		http.Error(w, "Message content or attachment is required", http.StatusBadRequest)
		return
	}

	// Get the sender's ID from the Authorization header
	senderID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate a new message ID
	messageID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate message ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Save the message to the database
	message, err := rt.db.SaveMessage(
		conversationID,
		senderID,
		messageID,
		req.Content,
		req.ForwardedMessage,
		req.Attachment,
	)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to save message")
		if errors.Is(err, database.ErrConversationDoesNotExist) {
			http.Error(w, "Conversation does not exist", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Insert delivery receipts for all conversation members
	members, err := rt.db.GetConversationMembers(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch conversation members")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, memberID := range members {
		if memberID != senderID { // Exclude the sender from receipts
			if err := rt.db.InsertDeliveryReceipt(messageID, memberID, message.Timestamp); err != nil {
				ctx.Logger.WithError(err).Error("Failed to insert delivery receipt")
			}
		}
	}

	// Respond with the saved message
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
	// Extract user ID from the Authorization header
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch the conversations for the user from the database
	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch user's conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return the conversations as JSON
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
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	originalMessage, err := rt.db.GetMessage(messageID, userID)
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
	newMessage := database.Message{
		Id:               newMessageID,
		ConversationId:   req.TargetConversationID,
		SenderId:         originalMessage.SenderId,
		Content:          originalMessage.Content,
		Timestamp:        time.Now().Format(time.RFC3339),
		ForwardedMessage: &originalMessage.Id,
	}

	if _, err := rt.db.SaveMessage(
		newMessage.ConversationId,
		newMessage.SenderId,
		newMessage.Id,
		newMessage.Content,
		newMessage.ForwardedMessage,
		nil,
	); err != nil {
		ctx.Logger.WithError(err).Error("Failed to save forwarded message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
