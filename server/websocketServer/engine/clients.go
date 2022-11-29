package engine

import (
	"github.com/liu-willow/leopard/server/websocketServer/iFace"
)

type clients struct {
	members map[string]iFace.IClient
}

var _ iFace.IClients = (*clients)(nil)

var dClients = &clients{members: make(map[string]iFace.IClient)}

func (clients *clients) All() map[string]iFace.IClient {
	return clients.members
}
func (clients *clients) Count() int {
	return len(clients.members)
}
func (clients *clients) Add(client iFace.IClient) bool {
	clients.members[client.ID()] = client
	return true
}
func (clients *clients) Delete(ID string) bool {
	delete(clients.members, ID)
	return true
}
func (clients *clients) Get(ID string) iFace.IClient {
	return clients.members[ID]
}
