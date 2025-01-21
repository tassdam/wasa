package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/tassdam/wasa/service/api/reqcontext"
)

func (rt *_router) commentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)

	if err != nil {
		http.Error(w, "Unauthorized: ", http.StatusUnauthorized)
		return
	}

	commentID, err := generateNewID()

	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate conversation ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := rt.db.CommentMessage(commentID, ps.ByName("messageId"), userID); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) uncommentMessage(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
	ctx reqcontext.RequestContext,
) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)

	if err != nil {
		http.Error(w, "Unauthorized: ", http.StatusUnauthorized)
		return
	}

	if err := rt.db.UncommentMessage(ps.ByName("messageId"), userID); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
