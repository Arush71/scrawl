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

type WriteRotationEnd struct {
	Word string `json:"word"`
	// PlayersPoints map[uuid.UUID]int `json:"players_points"` // omitting for now
	AllGuessed bool `json:"all_guessed"` // true if all guessed the word, false if time ran out
}

type WriteRoundComplete struct {
	RoundCompleted int `json:"round_completed"`
}

type WriteGameOver struct {
	// Winners map[uuid.UUID]int `json:"winners"`  // list of 3 players who won
}
