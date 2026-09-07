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
	mu               sync.RWMutex
	gameOwner        uuid.UUID
	roomID           string
	players          playersT
	gameState        gameState
	playerOrder      []uuid.UUID
	currentDrawerIdx int // index of the current drawer in playerOrder
	round            int
	totalRounds      int
	currentWord      string
	currentWordCh    chan string // unbuffered channel to receive the selected word from the drawer
	nextSeq          int
	guessedPlayers   map[uuid.UUID]struct{} // set to track players who have guessed the word
	guessListner     chan struct{}          // channel to signal when all players have guessed the word
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
		g.currentDrawerIdx = (g.currentDrawerIdx + 1) % len(g.playerOrder)
		id := g.playerOrder[g.currentDrawerIdx]
		if _, ok := g.players[id]; ok {
			return id
		}
	}
}

func (g *gameRoom) startGame(drawerID uuid.UUID) {
	g.broadcastPhase(WordSelection, drawerID)
	g.waitForWordSelection() // can run till 15 seconds max
	g.mu.Lock()
	g.gameState = Drawing
	g.mu.Unlock()
	g.broadcastPhase(Drawing, drawerID)
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
		dataPhase, err := protocol.Encode(protocol.TypePhasePick, protocol.WritePhasePick{
			DrawerID: drawerID,
		})
		dataPick, err2 := protocol.Encode(protocol.TypePickWord, protocol.WritePickWord{})
		if err != nil || err2 != nil {
			return
		}
		g.mu.RLock()
		g.players.broadcastExtra(dataPhase, dataPick, drawerID)
		g.mu.RUnlock()
	case Drawing:
		g.mu.RLock()
		wordLen := len(g.currentWord)
		g.mu.RUnlock()
		data, err := protocol.Encode(protocol.TypePhaseDraw, protocol.WritePhaseDraw{
			DrawerID:   drawerID,
			TimeLimit:  80,
			WordLength: wordLen,
		})
		if err != nil {
			return
		}
		g.mu.RLock()
		g.players.broadcast(data)
		g.mu.RUnlock()
	}
}

func (g *gameRoom) waitForWordSelection() {
	select {
	case word := <-g.currentWordCh:
		g.mu.Lock()
		defer g.mu.Unlock()
		g.currentWord = word
	case <-time.After(15 * time.Second):
		g.mu.Lock()
		defer g.mu.Unlock()
		g.currentWord = "apple" // default word if no selection is made
	}
}
