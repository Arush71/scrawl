// Package api is about the api
package api

import (
	"net/http"
)

func AddRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /joinRoom", handler.handleJoinRoom)
	mux.HandleFunc("POST /createRoom", handler.handleRoomCreation)
}
