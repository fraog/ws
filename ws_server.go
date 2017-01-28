package ws

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

/*
A WebSocket server.
*/
type Server struct {
	net.Listener
	Clients  map[int64]*Conn
	Handler  Handler
	OnAccept func(net.Conn) bool
	OnOpen   func(*Conn)
	OnClose  func(*Conn)
	SLog     *log.Logger
}

/*
Accept accepts a WebSocket connection, replying to the client with an accept response and returns it.
An error will be returned if the request was invalid or the connection was not accepted.
TODO: Check request for validity
*/
func (s *Server) Accept() (*Conn, error) {
	//s.SLog.Println("Accept connection.")

	//Accept on the underlying Listener
	c, e := s.Listener.Accept()
	if e != nil {
		s.SLog.Printf("listener couldnt accept: %s\n", e)
		return nil, e
	}

	if !s.OnAccept(c) {
		return nil, fmt.Errorf("Connection not approved: %s\n", c.RemoteAddr())
	}

	//Get Request
	//s.SLog.Printf("Reading request from %s\n", c.RemoteAddr())

	req, e := http.ReadRequest(bufio.NewReader(c))
	//req, e := http.ReadRequest(bufio.NewReader(reader))

	if e != nil {
		s.SLog.Printf("HTTP WS Request parse error: %s\n", e)
		return nil, e
	}

	e = createAcceptResponse(req).Write(c)
	if e != nil {
		s.SLog.Printf("HTTP WS Response parse error: %s\n", e)
		return nil, e
	}

	//createAcceptResponse(req).Write(os.Stdout)

	wsc := Conn{c, false, time.Now().Unix(), s.Handler, nil, nil}
	s.Clients[wsc.Id()] = &wsc
	s.SLog.Printf("Client connected, number of clients is now %v.\n", len(s.Clients))

	return &wsc, nil
}

/*
Close closes all connections with the server.
*/
func (s *Server) Close() {
	s.SLog.Println("Closing all connectons...")
	for _, client := range s.Clients {
		client.Close()
	}
}

/*
Serve tells the server to start accepting connections.
*/
func (s *Server) Serve(handler func(*Conn, []byte)) error {
	for {
		var c *Conn
		var e error

		//Accept
		if c, e = s.Accept(); e != nil {
			//s.SLog.Println("Error acceptig")
			s.SLog.Printf("Not accepted: %s\n", e)
			//return e
			continue
		}

		//s.SLog.Println("Connection accepted")
		c.Handler = s.Handler
		c.OnClose = s.OnClose
		c.server = s
		if s.OnOpen != nil {
			s.OnOpen(c)
		}

		//Handle client concurrently
		go c.Handle(handler)
	}
	//return nil
}

/*
Write sends a message to all clients.
*/
func (s *Server) Write(message []byte) (int, error) {
	for _, client := range s.Clients {
		go client.Write(message)
	}
	return 0, nil
}

/*
WriteString writes a string to the server.
*/
func (s *Server) WriteString(message string) (int, error) {
	asBytes := []byte(message)
	return s.Write(asBytes)
}
