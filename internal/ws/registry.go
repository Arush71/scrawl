// Package ws is for websockets and players
package ws

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/Arush71/scrawl/internal/helpers"
	"github.com/Arush71/scrawl/internal/protocol"
)

type Registry struct {
	gameRooms map[string]*gameRoom
	roomMu    sync.RWMutex
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewRegistry(logger *slog.Logger) *Registry {
	r := &Registry{
		gameRooms: make(map[string]*gameRoom),
		roomMu:    sync.RWMutex{},
		logger:    logger,
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	return r
}

func (r *Registry) CheckRoom(roomID string) bool {
	r.roomMu.RLock()
	defer r.roomMu.RUnlock()
	_, ok := r.gameRooms[roomID]
	return ok
}

func (r *Registry) attachPlayer(roomID string, p *Player) (*gameRoom, error) {
	r.roomMu.RLock()
	defer r.roomMu.RUnlock()
	room, ok := r.gameRooms[roomID]
	if !ok {
		return nil, errors.New("room not found")
	}
	room.mu.Lock()
	if room.gameState != Waiting {
		room.mu.Unlock()
		// TODO: defering the feature to allow people to join even while the game is in progress, for now we will just check if the game is in waiting state
		return nil, errors.New("room not found")
	}
	if len(room.players) == 0 {
		room.gameOwner = p.playerID
	}
	room.nextSeq++
	p.joinSeq = room.nextSeq
	room.players[p.playerID] = p
	room.mu.Unlock()
	return room, nil
}

func (r *Registry) CreateRoom() string {
	var roomID string
	for {
		roomID = helpers.RandomStr()
		r.roomMu.Lock()
		if _, ok := r.gameRooms[roomID]; ok {
			r.roomMu.Unlock()
			continue
		}
		r.gameRooms[roomID] = NewGameRoom(roomID, r.ctx)
		r.roomMu.Unlock()
		time.AfterFunc(time.Second*30, func() {
			r.roomMu.RLock()
			room, ok := r.gameRooms[roomID]
			r.roomMu.RUnlock()
			if !ok {
				return
			}
			room.mu.RLock()
			roomEmpty := len(room.players) == 0
			room.mu.RUnlock()
			if !roomEmpty {
				return
			}

			r.roomMu.Lock()
			// doing a ptr comparison to ensure that the room hasn't been replaced with a new one with the same ID
			if ptr, ok := r.gameRooms[roomID]; ok && ptr == room {
				room.mu.Lock()
				if len(room.players) == 0 {
					delete(r.gameRooms, roomID)
				}
				room.mu.Unlock()
			}
			r.roomMu.Unlock()
		})
		return roomID
	}
}

func (r *Registry) Release(p *Player) {
	r.roomMu.Lock()

	p.gameRoom.mu.Lock()
	delete(p.gameRoom.players, p.playerID)
	delete(p.gameRoom.guessedPlayers, p.playerID)
	isEmpty := len(p.gameRoom.players) == 0
	p.gameRoom.mu.Unlock()

	if isEmpty {
		p.gameRoom.cancel()
		delete(r.gameRooms, p.gameRoom.roomID)
		r.roomMu.Unlock()
		return
	}
	r.roomMu.Unlock()

	p.gameRoom.mu.Lock()
	defer p.gameRoom.mu.Unlock()
	if p.playerID == p.gameRoom.gameOwner {
		// change gameRoom owner to the player with the lowest joinSeq
		var newOwner *Player
		for _, player := range p.gameRoom.players {
			if newOwner == nil || newOwner.joinSeq > player.joinSeq {
				newOwner = player
			}
		}
		p.gameRoom.gameOwner = newOwner.playerID
		data, err := protocol.Encode(protocol.TypeOwnerChanged, protocol.WriteOwnerChanged{
			OwnerID: newOwner.playerID,
		})
		if err != nil {
			return
		}
		p.gameRoom.players.broadcast(data)
	}
	if p.gameRoom.gameState != Waiting {
		select {
		case p.gameRoom.signalPlayerChange <- struct{}{}:
		default:
		}
	}
}
