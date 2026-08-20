// Package protocol is responsible for creating the protocol
package protocol

import (
	"encoding/json"
	"errors"
)

var ErrInvalidProtocol error = errors.New("invalid protocol error")

type MessageType string

const (
	// Client-> Server
	TypeChat MessageType = "chat"

	// Server-> Client
	TypePlayerJoined MessageType = "player_joined"
	TypePlayerLeft   MessageType = "player_left"
)

type Envelope struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

func Encode(t MessageType, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{
		Type: t,
		Data: data,
	})
}

func Decode(data []byte) (Envelope, error) {
	var envelope Envelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Envelope{}, err
	}
	return envelope, nil
}
