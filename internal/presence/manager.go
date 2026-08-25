package presence

import (
	"github.com/Lunidra/Ether-Backend/internal/protocol"
	"github.com/Lunidra/Ether-Backend/internal/session"
)

type Manager struct {
	clients *session.Manager
}

func NewManager(clients *session.Manager) *Manager {
	return &Manager{
		clients: clients,
	}
}

type User struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	Server      string `json:"server"`
	GameVersion string `json:"gameVersion,omitempty"`
}

func userFromClient(client *session.Client) User {
	uuid, username := client.Identity()
	p := client.GetPresence()

	return User{
		UUID:        uuid,
		Username:    username,
		Server:      p.Server,
		GameVersion: p.GameVersion,
	}
}

func (m *Manager) Sync(client *session.Client) error {
	users := make([]User, 0)

	for _, other := range m.clients.AuthenticatedClients() {
		users = append(users, userFromClient(other))
	}

	data, err := protocol.LegacyMessage(
		"presence_sync",
		map[string]any{
			"users": users,
		},
	)
	if err != nil {
		return err
	}

	if err := client.Send(data); err != nil {
		return err
	}

	// Send authoritative presence updates as well.
	//
	// This is important because the legacy client creates users from
	// user_join with an "Unknown" server. The update packets ensure
	// that users already online receive their current presence.
	for _, other := range m.clients.AuthenticatedClients() {
		uuid, _ := other.Identity()

		// Don't need to send an update for the requesting client here.
		// Its own presence is already represented in presence_sync.
		if other.ID == client.ID {
			continue
		}

		presence := other.GetPresence()

		update, err := protocol.LegacyMessage(
			"user_update",
			map[string]any{
				"uuid":        uuid,
				"server":      presence.Server,
				"gameVersion": presence.GameVersion,
			},
		)
		if err != nil {
			return err
		}

		if err := client.Send(update); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Join(client *session.Client) error {
	uuid, username := client.Identity()
	p := client.GetPresence()

	data, err := protocol.LegacyMessage(
		"user_join",
		map[string]any{
			"uuid":        uuid,
			"username":    username,
			"server":      p.Server,
			"gameVersion": p.GameVersion,
		},
	)
	if err != nil {
		return err
	}

	return m.broadcastExcept(client, data)
}

func (m *Manager) Leave(client *session.Client) error {
	uuid, _ := client.Identity()

	data, err := protocol.LegacyMessage(
		"user_leave",
		map[string]any{
			"uuid": uuid,
		},
	)
	if err != nil {
		return err
	}

	return m.broadcastExcept(client, data)
}

func (m *Manager) Update(client *session.Client) error {
	uuid, username := client.Identity()
	p := client.GetPresence()

	data, err := protocol.LegacyMessage(
		"user_update",
		map[string]any{
			"uuid":        uuid,
			"username":    username,
			"server":      p.Server,
			"gameVersion": p.GameVersion,
		},
	)
	if err != nil {
		return err
	}

	return m.broadcast(data)
}

func (m *Manager) broadcast(data []byte) error {
	var firstErr error

	for _, client := range m.clients.AuthenticatedClients() {
		if err := client.Send(data); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (m *Manager) broadcastExcept(
	except *session.Client,
	data []byte,
) error {
	var firstErr error

	for _, client := range m.clients.AuthenticatedClients() {
		if client.ID == except.ID {
			continue
		}

		if err := client.Send(data); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
