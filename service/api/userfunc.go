package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
	"github.com/tassdam/wasa/service/database"
)

var ErrUnauthorized = errors.New("unauthorized request")

func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Invalid username length", http.StatusBadRequest)
		return
	}
	updatedUser, dbErr := rt.db.UpdateUserName(userID, req.Name)
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
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode updated user response")
	}
}

func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, err := rt.getAuthenticatedUserID(r)
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
	err = rt.db.UpdateUserPhoto(userID, photoData)
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

func (rt *_router) getAuthenticatedUserID(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", ErrUnauthorized
	}
	userID := authHeader[7:]
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

func (rt *_router) searchUsers(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	query := r.URL.Query().Get("username")
	if query == "" {
		http.Error(w, "Missing 'username' query parameter", http.StatusBadRequest)
		return
	}
	users, err := rt.db.SearchUsersByName(query)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to search users")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode([]User{}); err != nil {
			ctx.Logger.WithError(err).Error("Failed to encode empty users array")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode users response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (rt *_router) getMyPhoto(
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
	user, dbErr := rt.db.GetUsersPhoto(userID)
	if errors.Is(dbErr, database.ErrUserDoesNotExist) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("Failed to fetch user details")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"name":  user.Name,
		"photo": nil,
	}
	if user.Photo != nil {
		response["photo"] = base64.StdEncoding.EncodeToString(user.Photo)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode user response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
