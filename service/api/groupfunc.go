package api

import (
	"encoding/base64"
	"encoding/json"
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
	// Parse multipart form data with a memory limit of 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	membersStr := r.FormValue("members")

	// Unmarshal "members" JSON string into a slice of strings
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

	// Send the response with the conversation ID
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

	// Return the conversations as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversations); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode conversations")
	}
}

func (rt *_router) getGroupPhoto(
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

	group, dbErr := rt.db.GetGroupPhoto(groupID)

	if dbErr == database.ErrGroupDoesNotExist {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("Failed to fetch group details")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{}

	if group.Photo != nil {
		response["groupPhoto"] = base64.StdEncoding.EncodeToString(group.Photo)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode user response")
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

	if dbErr == database.ErrUserDoesNotExist {
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
	file, fileHeader, err := r.FormFile("photo")
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
	if len(photoData) > 10*1024*1024 { // 10 MB
		http.Error(w, "Photo too large. Maximum allowed size is 10 MB.", http.StatusRequestEntityTooLarge)
		return
	}
	fileType := http.DetectContentType(photoData)
	if fileType != "image/jpeg" && fileType != "image/png" {
		http.Error(w, "Invalid file type. Only JPEG and PNG are supported.", http.StatusUnsupportedMediaType)
		return
	}
	ctx.Logger.Infof("Received file: %s, size: %d bytes, type: %s", fileHeader.Filename, len(photoData), fileType)

	err = rt.db.UpdateGroupPhoto(groupID, photoData)

	if err == database.ErrUserDoesNotExist {
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
