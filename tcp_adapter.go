package go2p

import (
	"fmt"
	"net"
)

type adapterTCP struct {
	conn net.Conn
}

// NewAdapter creates a new TCP adapter that wraps the given net.Conn instance
func NewAdapter(conn net.Conn) Adapter {
	a := new(adapterTCP)
	a.conn = conn
	return a
}

func (a *adapterTCP) ReadMessage() (*Message, error) {
	m := NewMessage()
	err := m.ReadFromConn(a.conn)
	return m, err
}

func (a *adapterTCP) WriteMessage(m *Message) error {
	err := m.WriteIntoConn(a.conn)
	return err
}

func (a *adapterTCP) Close() {
	a.conn.Close()
}

func (a *adapterTCP) RemoteAddress() string {
	addr := a.conn.RemoteAddr()
	res := fmt.Sprintf("%s:%s", addr.Network(), addr.String())
	return res
}

func (a *adapterTCP) LocalAddress() string {
	addr := a.conn.LocalAddr()
	res := fmt.Sprintf("%s:%s", addr.Network(), addr.String())
	return res
}
