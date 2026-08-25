package ws

import (
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Player struct {
	gameRoom *gameRoom
	playerID uuid.UUID
	conn     *websocket.Conn
	username string
	send     chan []byte
}

func (p *Player) removePlayer() {
	p.conn.Close(websocket.StatusPolicyViolation, "connection too slow")
}

type playersT map[uuid.UUID]*Player

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
