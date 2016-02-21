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
	//Id int64
}

/*
Reads a framed message to the connection and writes it to a Writer.
*/
func (c *Conn) ReadTo(w io.Writer) error {
	
	//FIXME: I would rather not create the data frame here.
	var f DataFrame
	var e error

	//Read frame from connection
	if e = f.ReadFrom(c.nc); e != nil {
		//TODO: Handle error
		return e
	}

	//Evaluate opcode
	if f.op > opBinary {
		switch f.op {
		case opClose:
			c.nc.Close()
			return nil
		case opPing:
			//TODO: Return the payload as is (maxlen 125)
			//continue
			return nil
		default:
			return nil
		}
	}

	//Decode payload data from connection
	/*
	buffer = make([]byte, f.length)
	if e = f.Decode(c.nc, buffer); e != nil {
		//TODO: Handle error
		return e
	}
	*/
	
	if !c.client {
		//Read + Decode into writer
		if e = f.DecodeTo(c.nc, w); e != nil {
			return e
		}
	} else {
		//Read straight into writer
		//dest, src, n
		if _, e = io.CopyN(w, c.nc, int64(f.length)); e != nil {
			return e
		}

	}

	//Pass result to handler
	//handler(c, buffer)
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

//FIXME: What about an error handler as well?
func (c *Conn) Handle(handler func(*Conn, []byte)) error {

	var e error		
	
	//FIXME: It would be nice to keep a dataframe around to be
	//manipulated instead of creating a new one.
	//I could store c.frame
	//var f DataFrame

	for {
		var buffer bytes.Buffer
		e = c.ReadTo(&buffer)
		if e != nil {
			return e
		}
		
		//Should I create a copy of bytes or does it not matter since i wont use this buffer again?
		handler(c, buffer.Bytes())
	}
}

/*
Closes a websocket connection.
FIXME: All connections should send a close message on Close()
*/
func (c *Conn) Close() error {
	//TODO: Send close signal to either clients or server
	return c.nc.Close()
}



