package ws

import (
	"github.com/coder/websocket"
)

type Player struct {
	conn     *websocket.Conn
	username string
	send     chan []byte
}

type playersT map[string]*Player
