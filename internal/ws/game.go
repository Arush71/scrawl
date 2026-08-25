package ws

import (
	"fmt"
	"sync"

	"github.com/Arush71/scrawl/internal/protocol"
	"github.com/google/uuid"
)

type gameRoom struct {
	mu      sync.RWMutex
	roomID  string
	players playersT
}

func (g *gameRoom) broadcastJoin(username string, ID uuid.UUID) error {
	data, err := protocol.Encode(protocol.TypePlayerJoined, protocol.PlayerEvent{
		Username: username,
		PlayerID: ID,
	})
	if err != nil {
		return fmt.Errorf("marshal write message: %w", err)
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.players.broadcast(data)
	return nil
}

func (g *gameRoom) broadcastLeave(username string, ID uuid.UUID) {
	data, err := protocol.Encode(protocol.TypePlayerLeft, protocol.PlayerEvent{
		Username: username,
		PlayerID: ID,
	})
	if err != nil {
		return
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.players.broadcast(data)
}
