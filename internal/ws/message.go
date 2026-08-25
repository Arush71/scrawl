package ws

import (
	"fmt"

	"github.com/Arush71/scrawl/internal/protocol"
)

func (r *Registry) handleMessage(message protocol.ChatPayload, pl *Player) error {
	msg, err := protocol.Encode(protocol.TypeChat, protocol.WriteChatPayload{
		Username: pl.username,
		PlayerID: pl.playerID,
		Text:     message.Text,
	})
	if err != nil {
		return fmt.Errorf("marshal write message: %w", err)
	}
	r.roomMu.RLock()
	defer r.roomMu.RUnlock()
	pl.gameRoom.players.broadcast(msg)
	return nil
}
