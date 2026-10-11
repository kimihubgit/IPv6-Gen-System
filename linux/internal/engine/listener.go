package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"ipv6-gen-linux/internal/store"
)

// Zero-allocation 32KB buffer pool for ultra-fast streaming
var bufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024)
		return &b
	},
}

// BufferedConn wraps net.Conn with a fallback reader to consume any bytes already buffered in bufio.Reader
type BufferedConn struct {
	net.Conn
	r io.Reader
}

func (b *BufferedConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// PortListener listens on a TCP port and handles HTTP/SOCKS5 proxy connections
type PortListener struct {
	account  *store.ProxyAccount
	listener net.Listener
	ipNet    *net.IPNet
	sessions *SessionManager
	stopChan chan struct{}
	mu       sync.RWMutex
}

// NewPortListener creates a listener for a specific proxy account
func NewPortListener(acc *store.ProxyAccount, ipNet *net.IPNet) (*PortListener, error) {
	addr := fmt.Sprintf("0.0.0.0:%d", acc.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("không thể mở cổng %d: %w", acc.Port, err)
	}

	pl := &PortListener{
		account:  acc,
		listener: ln,
		ipNet:    ipNet,
		sessions: NewSessionManager(),
		stopChan: make(chan struct{}),
	}

	go pl.acceptLoop()
	return pl, nil
}

// Close terminates the listener
func (pl *PortListener) Close() {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	select {
	case <-pl.stopChan:
		return
	default:
		close(pl.stopChan)
		if pl.listener != nil {
			_ = pl.listener.Close()
		}
		if pl.sessions != nil {
			pl.sessions.Close()
		}
	}
}

func (pl *PortListener) acceptLoop() {
	for {
		clientConn, err := pl.listener.Accept()
		if err != nil {
			select {
			case <-pl.stopChan:
				return
			default:
				continue
			}
		}

		go pl.handleConn(clientConn)
	}
}

func (pl *PortListener) handleConn(clientConn net.Conn) {
	defer clientConn.Close()

	TuneTCPConn(clientConn)

	// Check if account is active & within quota
	if ok, reason := pl.account.CanAccess(); !ok {
		log.Printf("⛔ Từ chối kết nối cổng %d: %s", pl.account.Port, reason)
		return
	}

	pl.account.LastActive = time.Now()

	// Peek first byte to auto-detect protocol
	bufReader := bufio.NewReader(clientConn)
	firstByte, err := bufReader.Peek(1)
	if err != nil {
		return
	}

	wrappedClient := &BufferedConn{
		Conn: clientConn,
		r:    io.MultiReader(bufReader, clientConn),
	}

	// 0x05 -> SOCKS5 Protocol
	if firstByte[0] == 0x05 {
		pl.handleSOCKS5(wrappedClient, bufReader)
	} else {
		// HTTP / HTTPS Proxy Protocol
		pl.handleHTTP(wrappedClient, bufReader)
	}
}

// ----------------------------------------------------------------------
// HTTP & HTTPS CONNECT PROXY
// ----------------------------------------------------------------------

func (pl *PortListener) handleHTTP(clientConn net.Conn, br *bufio.Reader) {
	req, err := http.ReadRequest(br)
	if err != nil {
		return
	}

	// Check authentication if configured
	sessionKey := ""
	if pl.account.Username != "" && pl.account.Password != "" {
		authHeader := req.Header.Get("Proxy-Authorization")
		if authHeader == "" {
			_, _ = clientConn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"IPv6 Proxy Hub\"\r\n\r\n"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
			_, _ = clientConn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\n\r\n"))
			return
		}

		payload, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			_, _ = clientConn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}

		cred := strings.SplitN(string(payload), ":", 2)
		if len(cred) != 2 {
			_, _ = clientConn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}

		baseUser, sKey := ExtractSessionKey(cred[0])
		sessionKey = sKey

		if baseUser != pl.account.Username || cred[1] != pl.account.Password {
			_, _ = clientConn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\n\r\n"))
			return
		}
	}

	// Pick rotated IPv6 according to session and rotation policy
	outIP := pl.sessions.PickIPv6(sessionKey, pl.account, pl.ipNet)

	targetHost := req.URL.Host
	if targetHost == "" {
		targetHost = req.Host
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	destConn, _, err := DialOutbound(ctx, targetHost, outIP)
	if err != nil {
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer destConn.Close()

	countedClient := NewCountingConn(clientConn, pl.account)
	countedDest := NewCountingConn(destConn, pl.account)

	if req.Method == http.MethodConnect {
		// HTTPS Tunnel: Respond with 200 Connection Established and begin bidirectional streaming
		_, err = countedClient.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			return
		}
		relayBidirectional(countedClient, countedDest)
	} else {
		// HTTP Plain Forwarding
		req.Header.Del("Proxy-Authorization")
		req.Header.Del("Proxy-Connection")

		var buf bytes.Buffer
		_ = req.Write(&buf)
		_, err = countedDest.Write(buf.Bytes())
		if err != nil {
			return
		}
		relayBidirectional(countedClient, countedDest)
	}
}

// ----------------------------------------------------------------------
// SOCKS5 PROXY
// ----------------------------------------------------------------------

func (pl *PortListener) handleSOCKS5(clientConn net.Conn, br *bufio.Reader) {
	// Handshake
	ver, err := br.ReadByte()
	if err != nil || ver != 0x05 {
		return
	}

	nmethods, err := br.ReadByte()
	if err != nil {
		return
	}

	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(br, methods); err != nil {
		return
	}

	hasAuth := pl.account.Username != "" && pl.account.Password != ""
	if hasAuth {
		// Respond requiring username/password authentication (0x02)
		if _, err := clientConn.Write([]byte{0x05, 0x02}); err != nil {
			return
		}

		// Read Auth
		authVer, err := br.ReadByte()
		if err != nil || authVer != 0x01 {
			return
		}

		userLen, err := br.ReadByte()
		if err != nil {
			return
		}
		userBytes := make([]byte, userLen)
		if _, err := io.ReadFull(br, userBytes); err != nil {
			return
		}

		passLen, err := br.ReadByte()
		if err != nil {
			return
		}
		passBytes := make([]byte, passLen)
		if _, err := io.ReadFull(br, passBytes); err != nil {
			return
		}

		baseUser, sKey := ExtractSessionKey(string(userBytes))
		if baseUser != pl.account.Username || string(passBytes) != pl.account.Password {
			_, _ = clientConn.Write([]byte{0x01, 0x01}) // Auth Failed
			return
		}

		// Auth OK
		if _, err := clientConn.Write([]byte{0x01, 0x00}); err != nil {
			return
		}

		pl.handleSOCKS5Request(clientConn, br, sKey)
	} else {
		// No Auth (0x00)
		if _, err := clientConn.Write([]byte{0x05, 0x00}); err != nil {
			return
		}
		pl.handleSOCKS5Request(clientConn, br, "")
	}
}

func (pl *PortListener) handleSOCKS5Request(clientConn net.Conn, br *bufio.Reader, sessionKey string) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(br, header); err != nil {
		return
	}

	if header[0] != 0x05 || header[1] != 0x01 { // Only CMD 0x01 (CONNECT) is supported
		_, _ = clientConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var targetHost string
	switch header[3] {
	case 0x01: // IPv4
		ipv4 := make([]byte, 4)
		if _, err := io.ReadFull(br, ipv4); err != nil {
			return
		}
		targetHost = net.IP(ipv4).String()

	case 0x03: // Domain
		dlen, err := br.ReadByte()
		if err != nil {
			return
		}
		domain := make([]byte, dlen)
		if _, err := io.ReadFull(br, domain); err != nil {
			return
		}
		targetHost = string(domain)

	case 0x04: // IPv6
		ipv6 := make([]byte, 16)
		if _, err := io.ReadFull(br, ipv6); err != nil {
			return
		}
		targetHost = net.IP(ipv6).String()

	default:
		_, _ = clientConn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var portBytes [2]byte
	if _, err := io.ReadFull(br, portBytes[:]); err != nil {
		return
	}
	port := int(portBytes[0])<<8 | int(portBytes[1])
	targetAddr := fmt.Sprintf("%s:%d", targetHost, port)

	outIP := pl.sessions.PickIPv6(sessionKey, pl.account, pl.ipNet)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	destConn, _, err := DialOutbound(ctx, targetAddr, outIP)
	if err != nil {
		_, _ = clientConn.Write([]byte{0x05, 0x04, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer destConn.Close()

	// SOCKS5 Connection Established Reply
	_, err = clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return
	}

	countedClient := NewCountingConn(clientConn, pl.account)
	countedDest := NewCountingConn(destConn, pl.account)
	relayBidirectional(countedClient, countedDest)
}

// relayBidirectional connects two streams using a zero-allocation buffer pool.
// When one direction reaches EOF, it immediately sends a TCP FIN via CloseWrite()
// without any artificial sleeping or waiting for timeout deadlines!
func relayBidirectional(conn1, conn2 net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	pipe := func(dst, src net.Conn) {
		defer wg.Done()
		bufPtr := bufferPool.Get().(*[]byte)
		defer bufferPool.Put(bufPtr)

		_, _ = io.CopyBuffer(dst, src, *bufPtr)
		if tcpConn, ok := dst.(interface{ CloseWrite() error }); ok {
			_ = tcpConn.CloseWrite()
		} else {
			_ = dst.Close()
		}
	}

	go pipe(conn1, conn2)
	go pipe(conn2, conn1)

	wg.Wait()
}
