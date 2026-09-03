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

	AuthAttempts      int
	AuthAttemptWindow time.Time

	PacketCount  int
	PacketWindow time.Time

	ViolationCount  int
	ViolationWindow time.Time

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

func (c *Client) AllowPacket(
	maxPackets int,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	if c.PacketWindow.IsZero() ||
		now.Sub(c.PacketWindow) >= time.Second {

		c.PacketWindow = now
		c.PacketCount = 0
	}

	if c.PacketCount >= maxPackets {
		return false
	}

	c.PacketCount++

	return true
}

func (c *Client) AllowAuthAttempt(maxAttempts int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	if c.AuthAttemptWindow.IsZero() ||
		now.Sub(c.AuthAttemptWindow) >= time.Minute {
		c.AuthAttemptWindow = now
		c.AuthAttempts = 0
	}

	if c.AuthAttempts >= maxAttempts {
		return false
	}

	c.AuthAttempts++

	return true
}

func (c *Client) AllowViolationLog(
	maxViolations int,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	if c.ViolationWindow.IsZero() ||
		now.Sub(c.ViolationWindow) >= time.Second {

		c.ViolationWindow = now
		c.ViolationCount = 0
	}

	if c.ViolationCount >= maxViolations {
		return false
	}

	c.ViolationCount++

	return true
}

func (c *Client) IsAuthenticated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Authenticated
}

func (c *Client) IsIdentified() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Identified
}

func (c *Client) SetIdentified() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Identified = true
}

func (c *Client) SetAuthenticated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Authenticated {
		return false
	}

	c.Authenticated = true

	return true
}
