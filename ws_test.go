package ws

import (
	"fmt"
	//"os"
	"testing"
	"time"
	"bytes"
)

func TestFrame(t *testing.T) {

	df := NewFrame([]byte("Test Message"))
	if df.fin != msbOn {
		t.Errorf("FIN is not set: %b", df.fin)
	}

	if (df.fin | df.op) != stdResponse {
		t.Errorf("One ore more reserved bits is set: %b", df.fin|df.op)
	}

	if df.pl != 12 {
		t.Errorf("Payload indicator doesn't match size: %b", df.pl)
	}

	if df.length != 12 {
		t.Errorf("Length doesn't match message size: %v", df.length)
	}

	if df.masked != 0 {
		t.Errorf("Client data frame is marked as masked.")
	}

	if df.mask != nil {
		t.Errorf("Found mask on client data frame.")
	}

}

func TestMaskIO(t *testing.T) {

	msg := "Test Message"
	mb := []byte(msg)
	//fmt.Printf("Message:%s\n", string(mb))

	df := NewFrame(mb)
	df.GenerateMask()

	//Encode mb into writer(rb)
	rb := make([]byte, 0, len(msg)) // <--- Must have size of 0
	writer := bytes.NewBuffer(rb)
	df.Encode(mb, writer)
	rb = writer.Bytes()
	//fmt.Printf("Encoded:%s\n", string(rb))

	//Decode rb from a reader(rb) into slice
	db := make([]byte, len(msg))
	reader := bytes.NewReader(rb)
	df.Decode(reader, db)

	//fmt.Printf("Decoded:%s\n", string(db))
	dbs := string(db)
	if dbs != msg {
		t.Errorf("Masking/Unmasking failure.")
	}

}

func TestClientServer(t *testing.T) {
	
	var err error
	
	var server *Server
	server, err = Listen(":7331")
	if err != nil {
		t.Errorf("Error starting server.", err)
	}

	go server.Serve(func(c *Conn, m []byte) {
		fmt.Printf("Server got message: %s\n", string(m))
		c.Write([]byte("Hello, client."))
		return
	})

	//Sleep for a bit to let the server listen
	time.Sleep(100 * time.Millisecond)

	//Dial the server
	var conn *Conn
	conn, err = Dial(":7331")
	if err != nil {
		t.Errorf("Error dialing server.", err)
	}
	
	//Add a callback on a goroutine
	go conn.Handle(func(c *Conn, m []byte) { //<-- this one allows for possible polymorphism?
		fmt.Printf("Client got message: %s\n", string(m))
		return
	})
	
	//Sleep for a bit to let the client listen
	time.Sleep(100 * time.Millisecond)
	
	//Write to the connection
	_, err = conn.Write([]byte("Hello, server."))
	if err != nil {
		t.Errorf("Error writing to server connection.", err)
	}
	
	/*
	//Now read from it
	err = conn.ReadTo(os.Stdout)
	if err != nil {
		t.Errorf("Error reading response from server.", err)
	}
	*/
	
	time.Sleep(1000 * time.Millisecond)
	
	
	//fmt.Printf("Close server and connection.\n")
	server.Close()
	conn.Close()
}





