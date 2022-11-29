package iFace

type IClients interface {
	All() map[string]IClient
	Count() int
	Add(client IClient) bool
	Get(ID string) IClient
	Delete(ID string) bool
}
