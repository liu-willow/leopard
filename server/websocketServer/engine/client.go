package engine

import (
	"github.com/gorilla/websocket"
	"github.com/liu-willow/leopard/server/websocketServer/iFace"
	"net"
	"sync"
	"time"
)

type Client struct {
	context    map[string]interface{}
	uuid       string
	conn       *websocket.Conn
	output     chan *iFace.Envelope
	outputDone chan struct{}
	leopard    *Engine
	open       bool
	mutex      *sync.RWMutex
}

var _ iFace.IClient = (*Client)(nil)

func (c *Client) ID() string {
	return c.uuid
}

func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.context == nil {
		c.context = make(map[string]interface{})
	}

	c.context[key] = value
}

func (c *Client) Get(key string) (value interface{}, exists bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.context != nil {
		value, exists = c.context[key]
	}

	return
}

func (c *Client) closed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return !c.open
}

func (c *Client) WLoop() {
	ticker := time.NewTicker(c.leopard.Config.PingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case message := <-c.output:
			err := c.Binary(message.MessageType, message.Message)

			if err != nil {
				c.leopard.handles.Error(c, err)
				break loop
			}

			if message.MessageType == websocket.CloseMessage {
				break loop
			}

			if message.MessageType == websocket.TextMessage {
				c.leopard.handles.MessageEnd(c, message.Message)
			}

			if message.MessageType == websocket.BinaryMessage {
				c.leopard.handles.MessageEnd(c, message.Message)
			}
		case <-ticker.C:
			c.Binary(websocket.PingMessage, []byte{})
		case _, ok := <-c.outputDone:
			if !ok {
				break loop
			}
		}
	}
}
func (c *Client) RLoop() {
	c.conn.SetReadLimit(c.leopard.Config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.leopard.Config.PongWait))

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.leopard.Config.PongWait))
		c.leopard.handles.Ping(c)
		return nil
	})

	if c.leopard.handles.Close != nil {
		c.conn.SetCloseHandler(func(code int, text string) error {
			return c.leopard.handles.Close(c, code, text)
		})
	}

	for {
		t, message, err := c.conn.ReadMessage()

		if err != nil {
			c.leopard.handles.Error(c, err)
			break
		}

		if t == websocket.TextMessage || t == websocket.BinaryMessage {
			c.leopard.handles.Message(c, message)
		}
	}
}

func (c *Client) Close() {
	c.mutex.Lock()
	open := c.open
	c.open = false
	c.mutex.Unlock()
	if open {
		c.conn.Close()
		close(c.outputDone)
	}

}

func (c *Client) write(message *iFace.Envelope) {
	if c.closed() {
		c.leopard.handles.Error(c, ErrorClientClosed)
		return
	}

	select {
	case c.output <- message:
	default:
		c.leopard.handles.Error(c, ErrorBufferFull)
	}
}

func (c *Client) Binary(messageType int, message []byte) error {
	if c.closed() {
		return ErrorClientClosed
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.leopard.Config.WriteWait))
	err := c.conn.WriteMessage(messageType, message)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Text(messageType int, message []byte) error {
	if c.closed() {
		return websocket.ErrCloseSent
	}

	c.write(&iFace.Envelope{MessageType: messageType, Message: message})

	return nil
}
