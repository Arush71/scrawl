package api

import (
	"log/slog"
	"net/http"

	"github.com/Arush71/scrawl/internal/helpers"
	"github.com/Arush71/scrawl/internal/ws"
	"github.com/coder/websocket"
)

type Handler struct {
	Logger   *slog.Logger
	Registry *ws.Registry
}

func (h *Handler) handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("username")
	gameID := r.URL.Query().Get("gameId")
	if gameID == "" || name == "" {
		helpers.BadRequestError(w)
		return
	}
	if !h.Registry.CheckRoom(gameID) {
		helpers.Error(w, http.StatusNotFound, "room not found")
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		h.Logger.Debug("failed to accept connection", "err", err.Error())
		return
	}
	defer conn.CloseNow()
	h.Registry.HandleConnections(gameID, name, conn)
}

func (h *Handler) handleRoomCreation(w http.ResponseWriter, r *http.Request) {
	roomID := h.Registry.CreateRoom()
	type response struct {
		RoomID string `json:"roomId"`
	}
	helpers.WriteJSON(w, http.StatusCreated, response{
		RoomID: roomID,
	})
}
