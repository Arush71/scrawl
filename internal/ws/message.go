package ws

import (
	"fmt"
	"strings"

	"github.com/Arush71/scrawl/internal/protocol"
)

func handleMessage(message protocol.ChatPayload, pl *Player) error {
	var data []byte
	var err error
	pl.gameRoom.mu.RLock()
	defer pl.gameRoom.mu.RUnlock()

	if pl.gameRoom.gameState == Drawing {
		if _, ok := pl.gameRoom.guessedPlayers[pl.playerID]; ok {
			return nil
		}
		if pl.playerID == pl.gameRoom.playerOrder[pl.gameRoom.currentDrawerIdx] {
			// TODO: Currently, not allowing drawer and the people that have guessed the word chat
			// Later, there would be a seperate chat for them.
			return nil
		}
		if strings.EqualFold(message.Text, pl.gameRoom.currentWord) {
			pl.gameRoom.guessedPlayers[pl.playerID] = struct{}{}
			if len(pl.gameRoom.guessedPlayers) == len(pl.gameRoom.players)-1 {
				pl.gameRoom.guessListner <- struct{}{}
			}
			data, err = protocol.Encode(protocol.TypeWordGuessed, protocol.WriteWordGuessed{
				Username: pl.username,
				UserID:   pl.playerID,
			})
			if err != nil {
				return fmt.Errorf("marshal write message: %w", err)
			}
		} else {
			data, err = protocol.Encode(protocol.TypeChat, protocol.WriteChat{
				Username: pl.username,
				PlayerID: pl.playerID,
				Text:     message.Text,
			})
			if err != nil {
				return fmt.Errorf("marshal write message: %w", err)
			}
		}
	} else {
		data, err = protocol.Encode(protocol.TypeChat, protocol.WriteChat{
			Username: pl.username,
			PlayerID: pl.playerID,
			Text:     message.Text,
		})
		if err != nil {
			return fmt.Errorf("marshal write message: %w", err)
		}
	}
	pl.gameRoom.players.broadcast(data)
	return nil
}
