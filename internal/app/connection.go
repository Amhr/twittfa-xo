package app

import "net"

type Connection interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	RemoteAddr() net.Addr
	Close() error
}
