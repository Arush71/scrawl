package ws

import (
	"context"
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
)

type gameRoom struct {
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	gameOwner        uuid.UUID
	roomID           string
	players          playersT
	gameState        gameState
	playerOrder      []uuid.UUID
	currentDrawerIdx int // index of the current drawer in playerOrder
	currentRound     int
	totalRounds      int
	currentWord      string
	currentWordCh    chan string // unbuffered channel to receive the selected word from the drawer
	nextSeq          int
	guessedPlayers   map[uuid.UUID]struct{} // set to track players who have guessed the word
	guessListner     chan struct{}          // channel to signal when all players have guessed the word
}

func NewGameRoom(roomID string, ctx context.Context) *gameRoom {
	g := &gameRoom{
		mu:               sync.RWMutex{},
		roomID:           roomID,
		players:          make(playersT),
		gameState:        Waiting,
		nextSeq:          0,
		currentWordCh:    make(chan string),
		guessedPlayers:   make(map[uuid.UUID]struct{}),
		guessListner:     make(chan struct{}),
		currentDrawerIdx: -1,
		currentRound:     -1,
		totalRounds:      3, // default to 3 rounds, can be changed later
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	return g
}

// NOTE: need to hold a Lock on the gameRoom before calling this function
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

// NOTE: need to hold a Lock on the gameRoom before calling this function
func (g *gameRoom) nextDrawer() (uuid.UUID, bool) {
	roundComplete := false
	lenOrder := len(g.playerOrder)
	for range lenOrder {
		g.currentDrawerIdx = (g.currentDrawerIdx + 1) % len(g.playerOrder) // should init IDx to -1
		if g.currentDrawerIdx == 0 {
			roundComplete = true
		}
		id := g.playerOrder[g.currentDrawerIdx]
		if _, ok := g.players[id]; ok {
			return id, roundComplete
		}
	}
	return uuid.UUID{}, false
}

func (g *gameRoom) startGame(drawerID uuid.UUID, totalRounds int) {
	for range totalRounds {
		if g.ctx.Err() != nil {
			return
		}
		isRoundComplete := false
		for {
			g.broadcastPhase(WordSelection, drawerID)
			g.waitForWordSelection() // can run till 15 seconds max
			if g.ctx.Err() != nil {
				return
			}
			g.mu.Lock()
			g.gameState = Drawing
			g.mu.Unlock()
			g.broadcastPhase(Drawing, drawerID)
			g.waitForWordGuess() // can run till 80 seconds max
			if g.ctx.Err() != nil {
				return
			}

			g.mu.Lock()
			drawerID, isRoundComplete = g.nextDrawer()
			g.gameState = WordSelection
			g.guessedPlayers = make(map[uuid.UUID]struct{})
			if isRoundComplete {
				g.roundComplete()
				g.mu.Unlock()
				break
			}
			g.mu.Unlock()
		}
	}
	g.gameOver()
}

func (g *gameRoom) gameOver() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cleanUpOnEnd()
	data, err := protocol.Encode(protocol.TypeGameOver, protocol.WriteGameOver{})
	if err != nil {
		return
	}
	g.players.broadcast(data)
	g.gameState = Waiting
}

// NOTE: Need to hold a Lock
func (g *gameRoom) roundComplete() {
	data, err := protocol.Encode(protocol.TypeRoundComplete, protocol.WriteRoundComplete{
		RoundCompleted: g.currentRound,
	})
	if err != nil {
		return
	}
	g.currentRound++
	g.players.broadcast(data)
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
	case <-g.ctx.Done():
		return
	}
}

func (g *gameRoom) waitForWordGuess() {
	allGuessed := true
	select {
	case <-g.guessListner:
	case <-time.After(80 * time.Second):
		allGuessed = false
	case <-g.ctx.Done():
		return
	}
	g.mu.RLock()
	word := g.currentWord
	g.mu.RUnlock()

	data, err := protocol.Encode(protocol.TypeRotationEnd, protocol.WriteRotationEnd{
		Word:       word,
		AllGuessed: allGuessed,
	})
	if err != nil {
		return
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.players.broadcast(data)
}

// NOTE: Need to hold a Lock
func (g *gameRoom) cleanUpOnEnd() {
	g.currentWord = ""
	g.playerOrder = make([]uuid.UUID, 0)
	g.currentRound = -1
	g.currentDrawerIdx = -1
}
