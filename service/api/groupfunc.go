package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
	"github.com/tassdam/wasa/service/database"
)

func (rt *_router) createGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	membersStr := r.FormValue("members")
	var members []string
	err = json.Unmarshal([]byte(membersStr), &members)
	if err != nil {
		http.Error(w, "Invalid members format", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()
	photo, err := io.ReadAll(file)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to read image file")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	conversationID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate conversation ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = rt.db.CreateGroupConversation(conversationID, members, name, photo)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to create new conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"conversationId": conversationID,
	}); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode response")
	}
}

func (rt *_router) getMyGroups(
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
	conversations, err := rt.db.GetMyGroups(userID)
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

func (rt *_router) getGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := ps.ByName("groupId")
	_, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	group, dbErr := rt.db.GetGroupInfo(groupID)
	if dbErr != nil {
		if errors.Is(dbErr, database.ErrGroupDoesNotExist) {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(dbErr).Error("Failed to fetch group details")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"id":      group.Id,
		"name":    group.Name,
		"members": group.Members,
	}
	if group.ConversationPhoto.Valid {
		response["groupPhoto"] = group.ConversationPhoto.String
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode group response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (rt *_router) setGroupName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	groupID := ps.ByName("groupId")
	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Invalid group name length", http.StatusBadRequest)
		return
	}
	dbErr := rt.db.UpdateGroupName(groupID, req.Name)
	if errors.Is(dbErr, database.ErrUserDoesNotExist) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to update username")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (rt *_router) setGroupPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := ps.ByName("groupId")
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err = r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, "Failed to parse form. Ensure the file is below 10 MB.", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Failed to retrieve photo file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	photoData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read photo file", http.StatusInternalServerError)
		return
	}
	if len(photoData) > 10*1024*1024 {
		http.Error(w, "Photo too large. Maximum allowed size is 10 MB.", http.StatusRequestEntityTooLarge)
		return
	}
	fileType := http.DetectContentType(photoData)
	if fileType != "image/jpeg" && fileType != "image/png" {
		http.Error(w, "Invalid file type. Only JPEG and PNG are supported.", http.StatusUnsupportedMediaType)
		return
	}
	err = rt.db.UpdateGroupPhoto(groupID, photoData)
	if errors.Is(err, database.ErrUserDoesNotExist) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		ctx.Logger.WithError(err).Error("Failed to update user photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"message": "Photo updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode photo update response")
	}
}

func (rt *_router) leaveGroup(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	groupID := ps.ByName("groupId")
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err = rt.db.LeaveGroup(groupID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to leave group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	_, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var request struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	err = rt.db.AddUserToGroup(groupID, request.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to add user to group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
