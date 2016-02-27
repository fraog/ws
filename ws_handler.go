package ws

import (
	"bytes"
)

type Handler interface {
	Handle(*Conn, []byte) bool
}

type MessageHandler map[string]func(*Conn, []byte) bool

func NewMessageHandler() MessageHandler {
	return make(map[string]func(*Conn, []byte) bool)
}

func (mh *MessageHandler) AddHandler(key string, handler func(*Conn, []byte) bool) {
	(*mh)[key] = handler
}

func (mh *MessageHandler) RemoveHandler(key string) {
	delete(*mh, key)
}

func (mh *MessageHandler) Handle(c *Conn, msg []byte) bool {

	buffer := bytes.NewBuffer(msg)

	key, e := buffer.ReadString(':')
	if e != nil {
		return false
	}
	key = key[:len(key)-1]

	if handler, k := (*mh)[key]; k {
		return handler(c, buffer.Bytes())
	}

	return false
}
