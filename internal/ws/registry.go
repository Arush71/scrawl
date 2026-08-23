// Package ws is for websockets and players
package ws

import (
	"context"
	"log/slog"
	"sync"

	"github.com/Arush71/scrawl/internal/helpers"
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

// func (r *Registry) TryClaim(username string) bool {
// 	r.roomMu.Lock()
// 	defer r.roomMu.Unlock()
// 	if _, ok := r.claimed[username]; ok {
// 		return false
// 	}
// 	r.claimed[username] = struct{}{}
// 	return true
// }
//
// func (r *Registry) Release(username string) {
// 	r.roomMu.Lock()
// 	defer r.roomMu.Unlock()
// 	delete(r.claimed, username)
// 	delete(r.players, username)
// }
//
// func (r *Registry) Attach(username string, p *Player) {
// 	r.roomMu.Lock()
// 	defer r.roomMu.Unlock()
// 	r.players[username] = p
// }

func (r *Registry) CreateRoom() string {
	var roomID string
	for {
		roomID = helpers.RandomStr()
		r.roomMu.Lock()
		if _, ok := r.gameRooms[roomID]; ok {
			r.roomMu.Unlock()
			continue
		}
		r.gameRooms[roomID] = &gameRoom{
			mu:      sync.RWMutex{},
			roomID:  roomID,
			players: make(map[string]*Player),
		}
		r.roomMu.Unlock()
		return roomID
	}
}
