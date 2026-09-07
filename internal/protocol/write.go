package protocol

import "github.com/google/uuid"

type PlayerEvent struct {
	Username string    `json:"username"`
	PlayerID uuid.UUID `json:"player_id"`
}

type WriteChat struct {
	Text     string    `json:"text"`
	Username string    `json:"username"`
	PlayerID uuid.UUID `json:"player_id"`
}

type WritePickWord struct{}

type WritePhasePick struct {
	DrawerID uuid.UUID `json:"drawer_id"`
}

type WritePhaseDraw struct {
	DrawerID   uuid.UUID `json:"drawer_id"`
	WordLength int       `json:"word_length"`
	TimeLimit  int       `json:"time_limit"` // seconds
}

type WriteWordGuessed struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}
