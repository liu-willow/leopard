package iFace

import "time"

type Config struct {
	WriteWait         time.Duration
	PongWait          time.Duration
	PingPeriod        time.Duration
	MaxMessageSize    int64
	MessageBufferSize int
	ReadBufferSize    int
	WriteBufferSize   int
}
