package engine

import (
	"github.com/liu-willow/leopard/server/websocketServer/iFace"
	"sync"
)

type hub struct {
	clients    iFace.IClients
	broadcast  chan *iFace.Envelope
	register   chan iFace.IClient
	unregister chan iFace.IClient
	exit       chan *iFace.Envelope
	open       bool
	mutex      *sync.RWMutex
}

var _ iFace.IHub = (*hub)(nil)

func newHub(clients iFace.IClients) *hub {
	return &hub{
		clients:    clients,
		broadcast:  make(chan *iFace.Envelope),
		register:   make(chan iFace.IClient),
		unregister: make(chan iFace.IClient),
		exit:       make(chan *iFace.Envelope),
		open:       true,
		mutex:      &sync.RWMutex{},
	}
}

func (h *hub) Run() {
loop:
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients.Add(client)
			h.mutex.Unlock()
		case s := <-h.unregister:
			h.mutex.Lock()
			h.clients.Delete(s.ID())
			h.mutex.Unlock()
		case m := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients.All() {
				if m.Filter != nil {
					if m.Filter(client) {
						client.Write(m.MessageType, m.Message)
					}
				} else {
					client.Write(m.MessageType, m.Message)
				}
			}
			h.mutex.RUnlock()
		case m := <-h.exit:
			h.mutex.Lock()
			for ID, client := range h.clients.All() {
				client.Write(m.MessageType, m.Message)
				h.clients.Delete(ID)
				client.Close()
			}
			h.open = false
			h.mutex.Unlock()
			break loop
		}
	}
}

func (h *hub) closed() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return !h.open
}

func (h *hub) Count() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.clients.All())
}

func (h *hub) All() map[string]iFace.IClient {
	return h.clients.All()
}
