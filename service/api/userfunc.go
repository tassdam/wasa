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

var ErrUnauthorized = errors.New("unauthorized request")

// setMyUserName handles PUT /users/me to update the authenticated user’s username.
// Example usage in the router: rt.router.PUT("/users/me", rt.wrap(rt.setMyUserName))
func (rt *_router) setMyUserName(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1. Enforce correct HTTP method (optional if your router already does this).
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Get the authenticated user ID from the request.
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Parse the request body into UpdateUserRequest.
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 4. Validate the new username if needed.
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Invalid username length", http.StatusBadRequest)
		return
	}

	// 5. Update username in the database.
	updatedUser, dbErr := rt.db.UpdateUserName(userID, req.Name)
	if dbErr == database.ErrUserDoesNotExist {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to update username")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 6. Return updated user info as JSON (adjust structure to match your OpenAPI).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode updated user response")
		// Decide if you want to silently fail or send another error response
	}
}

// setMyPhoto handles PUT /users/me/photo to update the authenticated user’s photo.
// Example usage: rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))
func (rt *_router) setMyPhoto(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1. Enforce correct HTTP method.
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Get the authenticated user ID from the request.
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. Read the entire photo (binary data) from the request body.
	photo, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read photo", http.StatusBadRequest)
		return
	}

	// 4. Optional: Validate photo content size/type if needed.
	if len(photo) == 0 {
		http.Error(w, "Empty photo", http.StatusBadRequest)
		return
	}
	// Suppose your OpenAPI says max 10 MB:
	if len(photo) > 10*1024*1024 {
		http.Error(w, "Photo too large", http.StatusRequestEntityTooLarge)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" {
		http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	// 5. Update the photo in the database.
	if dbErr := rt.db.UpdateUserPhoto(userID, photo); dbErr == database.ErrUserDoesNotExist {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to update user photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 6. Respond with a success message. Adjust structure to match your OpenAPI.
	response := map[string]string{
		"message": "Photo updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode photo update response")
		// Decide if you want to handle this error more thoroughly.
	}
}

// getAuthenticatedUserID is a helper function to retrieve the user ID from the request.
// This is just an example; your actual auth logic may differ (e.g., Bearer tokens).
func (rt *_router) getAuthenticatedUserID(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", ErrUnauthorized // define a custom ErrUnauthorized if needed
	}
	userID := authHeader[7:]
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

// searchUsers handles GET /users/search to search for users by a partial username match.
func (rt *_router) searchUsers(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1. Get the search query from the URL
	query := r.URL.Query().Get("username")
	if query == "" {
		http.Error(w, "Missing 'username' query parameter", http.StatusBadRequest)
		return
	}

	// 2. Query the database for matching users
	users, err := rt.db.SearchUsersByName(query)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to search users")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]User{}) // Return an empty array instead of null
		return
	}

	// 3. Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
