package auth

import (
	"fmt"
	"log"
	"strings"

	"github.com/Lunidra/Ether-Backend/internal/protocol"
	"github.com/Lunidra/Ether-Backend/internal/session"
)

type IdentifyPayload struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type VerifyPayload struct {
	ServerID string `json:"serverId"`
}

type Handler struct {
	service         *Service
	onAuthenticated func(*session.Client)
}

func NewHandler(
	service *Service,
	onAuthenticated func(*session.Client),
) *Handler {
	return &Handler{
		service:         service,
		onAuthenticated: onAuthenticated,
	}
}

func (h *Handler) Identify(
	client *session.Client,
	message protocol.Message,
) error {

	var payload IdentifyPayload

	if err := message.Field(
		"uuid",
		&payload.UUID,
	); err != nil {
		return err
	}

	if err := message.Field(
		"username",
		&payload.Username,
	); err != nil {
		return err
	}

	payload.UUID = strings.TrimSpace(
		payload.UUID,
	)

	payload.Username = strings.TrimSpace(
		payload.Username,
	)

	if payload.UUID == "" {
		return fmt.Errorf(
			"missing UUID",
		)
	}

	if payload.Username == "" {
		return fmt.Errorf(
			"missing username",
		)
	}

	client.UUID = payload.UUID
	client.Username = payload.Username

	challenge, err :=
		h.service.Challenges().Create(
			client.ID,
		)

	if err != nil {
		return fmt.Errorf(
			"failed to create authentication challenge: %w",
			err,
		)
	}

	data, err := protocol.LegacyMessage(
		"auth_challenge",
		map[string]any{
			"serverId": challenge.ServerID,
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to encode auth challenge: %w",
			err,
		)
	}

	if err := client.Send(data); err != nil {
		return fmt.Errorf(
			"failed to send auth challenge: %w",
			err,
		)
	}

	log.Printf(
		"authentication challenge issued to %s (%s)",
		client.Username,
		client.ID,
	)

	return nil
}

func (h *Handler) Verify(
	client *session.Client,
	message protocol.Message,
) error {

	var payload VerifyPayload

	if err := message.Field(
		"serverId",
		&payload.ServerID,
	); err != nil {
		return err
	}

	payload.ServerID = strings.TrimSpace(
		payload.ServerID,
	)

	if payload.ServerID == "" {
		return fmt.Errorf(
			"missing serverId",
		)
	}

	challenge, exists :=
		h.service.Challenges().Get(
			payload.ServerID,
		)

	if !exists {
		return fmt.Errorf(
			"invalid or expired authentication challenge",
		)
	}

	if challenge.ClientID != client.ID {
		return fmt.Errorf(
			"authentication challenge belongs to another client",
		)
	}

	if client.Username == "" {
		return fmt.Errorf(
			"client has not completed authentication identification",
		)
	}

	profile, verified, err :=
		h.service.Mojang().HasJoined(
			client.Username,
			payload.ServerID,
		)

	if err != nil {
		return fmt.Errorf(
			"Mojang verification failed: %w",
			err,
		)
	}

	if !verified {
		return fmt.Errorf(
			"Minecraft account ownership could not be verified",
		)
	}

	if !strings.EqualFold(
		profile.Name,
		client.Username,
	) {
		return fmt.Errorf(
			"Mojang username does not match identified username",
		)
	}

	if !h.service.Challenges().Consume(
		payload.ServerID,
	) {
		return fmt.Errorf(
			"authentication challenge could not be consumed",
		)
	}

	client.Authenticated = true
	if h.onAuthenticated != nil {
		h.onAuthenticated(client)
	}

	log.Printf(
		"authenticated %s (%s)",
		client.Username,
		client.ID,
	)

	// Temporary token.
	// Session-token generation will become its own component
	// before this is considered production authentication.
	token := client.ID

	data, err := protocol.LegacyMessage(
		"auth_success",
		map[string]any{
			"token": token,
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to encode auth success: %w",
			err,
		)
	}

	if err := client.Send(data); err != nil {
		return fmt.Errorf(
			"failed to send auth success: %w",
			err,
		)
	}

	return nil
}

// Helper retained for tests that want to inspect the raw payload.
