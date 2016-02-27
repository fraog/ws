package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type JSONMessageHandler map[string]func(*Conn, interface{}) bool

func NewJSONMessageHandler() JSONMessageHandler {
	return make(map[string]func(*Conn, interface{}) bool)
}

func (mh *JSONMessageHandler) AddHandler(key string, handler func(*Conn, interface{}) bool) {
	(*mh)[key] = handler
}

func (mh *JSONMessageHandler) RemoveHandler(key string) {
	delete(*mh, key)
}

func (mh *JSONMessageHandler) Handle(c *Conn, msg []byte) bool {

	buffer := bytes.NewBuffer(msg)

	key, e := buffer.ReadString(':')
	if e != nil {
		return false
	}
	key = key[:len(key)-1]

	if handler, k := (*mh)[key]; k {

		var f interface{}
		e = json.Unmarshal(buffer.Bytes(), &f)
		if e != nil {
			fmt.Printf("Error while unmarshalling JSON: %s\n", e)
			return false
		}

		return handler(c, f)
	}

	fmt.Printf("Couldn't find handler for %s.\n", key)
	return false
}
