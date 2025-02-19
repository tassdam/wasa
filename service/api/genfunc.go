package api

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) Close() error {
	return nil
}

func (rt *_router) liveness(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := rt.db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func generateNewID() (string, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}
