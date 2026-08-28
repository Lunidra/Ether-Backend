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

// Broadcast sends a server-originated Ether broadcast to every
// authenticated client.
func (s *Service) Broadcast(sender, message string) error {
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

	return s.BroadcastRaw(data)
}

// BroadcastRaw sends an already encoded packet to all authenticated
// clients.
//
// This is intentionally kept separate from Broadcast because chat,
// presence, and future services may need to broadcast their own
// protocol payloads.
func (s *Service) BroadcastRaw(data []byte) error {
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

func (s *Service) BroadcastExcept(
	except *session.Client,
	data []byte,
) error {
	var firstErr error

	for _, client := range s.clients.AuthenticatedClients() {
		if client.ID == except.ID {
			continue
		}

		if err := client.Send(data); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
