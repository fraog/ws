package ws

import (
	"bufio"
	"net"
	"net/http"
	"time"
)

/*
A WebSocket server.
*/
type Server struct {
	net.Listener
	Clients map[int64]*Conn
	Handler Handler
	OnOpen  func(*Conn)
	OnClose func(*Conn)
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

	wsc := Conn{c, false, time.Now().Unix(), s.Handler, nil, nil}
	s.Clients[wsc.Id()] = &wsc
	return &wsc, nil
}

/*
Closes all connections with the server.
*/
func (s *Server) Close() {
	for _, client := range s.Clients {
		client.Close()
	}
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

		c.Handler = s.Handler
		c.OnClose = s.OnClose
		c.server = s
		if s.OnOpen != nil {
			s.OnOpen(c)
		}
		
		//Handle client concurrently
		go c.Handle(handler)
	}
	return nil
}

/*
Sends a message to all clients.
*/
func (s *Server) Write(message []byte) (int, error) {
	for _, client := range s.Clients {
		go client.Write(message)
	}
	return 0, nil
}

func (s *Server) WriteString(message string) (int, error) {
	asBytes := []byte(message)
	return s.Write(asBytes)
}
