package ws

import "sync"

type gameRoom struct {
	mu      sync.RWMutex
	roomID  string
	players playersT
}
