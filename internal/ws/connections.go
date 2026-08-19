package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Arush71/scrawl/internal/protocol"
	"github.com/coder/websocket"
)

func (r *Registry) HandleConnections(username string, conn *websocket.Conn) {
	p := &Player{
		username: username,
		conn:     conn,
		send:     make(chan []byte, 16),
	}
	defer r.Release(username)
	r.Attach(username, p)
	r.logger.Info("new connection!")
	go p.writeLoop(r.ctx, r.logger)
	readConnection(r.ctx, r.logger, p)
}

func (p *Player) writeLoop(ctx context.Context, logger *slog.Logger) {
	for {
		select {
		case data, ok := <-p.send:
			if !ok {
				return
			}
			if err := p.conn.Write(ctx, websocket.MessageText, data); err != nil {
				logger.Error("Write error encounterd", "error", err.Error())
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func readConnection(ctx context.Context, logger *slog.Logger, player *Player) {
	for {
		_, data, err := player.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				return
			}
			logger.Error("Read connection failed", "error", err.Error())
			return
		}
		baseData, err := decodeData(data)
		if err != nil {
			logger.Debug("unmaershalling failed, invalid data", "data", string(data), "error", err.Error())
			continue
		}
		err = handleReqData(baseData, player)
		if err != nil {
			logger.Debug("invalid type")
			continue
		}
	}
}

func decodeData(d []byte) (protocol.BaseStructure, error) {
	var baseStruct protocol.BaseStructure
	if err := json.Unmarshal(d, &baseStruct); err != nil {
		return protocol.BaseStructure{}, err
	}
	return baseStruct, nil
}

func handleReqData(d protocol.BaseStructure, player *Player) error {
	switch d.Type {
	case protocol.TypeChat:
		var message protocol.Message
		if err := json.Unmarshal(d.Data, &message); err != nil || message.Text == "" {
			return protocol.ErrInvalidProtocol
		}
		return player.handleMessage(message)
	case protocol.TypeJoinChat:
		return nil
	default:
		return protocol.ErrInvalidProtocol
	}
}
