package iFace

import (
	"net"
)

type IClient interface {
	ID() string
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	Write(messageType int, message []byte) error
	WriteRaw(messageType int, message []byte) error
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	WLoop()
	RLoop()
	Close()
}
