package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
	gws "github.com/gorilla/websocket"
)

type Presence struct {
	Server      string
	ServerIP    string
	GameVersion string
}

type Client struct {
	ID string

	Conn *gws.Conn

	ConnectedAt time.Time

	mu sync.Mutex

	Authenticated bool
	Identified    bool

	UUID     string
	Username string

	Presence Presence
}

func NewClient(conn *gws.Conn) *Client {
	return &Client{
		ID:          uuid.NewString(),
		Conn:        conn,
		ConnectedAt: time.Now(),
	}
}

func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Conn.WriteMessage(
		gws.TextMessage,
		data,
	)
}

func (c *Client) SetPresence(presence Presence) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Presence = presence
}

func (c *Client) GetPresence() Presence {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Presence
}

func (c *Client) Identity() (uuid string, username string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.UUID, c.Username
}
