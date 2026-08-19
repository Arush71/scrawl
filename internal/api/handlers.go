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

func (h *Handler) handleConnection(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		helpers.BadRequestError(w)
		return
	}
	if !h.Registry.TryClaim(username) {
		helpers.Error(w, http.StatusConflict, "username is already taken")
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		h.Registry.Release(username)
		h.Logger.Debug("failed to accept connection", "err", err.Error())
		return
	}
	defer conn.CloseNow()
	ws.HandleConnection(username, conn)
}
