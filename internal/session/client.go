package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
	gws "github.com/gorilla/websocket"
)

type Client struct {
	ID string

	Conn *gws.Conn

	ConnectedAt time.Time

	mu sync.Mutex

	Authenticated bool

	UUID     string
	Username string
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