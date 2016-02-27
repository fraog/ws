# Simple WebSockets for fun

## Installation:

	go get github.com/fraog/ws

## Examples:

Creating a very simple WebSocket server that replies with "Hello, client.":

	server, err := ws.Listen(":1337")
	if err != nil {
		//Error starting server
	}

	server.Serve(func(c *ws.Conn, m []byte) {
		fmt.Printf("Server got message: %s\n", string(m))
		c.Write([]byte("Hello, client."))
	})

Or, with less code:

	server, err := ws.ListenAndServe(":1337", func(c *ws.Conn, m []byte) {
		fmt.Printf("Server got message: %s\n", string(m))
		c.Write([]byte("Hello, client."))
	})

Sending data to a WebSocket server:

	conn, err := ws.Dial(":1337")
	if err != nil {
		//Error dialing server
	}

	_, err = conn.Write([]byte("Hello, server."))
	if err != nil {
		//Error writing to server
	}

You could then read a response from the server to stdout (or any other Writer) like so:

	err = conn.ReadTo(os.Stdout)
	if err != nil {
		//Error reading response from server
	}


Additionally, you could use a handler function with a client connection, instead of ReadTo:

	conn.Handle(func(c *ws.Conn, m []byte) {
		fmt.Printf("Client got response: %s\n", string(m))
	})

## Handlers:

Handlers provide an additional layer of processing data.

Below is the basic implementation of MessageHandler. It extracts a key from the beginning of the message, finds the handler function associated with that key (if it exists),
and routes the rest of the data to it. The return value of Handle signals whether or not the message was handled. If the return value is false, the data is sent to the default handler function (if it exists):

	type MessageHandler map[string]func(*Conn, []byte) bool

	func (mh *MessageHandler) Handle(c *Conn, msg []byte) bool {
			
		buffer := bytes.NewBuffer(msg)
		
		key, e := buffer.ReadString(':')
		if e != nil {
			return false
		}
		key = key[:len(key)]
		
		if handler, k := (*mh)[key]; k {
			return handler(c, buffer.Bytes())
		}
		
		return false
	}

You could use a MessageHandler like so:

	server, err := ws.Listen(":1337")
	if err != nil {
		//Error starting server
	}

	mh := ws.NewMessageHandler()
	mh["greeting"] = func(c *ws.Conn, m []byte) bool {
		fmt.Printf("Client said hello: %s\n", string(m))
		c.Write([]byte("Hello, client."))
		return true
	}
	
	mh["farewell"] = func(c *ws.Conn, m []byte) bool {
		fmt.Printf("Client said goodbye: %s\n", string(m))
		c.Write([]byte("Goodbye, client."))
		return true
	}
	
	server.Handler = &mh
	server.Serve(func(c *ws.Conn, m []byte) {
		fmt.Printf("Server got unhandled message: %s\n", string(m))
	})
	
A client connection can also have a Handler:

	conn, err := ws.Dial(":1337")
	if err != nil {
		//Error dialing server
	}
	
	mh := ws.NewMessageHandler()
	mh["print"] = func(c *ws.Conn, m []byte) bool {
		fmt.Printf(string(m))
		return true
	}
	
	conn.Handler = &mh
	conn.Handle()



