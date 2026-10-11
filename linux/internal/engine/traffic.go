package engine

import (
	"net"

	"ipv6-gen-linux/internal/store"
)

// CountingConn wraps a net.Conn to count bytes read and written
type CountingConn struct {
	net.Conn
	account *store.ProxyAccount
}

// NewCountingConn wraps conn with an account traffic counter
func NewCountingConn(conn net.Conn, acc *store.ProxyAccount) *CountingConn {
	return &CountingConn{
		Conn:    conn,
		account: acc,
	}
}

func (c *CountingConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 && c.account != nil {
		c.account.AddTraffic(int64(n))
	}
	return n, err
}

func (c *CountingConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 && c.account != nil {
		c.account.AddTraffic(int64(n))
	}
	return n, err
}

// CloseWrite delegates half-close to the underlying connection
func (c *CountingConn) CloseWrite() error {
	if cw, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return nil
}
