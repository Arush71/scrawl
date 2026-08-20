package ws

import (
	"fmt"

	"github.com/Arush71/scrawl/internal/protocol"
)

func (r *Registry) handleMessage(message protocol.ChatPayload, pl *Player) error {
	msg, err := protocol.Encode(protocol.TypeChat, protocol.WriteChatPayload{
		Username: pl.username,
		Text:     message.Text,
	})
	if err != nil {
		return fmt.Errorf("marshal write message: %w", err)
	}
	r.playerMu.RLock()
	defer r.playerMu.RUnlock()
	r.players.broadcast(msg)
	return nil
}
