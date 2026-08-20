package ws

import (
	"fmt"

	"github.com/Arush71/scrawl/internal/protocol"
	"github.com/coder/websocket"
)

type Player struct {
	conn     *websocket.Conn
	username string
	send     chan []byte
}

func (p *Player) removePlayer() {
	p.conn.Close(websocket.StatusPolicyViolation, "connection too slow")
}

type playersT map[string]*Player

// NOTE: Should be called in a read lock
func (pt playersT) broadcast(text []byte) {
	for _, player := range pt {
		select {
		case player.send <- text:
		default:
			go player.removePlayer()
		}
	}
}

func (r *Registry) broadcastJoin(username string) error {
	data, err := protocol.Encode(protocol.TypePlayerJoined, protocol.PlayerEvent{
		Username: username,
	})
	if err != nil {
		return fmt.Errorf("marshal write message: %w", err)
	}
	r.playerMu.RLock()
	defer r.playerMu.RUnlock()
	r.players.broadcast(data)
	return nil
}

func (r *Registry) broadcastLeave(username string) {
	data, err := protocol.Encode(protocol.TypePlayerLeft, protocol.PlayerEvent{
		Username: username,
	})
	if err != nil {
		return
	}
	r.playerMu.RLock()
	defer r.playerMu.RUnlock()
	r.players.broadcast(data)
}
