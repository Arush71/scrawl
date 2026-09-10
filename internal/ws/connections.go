package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/Arush71/scrawl/internal/protocol"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

func (r *Registry) HandleConnections(gameID string, username string, conn *websocket.Conn) {
	p := &Player{
		playerID: uuid.New(),
		username: username,
		conn:     conn,
		send:     make(chan []byte, 8),
	}
	gameRoom, err := r.attachPlayer(gameID, p)
	if err != nil {
		conn.Close(CloseRoomNotFound, "room does not exist")
		return
	}
	p.gameRoom = gameRoom
	defer r.Release(p)
	if err := p.gameRoom.broadcastJoin(username, p.playerID); err != nil {
		r.logger.Error("failed to broadcast join", "error", err)
		return
	}
	r.logger.Info("new connection!")
	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()
	go p.writeLoop(ctx, r.logger)
	go heartBeat(ctx, conn)
	r.readConnection(ctx, p)
}

func heartBeat(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-time.After(time.Second * 5):
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
				if websocket.CloseStatus(err) != -1 || errors.Is(err, context.Canceled) {
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

func (r *Registry) readConnection(ctx context.Context, player *Player) {
	defer player.gameRoom.broadcastLeave(player.username, player.playerID)
	for {
		_, data, err := player.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				return
			}
			r.logger.Error("Read connection failed", "error", err.Error())
			return
		}
		envelope, err := protocol.Decode(data)
		if err != nil {
			r.logger.Debug("unmaershalling failed, invalid data", "data", string(data), "error", err.Error())
			continue
		}
		err = r.handleReqData(envelope, player)
		if err != nil {
			r.logger.Error("error while handling req", "error", err.Error(), "username", player.username)
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
		return handleMessage(message, player)
	case protocol.TypeStartGame:
		player.gameRoom.mu.Lock()
		if player.gameRoom.gameOwner != player.playerID || player.gameRoom.gameState != Waiting {
			player.gameRoom.mu.Unlock()
			return protocol.ErrInvalidProtocol
		}
		player.gameRoom.gameState = WordSelection
		player.gameRoom.orderedPlayers()
		drawerID, _ := player.gameRoom.nextDrawer()
		totalRounds := player.gameRoom.totalRounds // TODO: make this configurable later
		player.gameRoom.currentRound = 1
		player.gameRoom.mu.Unlock()

		go player.gameRoom.startGame(drawerID, totalRounds)
		return nil
	case protocol.TypeSelectedWord:
		player.gameRoom.mu.RLock()
		if player.gameRoom.gameState != WordSelection || player.gameRoom.playerOrder[player.gameRoom.currentDrawerIdx] != player.playerID {
			player.gameRoom.mu.RUnlock()
			return protocol.ErrInvalidProtocol
		}
		player.gameRoom.mu.RUnlock()
		var data protocol.WordSelected
		if err := json.Unmarshal(d.Data, &data); err != nil || data.Word == "" {
			return protocol.ErrInvalidProtocol
		}
		select {
		case player.gameRoom.currentWordCh <- data.Word:
			return nil
		default:
			return protocol.ErrInvalidProtocol
		}
	default:
		return protocol.ErrInvalidProtocol
	}
}
