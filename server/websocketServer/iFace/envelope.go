package iFace

type Envelope struct {
	MessageType int
	Message     []byte
	Filter      func(IClient) bool
}
