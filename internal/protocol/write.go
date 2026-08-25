package protocol

import "github.com/google/uuid"

type PlayerEvent struct {
	Username string    `json:"username"`
	PlayerID uuid.UUID `json:"player_id"`
}

type WriteChatPayload struct {
	Text     string    `json:"text"`
	Username string    `json:"username"`
	PlayerID uuid.UUID `json:"player_id"`
}
