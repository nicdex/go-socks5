package go_socks5

import (
	"net"
)

type TCPConnection struct {
	forward	net.Conn
}

