package main

import (
	"fmt"
	"github.com/fraog/ws"
	"log"
	"net/http"
)

func main() {

	//Create a websocket server on :7331
	websockets, err := ws.Listen(":7331")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer websockets.Close()

	//Broadcast a message when a client connects or disconnects.
	websockets.OnOpen = func(conn *ws.Conn) {
		_, err := websockets.WriteString(fmt.Sprintf("Client connected from %s", conn.Base().RemoteAddr()))
		if err != nil {
			log.Fatal(err)
		}
	}

	websockets.OnClose = func(conn *ws.Conn) {
		_, err := websockets.WriteString(fmt.Sprintf("Client disconnected from %s", conn.Base().RemoteAddr()))
		if err != nil {
			log.Fatal(err)
		}
	}

	//Start serving as a Goroutine
	go websockets.Serve(func(conn *ws.Conn, msg []byte) {
		//Send the message to all clients
		_, err := websockets.Write(msg)
		if err != nil {
			log.Fatal(err)
		}
	})

	//Now, start a simple http file server
	log.Fatal(http.ListenAndServe(":1337", http.FileServer(http.Dir("./client"))))
}
