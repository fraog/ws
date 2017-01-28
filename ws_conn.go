package ws

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

/*
A WebSocket connection.
*/
type Conn struct {
	nc      net.Conn
	client  bool
	id      int64
	Handler Handler
	server  *Server
	OnClose func(*Conn)
}

/*
Get the unique ID associated with this connection.
*/
func (c *Conn) Id() int64 {
	return c.id
}

/*
Returns true if this connection is a client connected to a server.
*/
func (c *Conn) IsClient() bool {
	return c.client
}

/*
Returns the base net.Conn interface belonging to this connection.
*/
func (c *Conn) Base() net.Conn {
	return c.nc
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
			//TODO: This closes WHILE sending a response Close frame
			fmt.Println("Received close opcode.")
			c.Close()
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
		//FIXME: SHould i be writing all at once? Sometimes th client reads just the frame
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
Writes a framed message to the connection as a string.
*/
func (c *Conn) WriteString(msg string) (int, error) {
	return c.Write([]byte(msg))
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
			fmt.Printf("Error reading from connection: %s\n", e)
			return c.Close()
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

	fmt.Println("Connection closed.")

	//Close callback
	if c.OnClose != nil {
		c.OnClose(c)
	}

	//Send a close frame
	df := NewFrame(nil)
	df.op = opClose
	df.WriteTo(c.nc)

	//If its a server connection, remove from clients
	if c.server != nil {
		delete(c.server.Clients, c.id)
		fmt.Printf("Client close, number of clients is now %v.\n", len(c.server.Clients))
	}

	return c.nc.Close()
}
