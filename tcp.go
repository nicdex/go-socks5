package go_socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type TCPListener struct {
	remoteBindAddr net.Addr
	controlconn    net.Conn
	localListener  net.Listener
	debug          bool
}

type TCPConnection struct {
	debug      bool
	remoteAddr net.Addr
	localAddr  net.Addr
	forward    net.Conn
}

func (c *TCPListener) Accept() (net.Conn, error) {
	buf := make([]byte, 10)
	n, err := c.controlconn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != 10 {
		return nil, errors.New(fmt.Sprintf("expected packet of size 10 got %v", n))
	}
	// set remote address to the one returned by the proxy
	remoteAddr := &net.TCPAddr{
		IP:   buf[4:8],
		Port: int(binary.BigEndian.Uint16(buf[8:10])),
	}
	f, err := c.localListener.Accept()
	if err != nil {
		return nil, err
	}
	if c.debug {
		log.Println("TCPListener on", c.Addr().String(), "accepted connection from", remoteAddr.String())
	}
	return &TCPConnection{
		debug:      c.debug,
		remoteAddr: remoteAddr,
		localAddr:  c.remoteBindAddr,
		forward:    f,
	}, nil
}

func (c *TCPListener) Addr() net.Addr {
	return c.remoteBindAddr
}

func (c *TCPListener) Close() error {
	return c.controlconn.Close()
}

func (c *TCPConnection) Close() error {
	return nil
}

func (c *TCPConnection) Read(b []byte) (int, error) {
	if n, err := c.forward.Read(b); err != nil {
		return n, nil
	} else {
		if c.debug {
			log.Println("SR", "R="+c.remoteAddr.String(), "L="+c.localAddr.String(), n, b[:n])
		}
		return n, nil
	}
}

func (c *TCPConnection) Write(b []byte) (int, error) {
	if n, err := c.forward.Write(b); err != nil {
		return 0, err
	} else {
		if c.debug {
			log.Println("SW", "R="+c.remoteAddr.String(), "L="+c.localAddr.String(), n, b[:n])
		}
		return n, nil
	}
}

func (c *TCPConnection) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *TCPConnection) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *TCPConnection) SetDeadline(t time.Time) error {
	return c.forward.SetDeadline(t)
}

func (c *TCPConnection) SetReadDeadline(t time.Time) error {
	return c.forward.SetReadDeadline(t)
}

func (c *TCPConnection) SetWriteDeadline(t time.Time) error {
	return c.forward.SetWriteDeadline(t)
}
