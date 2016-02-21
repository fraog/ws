# Simple WebSockets for fun

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


Additionally, you could use a handler with a client connection as well, instead of ReadTo:

	conn.Handle(func(c *ws.Conn, m []byte) {
		fmt.Printf("Client got response: %s\n", string(m))
	})


