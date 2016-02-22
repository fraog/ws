package ws

import (
	"net"
	"io"
	"bytes"
)

/*
A WebSocket connection.
*/
type Conn struct {
	nc     net.Conn
	client bool
	id     int64
	Handler Handler
}

/*
Reads a framed message to the connection and writes it to a Writer.
*/
func (c *Conn) ReadTo(w io.Writer) error {
	
	var f DataFrame
	var e error

	//Read frame from connection
	if e = f.ReadFrom(c.nc); e != nil {
		return e
	}

	//Evaluate opcode
	if f.op > opBinary {
		switch f.op {
		case opClose:
			c.nc.Close()
			return nil
		case opPing:
			return nil
		default:
			return nil
		}
	}

	if !c.client {
		//Read and decode into writer
		if e = f.DecodeTo(c.nc, w); e != nil {
			return e
		}
	} else {
		//Read straight into writer
		if _, e = io.CopyN(w, c.nc, int64(f.length)); e != nil {
			return e
		}

	}
	return nil
}

/*
Writes a framed message to the connection.
*/
func (c *Conn) Write(b []byte) (int, error) {
	frame := NewFrame(b)
	if !c.client {
		frame.WriteTo(c.nc)
		c.nc.Write(b)
	} else {
		frame.GenerateMask()
		frame.WriteTo(c.nc)
		frame.Encode(b, c.nc)
	}
	return 0, nil
}

/*
Listen on connection continuously, routing to messages to handler.
*/
func (c *Conn) Handle(handler func(*Conn, []byte)) error {

	var e error		

	for {
		var buffer bytes.Buffer
		e = c.ReadTo(&buffer)
		if e != nil {
			return e
		}
		
		if c.Handler != nil {
			if c.Handler.Handle(c, buffer.Bytes()) {
				continue
			}
		}
		
		if handler != nil {
			handler(c, buffer.Bytes())
		}
	}
}

/*
Closes a websocket connection.
*/
func (c *Conn) Close() error {
	return c.nc.Close()
}



