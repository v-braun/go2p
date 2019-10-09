package tcp

import (
	"fmt"
	"net"

	"github.com/v-braun/go2p/core"
)

var _ core.Conn = (*tcpConn)(nil)

type tcpConn struct {
	conn net.Conn
}

// NewConn creates a new TCP conn that wraps the given net.Conn instance
func NewConn(conn net.Conn) core.Conn {
	a := new(tcpConn)
	a.conn = conn
	return a
}

func (a *tcpConn) ReadMessage() (*core.Message, error) {
	m := core.NewMessage()
	err := m.ReadFromConn(a.conn)
	return m, err
}

func (a *tcpConn) WriteMessage(m *core.Message) error {
	err := m.WriteIntoConn(a.conn)
	return err
}

func (a *tcpConn) Close() {
	a.conn.Close()
}

func (a *tcpConn) RemoteAddress() string {
	addr := a.conn.RemoteAddr()
	res := fmt.Sprintf("%s:%s", addr.Network(), addr.String())
	return res
}

func (a *tcpConn) LocalAddress() string {
	addr := a.conn.LocalAddr()
	res := fmt.Sprintf("%s:%s", addr.Network(), addr.String())
	return res
}
