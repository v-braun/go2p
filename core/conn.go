package core

// Conn represents a wrapper around a network connection
type Conn interface {

	// ReadMessage should read from the underline connection
	// and return a Message object until all data was readed
	// The call should block until an entire Message was readed,
	// an error occurred or the underline connection was closed
	ReadMessage() (*Message, error)

	// WriteMessage write the given message to the underline connection
	WriteMessage(m *Message) error

	// Close should close the underline connection
	Close()

	// LocalAddress returns the local address (example: tcp:127.0.0.1:7000)
	LocalAddress() string

	// RemoteAddress returns the remote address (example: tcp:127.0.0.1:7000)
	RemoteAddress() string
}
