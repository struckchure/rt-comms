package ws

import (
	"github.com/gorilla/websocket"
)

const MaxMessageSize = 64 * 1024

type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	Done   chan struct{}
}

func NewClient(userID string, conn *websocket.Conn) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Done:   make(chan struct{}),
	}
}

func (c *Client) Close() {
	select {
	case <-c.Done:
	default:
		close(c.Done)
	}
}
