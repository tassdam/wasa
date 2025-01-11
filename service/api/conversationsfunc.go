package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
)

func (rt *_router) getMyConversations(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	// 1) Extract user ID from the Authorization header
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Query the DB
	convs, dbErr := rt.db.GetMyConversations(userID)
	if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to retrieve user conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3) Encode as JSON
	w.Header().Set("Content-Type", "application/json")
	// Typically 200 OK
	if err := json.NewEncoder(w).Encode(convs); err != nil {
		ctx.Logger.WithError(err).Error("failed to encode getMyConversations result")
	}
}
