package ws

import (
	"bufio"
	"net"
	"net/http"
)



/*
A WebSocket server.
*/
type Server struct {
	net.Listener
}

/*
Accepts a WebSocket connection, replying to the client with an accept response and returns it.
An error will be returned if the request was invalid or the connection was not accepted.
*/
func (s *Server) Accept() (*Conn, error) {
	//Accept on the underlying Listener
	c, e := s.Listener.Accept()
	if e != nil {
		return nil, e
	}

	//Get Request
	req, e := http.ReadRequest(bufio.NewReader(c))
	if e != nil {
		return nil, e
	}

	//Send Response
	e = createAcceptResponse(req).Write(c)
	if e != nil {
		return nil, e
	}

	wsc := Conn{c, false}
	return &wsc, nil
}

/*
Tells the server to start accepting connections.
*/
func (s *Server) Serve(handler func(*Conn, []byte)) error {
	for {
		var c *Conn
		var e error

		//Accept
		if c, e = s.Accept(); e != nil {
			return e
		}

		//Handle client concurrently
		//TODO: I would like to catch errors from this or route them to a place where I could evaluate them
		go c.Handle(handler)
	}
	return nil
}




