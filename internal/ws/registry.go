// Package ws is for websockets and players
package ws

import (
	"context"
	"log/slog"
	"sync"
)

type Registry struct {
	players  playersT
	claimed  map[string]struct{}
	playerMu sync.RWMutex
	logger   *slog.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewRegistry(logger *slog.Logger) *Registry {
	r := &Registry{
		players:  make(playersT),
		claimed:  make(map[string]struct{}),
		playerMu: sync.RWMutex{},
		logger:   logger,
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	return r
}

func (r *Registry) TryClaim(username string) bool {
	r.playerMu.Lock()
	defer r.playerMu.Unlock()
	if _, ok := r.claimed[username]; ok {
		return false
	}
	r.claimed[username] = struct{}{}
	return true
}

func (r *Registry) Release(username string) {
	r.playerMu.Lock()
	defer r.playerMu.Unlock()
	delete(r.claimed, username)
	delete(r.players, username)
}

func (r *Registry) Attach(username string, p *Player) {
	r.playerMu.Lock()
	defer r.playerMu.Unlock()
	r.players[username] = p
}
