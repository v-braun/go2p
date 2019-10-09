package go2p

// NetworkConnectionBuilder provides a fluent interface to
// create a NetworkConnection
type NetworkConnectionBuilder struct {
	middlewares []*Middleware
	operators   []PeerOperator
}

// NewNetworkConnection creates a new NetworkBuilder instance to setup a new NetworkConnection
func NewNetworkConnection() *NetworkConnectionBuilder {
	b := new(NetworkConnectionBuilder)

	return b
}

// WithMiddleware attach a new Middleware to the NetworkConnection setup
func (b *NetworkConnectionBuilder) WithMiddleware(name string, impl MiddlewareFunc) *NetworkConnectionBuilder {
	m := NewMiddleware(name, impl)
	b.middlewares = append(b.middlewares, m)
	return b
}

// WithOperator attach a new PeerOperator to the NetworkConnection setup
func (b *NetworkConnectionBuilder) WithOperator(op PeerOperator) *NetworkConnectionBuilder {
	b.operators = append(b.operators, op)
	return b
}

// Build finalize the NetworkConnection setup and creates the new instance
func (b *NetworkConnectionBuilder) Build() *NetworkConnection {
	nc := new(NetworkConnection)
	nc.peers = newPeers()
	nc.middlewares = newMiddlewares(b.middlewares...)
	nc.operators = b.operators
	nc.emitter = newEventEmitter()
	nc.log = newLogger("network-connection")

	return nc
}
