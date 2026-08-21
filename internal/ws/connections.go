package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Arush71/scrawl/internal/protocol"
	"github.com/coder/websocket"
)

func (r *Registry) HandleConnections(username string, conn *websocket.Conn) {
	p := &Player{
		username: username,
		conn:     conn,
		send:     make(chan []byte, 16),
	}
	r.Attach(username, p)
	defer r.Release(username)
	r.logger.Info("new connection!")
	if err := r.broadcastJoin(username); err != nil {
		r.logger.Error("failed to broadcast join", "error", err)
		return
	}
	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()
	go p.writeLoop(ctx, r.logger)
	go heartBeat(ctx, conn)
	r.readConnection(ctx, r.logger, p)
}

func heartBeat(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-time.After(time.Second * 4):
			newCtx, cancel := context.WithTimeout(ctx, time.Second*5)
			if err := conn.Ping(newCtx); err != nil {
				_ = conn.CloseNow()
				cancel()
				return
			}
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

func (p *Player) writeLoop(ctx context.Context, logger *slog.Logger) {
	for {
		select {
		case data, ok := <-p.send:
			if !ok {
				return
			}
			if err := p.conn.Write(ctx, websocket.MessageText, data); err != nil {
				if websocket.CloseStatus(err) != -1 {
					return
				}
				logger.Error("Write error encounterd", "error", err.Error())
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *Registry) readConnection(ctx context.Context, logger *slog.Logger, player *Player) {
	defer r.broadcastLeave(player.username)
	for {
		_, data, err := player.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				return
			}
			logger.Error("Read connection failed", "error", err.Error())
			return
		}
		envelope, err := protocol.Decode(data)
		if err != nil {
			logger.Debug("unmaershalling failed, invalid data", "data", string(data), "error", err.Error())
			continue
		}
		err = r.handleReqData(envelope, player)
		if err != nil {
			logger.Error("error while handling req", "error", err.Error(), "username", player.username)
			continue
		}
	}
}

func (r *Registry) handleReqData(d protocol.Envelope, player *Player) error {
	switch d.Type {
	case protocol.TypeChat:
		var message protocol.ChatPayload
		if err := json.Unmarshal(d.Data, &message); err != nil || message.Text == "" {
			return protocol.ErrInvalidProtocol
		}
		return r.handleMessage(message, player)
	default:
		return protocol.ErrInvalidProtocol
	}
}
