package ws

/*
A Namespace is a subset of clients connected to the server.
*/
type Namespace struct {
	name     string
	clients  map[int64]*Conn
	children map[string]*Namespace
}

/*
Creates a new Namespace.
*/
func NewNamespace(n string) *Namespace {
	var ns Namespace
	ns = Namespace{n, make(map[int64]*Conn), make(map[string]*Namespace)}
	return &ns
}

/*
Add client to the namespace.
*/
func (n *Namespace) AddClient(c *Conn) {
	n.clients[c.Id()] = c
}

/*
Remove client from the namespace.
*/
func (n *Namespace) RemoveClient(c *Conn) {
	delete(n.clients, c.Id())
}

/*
Add a child namespace.
*/
func (n *Namespace) AddChild(nc *Namespace) {
	n.children[nc.name] = nc
}

/*
Remove a child namespace.
*/
func (n *Namespace) RemoveChild(nc *Namespace) {
	delete(n.children, nc.name)
}

/*
Write a message to all clients in this namespace and all clients of child namespaces.
*/
func (n *Namespace) Write(b []byte) (int, error) {
	for _, client := range n.clients {
		go client.Write(b)
	}
	for _, child := range n.children {
		child.Write(b)
	}
	return 0, nil
}

/*
Write a message as a string to all clients in this namespace and all clients of child namespaces.
*/
func (n *Namespace) WriteString(msg string) (int, error) {
	return n.Write([]byte(msg))
}
