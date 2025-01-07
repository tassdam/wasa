package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
	"github.com/tassdam/wasa/service/database"
)

// LoginRequest matches the schema defined in your OpenAPI specification

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Ensure the method is POST. If checked by middleware already, this can be skipped.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body into a LoginRequest struct
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the username
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Invalid username length", http.StatusBadRequest)
		return
	}

	// Try to get the user by name
	user, err := rt.db.GetUserByName(req.Name)
	if err == database.ErrUserDoesNotExist {
		// User does not exist, create a new one
		newID, genErr := generateNewID()
		if genErr != nil {
			ctx.Logger.WithError(genErr).Error("Failed to generate user ID")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		newUser := database.User{
			Id:   newID,
			Name: req.Name,
		}
		createdUser, createErr := rt.db.CreateUser(newUser)
		if createErr != nil {
			ctx.Logger.WithError(createErr).Error("cannot create user")
			http.Error(w, "Internal Server Error: cannot create user", http.StatusInternalServerError)
			return
		}
		user = createdUser
	} else if err != nil {
		// Some other error from the database
		ctx.Logger.WithError(err).Error("error retrieving user")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := LoginResponse{
		Identifier: user.Id,
	}

	// According to the OpenAPI spec, successful login action is '201'
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Println("TEST")
	json.NewEncoder(w).Encode(resp)
}

// generateNewID uses github.com/gofrs/uuid to generate a unique user ID.
func generateNewID() (string, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}
