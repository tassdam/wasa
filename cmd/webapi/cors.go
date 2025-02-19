package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"content-type", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "X-Requested-With", "Authorization",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(1),
	)(h)
}
