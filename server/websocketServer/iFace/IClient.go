package iFace

import (
	"net"
)

type IClient interface {
	ID() string
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	Text(messageType int, message []byte) error
	Binary(messageType int, message []byte) error
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	WLoop()
	RLoop()
	Close()
}
