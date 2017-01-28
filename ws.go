package ws

//FIXME: The server should validate origin?

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
)

const (
	wsHash = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

/*
Dial dials a connection to a webserver at the specified host.
Returns the connection or an error if no connection was made.
*/
func Dial(host string) (*Conn, error) {
	c, e := net.Dial("tcp", host)
	if e != nil {
		return nil, e
	}

	//Send an http websocket request
	req := createRequest("/ws")
	e = req.Write(c)
	if e != nil {
		return nil, e
	}

	//Read Response
	res, e := http.ReadResponse(bufio.NewReader(c), nil)
	if e != nil {
		return nil, e
	}

	if res == nil {
		return nil, errors.New("Did not get accept response from server.")
	}

	//TODO: Verify accept key?
	//fmt.Printf("AcceptKey: %s\n", res.Header.Get("Sec-WebSocket-Accept"))

	return &Conn{c, true, 0, nil, nil, nil}, nil
}

/*
DialProtocol dials a connection to a webserver with protocols specified.
Returns the connection or an error if no connection was made.
*/
func DialProtocol(host string, proto string) (*Conn, error) {
	c, e := net.Dial("tcp", host)
	if e != nil {
		return nil, e
	}

	//Send an http websocket request
	req := createRequest("/ws")
	e = req.Write(c)
	if e != nil {
		return nil, e
	}

	//Add protocol to header
	req.Header.Add("Sec-WebSocket-Protocol", proto)

	//Read response
	reader := bufio.NewReader(c)
	res, e := http.ReadResponse(reader, nil)
	if e != nil {
		return nil, e
	}

	if res == nil {
		return nil, errors.New("Did not get accept response from server.")
	}

	//TODO: Verify accept key?
	//fmt.Printf("AcceptKey: %s\n", res.Header.Get("Sec-WebSocket-Accept"))

	//TODO: Verify protocol?
	//fmt.Printf("AcceptProto: %s\n", res.Header.Get("Sec-WebSocket-Protocol"))

	return &Conn{c, true, 0, nil, nil, nil}, nil
}

/*
Listen creates a new server, listening on host.
Returns nil and error if an error was encountered.
*/
func Listen(host string) (*Server, error) {
	var s net.Listener
	var e error
	if s, e = net.Listen("tcp", host); e != nil {
		return nil, e
	}
	return &Server{s, make(map[int64]*Conn), nil, nil, nil, nil, nil}, nil
}

/*
ListenAndServe creates a new server, listening on host and serves with handler function.
Returns an error if encountered.
*/
func ListenAndServe(host string, handler func(*Conn, []byte)) error {
	var s *Server
	var e error
	if s, e = Listen(host); e != nil {
		return e
	}
	return s.Serve(handler)
}

/*
Creates a websocket key for a websocket request
*/
func createRequestHash() string {
	b := make([]byte, 8)
	rand.Read(b)
	//TODO: ERROR
	return base64.StdEncoding.EncodeToString(b)
}

/*
Creates an http.Request to send to a websocket server.
*/
func createRequest(host string) *http.Request { //, protocols string) {
	req, _ := http.NewRequest("GET", host, nil)
	//TODO: Error handle
	req.Header.Add("Upgrade", "websocket")
	req.Header.Add("Connection", "Upgrade")
	req.Header.Add("Sec-WebSocket-Key", createRequestHash())
	req.Header.Add("Sec-WebSocket-Version", "13")
	return req
}

/*
Creates the hash to send back to a client if the request is accepted.
*/
func createAcceptHash(key string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sha1.Sum(nil))
}

/*
Creates the HTTP Response to send back to the client if the request is accepted.
*/
func createAcceptResponse(req *http.Request) *http.Response {
	response := new(http.Response)
	response.StatusCode = http.StatusSwitchingProtocols
	response.Header = http.Header{
		"Upgrade":              []string{"websocket"},
		"Connection":           []string{"Upgrade"},
		"Sec-WebSocket-Accept": []string{createAcceptHash(req.Header.Get("Sec-WebSocket-Key"))},
	}
	//FIXME: For now just accept any protocols
	ptcl := req.Header.Get("Sec-WebSocket-Protocol")
	if len(ptcl) > 0 {
		response.Header.Add("Sec-WebSocket-Protocol", req.Header.Get("Sec-WebSocket-Protocol"))
	}
	return response
}
