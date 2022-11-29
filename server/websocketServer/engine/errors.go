package engine

import "errors"

var (
	ErrorHubClosed    = errors.New("hub closed")
	ErrorClientClosed = errors.New("client closed")
	ErrorBufferFull   = errors.New("client message buffer is full")
)
