package presence

import (
	"fmt"
	"strings"

	"github.com/Lunidra/Ether-Backend/internal/protocol"
	"github.com/Lunidra/Ether-Backend/internal/session"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {
	return &Handler{
		manager: manager,
	}
}

type PresencePayload struct {
	Server      string `json:"server"`
	ServerIP    string `json:"serverIP"`
	GameVersion string `json:"gameVersion"`
}

func (h *Handler) Request(
	client *session.Client,
	message protocol.Message,
) error {
	if !client.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	return h.manager.Sync(client)
}

func (h *Handler) Update(
	client *session.Client,
	message protocol.Message,
) error {
	if !client.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	var payload PresencePayload

	if err := message.Field(
		"server",
		&payload.Server,
	); err != nil {
		return err
	}

	if err := message.Field(
		"serverIP",
		&payload.ServerIP,
	); err != nil {
		return err
	}

	if err := message.Field(
		"gameVersion",
		&payload.GameVersion,
	); err != nil {
		return err
	}

	payload.Server = strings.TrimSpace(payload.Server)
	payload.ServerIP = strings.TrimSpace(payload.ServerIP)
	payload.GameVersion = strings.TrimSpace(payload.GameVersion)

	if len(payload.Server) > 128 {
		return fmt.Errorf("server name too long")
	}

	if len(payload.ServerIP) > 255 {
		return fmt.Errorf("server IP too long")
	}

	if len(payload.GameVersion) > 32 {
		return fmt.Errorf("game version too long")
	}

	if payload.Server == "" {
		payload.Server = "Unknown"
	}

	if payload.ServerIP == "" {
		payload.ServerIP = "Unknown"
	}

	if payload.GameVersion == "" {
		payload.GameVersion = "Unknown"
	}

	client.SetPresence(session.Presence{
		Server:      payload.Server,
		ServerIP:    payload.ServerIP,
		GameVersion: payload.GameVersion,
	})

	return h.manager.Update(client)
}
