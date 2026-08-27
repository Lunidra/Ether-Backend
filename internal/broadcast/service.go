package broadcast

import (
	"fmt"
	"strings"

	"github.com/Lunidra/Ether-Backend/internal/protocol"
	"github.com/Lunidra/Ether-Backend/internal/session"
)

const (
	MaxSenderLength  = 64
	MaxMessageLength = 2048
)

type Service struct {
	clients *session.Manager
}

func NewService(clients *session.Manager) *Service {
	return &Service{
		clients: clients,
	}
}

// SendRaw sends an already-encoded packet to every authenticated client.
//
// This is useful for packets whose wire format is not the normal
// protocol.Message format, such as legacy EtherChat ChatMessage packets.
func (s *Service) SendRaw(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot broadcast empty packet")
	}

	var firstErr error

	for _, client := range s.clients.AuthenticatedClients() {
		if err := client.Send(data); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Send sends the legacy Ether broadcast packet.
//
// This is intentionally server-originated. Clients must never be allowed
// to invoke this directly, otherwise any client could impersonate Ether.
func (s *Service) Send(sender string, message string) error {
	sender = strings.TrimSpace(sender)
	message = strings.TrimSpace(message)

	if sender == "" {
		return fmt.Errorf("broadcast sender cannot be empty")
	}

	if len([]rune(sender)) > MaxSenderLength {
		return fmt.Errorf(
			"broadcast sender exceeds %d characters",
			MaxSenderLength,
		)
	}

	if message == "" {
		return fmt.Errorf("broadcast message cannot be empty")
	}

	if len([]rune(message)) > MaxMessageLength {
		return fmt.Errorf(
			"broadcast message exceeds %d characters",
			MaxMessageLength,
		)
	}

	data, err := protocol.LegacyMessage(
		"broadcast",
		map[string]any{
			"sender":  sender,
			"message": message,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"failed to encode broadcast: %w",
			err,
		)
	}

	return s.SendRaw(data)
}
