package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
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
