// Package protocol is responsible for creating the protocol for ws
package protocol

import (
	"encoding/json"
	"errors"
)

var ErrInvalidProtocol error = errors.New("invalid protocol error")

type MessageType string

const (
	TypeChat     MessageType = "chat"
	TypeJoinChat MessageType = "join"
)

type BaseStructure struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Message struct {
	Text string `json:"text"`
}
