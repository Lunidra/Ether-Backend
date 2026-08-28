package session

import (
	"strings"
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewManager() *Manager {
	return &Manager{
		clients: make(
			map[string]*Client,
		),
	}
}

func (m *Manager) Add(
	client *Client,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client.ID] = client
}

func (m *Manager) Remove(
	id string,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(
		m.clients,
		id,
	)
}

func (m *Manager) Get(
	id string,
) (*Client, bool) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists :=
		m.clients[id]

	return client, exists
}

func (m *Manager) Count() int {

	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clients)
}

func (m *Manager) AuthenticatedClients() []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]*Client, 0)

	for _, client := range m.clients {
		if client.Authenticated {
			clients = append(clients, client)
		}
	}

	return clients
}

// Authenticated UUID
func (m *Manager) HasAuthenticatedUUID(uuid string) bool {
	for _, client := range m.AuthenticatedClients() {
		clientUUID, _ := client.Identity()

		if strings.EqualFold(clientUUID, uuid) {
			return true
		}
	}

	return false
}
