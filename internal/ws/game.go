package ws

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Arush71/scrawl/internal/protocol"
	"github.com/google/uuid"
)

type gameState string

const (
	Waiting       gameState = "waiting"
	WordSelection gameState = "wordSelection"
	Drawing       gameState = "Drawing"
	GameOver      gameState = "gameOver"
)

type gameRoom struct {
	mu            sync.RWMutex
	gameOwner     uuid.UUID
	roomID        string
	players       playersT
	gameState     gameState
	playerOrder   []uuid.UUID
	currentDrawer int // index of the current drawer in playerOrder
	round         int
	totalRounds   int
	currentWord   string
	nextSeq       int
}

// NOTE: neet to hold a Lock on the gameRoom before calling this function
func (g *gameRoom) orderedPlayers() {
	totalPlayers := len(g.players)
	players := make([]uuid.UUID, 0, totalPlayers)
	for playerID := range g.players {
		players = append(players, playerID)
	}
	sort.Slice(players, func(i, j int) bool {
		return g.players[players[i]].joinSeq < g.players[players[j]].joinSeq
	})
	g.playerOrder = players
}

// NOTE: neet to hold a Lock on the gameRoom before calling this function
func (g *gameRoom) nextDrawer() uuid.UUID {
	for {
		g.currentDrawer = (g.currentDrawer + 1) % len(g.playerOrder)
		id := g.playerOrder[g.currentDrawer]
		if _, ok := g.players[id]; ok {
			return id
		}
	}
}

func (g *gameRoom) startGame() {
	g.mu.Lock()
	g.gameState = WordSelection
	g.orderedPlayers()
	drawerID := g.nextDrawer()
	g.mu.Unlock()
	go g.broadcastPhase(WordSelection, drawerID)
	time.AfterFunc(15*time.Second, func() {})
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

func (g *gameRoom) broadcastPhase(state gameState, drawerID uuid.UUID) {
	switch state {
	case WordSelection:
	}
}
