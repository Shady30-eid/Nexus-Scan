package ipc

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
)

type Event struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

type Command struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func ErrorEvent(code, message string) Event {
	return Event{
		Type:      "error",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

type Client struct {
	conn   *websocket.Conn
	send   chan Event
	log    *logger.Logger
	once   sync.Once
	closed chan struct{}
}

func NewClient(conn *websocket.Conn, log *logger.Logger) *Client {
	return &Client{
		conn:   conn,
		send:   make(chan Event, 256),
		log:    log,
		closed: make(chan struct{}),
	}
}

func (c *Client) Send(event Event) {
	select {
	case c.send <- event:
	case <-c.closed:
	default:
		c.log.Warn("client send buffer full, dropping event", "type", event.Type)
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.once.Do(func() { close(c.closed) })
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				c.log.Error("marshal event failed", "error", err, "type", event.Type)
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-c.closed:
			return
		}
	}
}

func (c *Client) Close() {
	c.once.Do(func() { close(c.closed) })
}

type Broadcaster struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	log     *logger.Logger
}

func NewBroadcaster(log *logger.Logger) *Broadcaster {
	return &Broadcaster{
		clients: make(map[*Client]struct{}),
		log:     log,
	}
}

func (b *Broadcaster) Register(c *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c] = struct{}{}
	b.log.Info("broadcaster: client registered", "total_clients", len(b.clients))
}

func (b *Broadcaster) Unregister(c *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, c)
	c.Close()
	b.log.Info("broadcaster: client unregistered", "total_clients", len(b.clients))
}

func (b *Broadcaster) Broadcast(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for c := range b.clients {
		c.Send(event)
	}
}

func (b *Broadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
