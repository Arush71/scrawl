package ws

import (
	"fmt"
	"strings"

	"github.com/Arush71/scrawl/internal/protocol"
)

func handleMessage(message protocol.ChatPayload, pl *Player) error {
	var data []byte
	var err error
	pl.gameRoom.mu.Lock()
	switch {
	case pl.gameRoom.gameState != Drawing:
		pl.gameRoom.mu.Unlock()
		data, err = protocol.Encode(protocol.TypeChat, protocol.WriteChat{
			Username: pl.username,
			PlayerID: pl.playerID,
			Text:     message.Text,
		})
	default:
		if _, ok := pl.gameRoom.guessedPlayers[pl.playerID]; ok {
			pl.gameRoom.mu.Unlock()
			return nil
		}
		if pl.playerID == pl.gameRoom.playerOrder[pl.gameRoom.currentDrawerIdx] {
			pl.gameRoom.mu.Unlock()
			// TODO: Currently, not allowing drawer and the people that have guessed the word chat
			// Later, there would be a seperate chat for them.
			return nil
		}

		if strings.EqualFold(message.Text, pl.gameRoom.currentWord) {
			pl.gameRoom.guessedPlayers[pl.playerID] = struct{}{}
			if len(pl.gameRoom.guessedPlayers) == len(pl.gameRoom.players)-1 {
				select {
				case pl.gameRoom.guessListner <- struct{}{}:
				default:
				}
			}
			pl.gameRoom.mu.Unlock()
			data, err = protocol.Encode(protocol.TypeWordGuessed, protocol.WriteWordGuessed{
				Username: pl.username,
				UserID:   pl.playerID,
			})
		} else {
			pl.gameRoom.mu.Unlock()
			data, err = protocol.Encode(protocol.TypeChat, protocol.WriteChat{
				Username: pl.username,
				PlayerID: pl.playerID,
				Text:     message.Text,
			})
		}
	}
	if err != nil {
		return fmt.Errorf("marshal write message: %w", err)
	}
	pl.gameRoom.mu.RLock()
	pl.gameRoom.players.broadcast(data)
	pl.gameRoom.mu.RUnlock()
	return nil
}
