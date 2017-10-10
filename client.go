package go_socks5

import (
	"net"
	"errors"
	"fmt"
)

const MaxProtoSize = 262

type Client struct {
	Addr     string
	Username string
	Password string
	Debug    bool
}

func (c Client) dialTCP(network, address string) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr(network, c.Addr)
	if err != nil {
		return nil, errors.New("can't resolve socks5 server: " + err.Error())
	}
	conn, err := net.DialTCP(network, nil, addr)
	if err != nil {
		return nil, errors.New("can't connect to socks5 server: " + err.Error())
	}
	connection := &Connection{
		conn: conn,
		client: &c,
	}
	if err := connection.authenticate(); err != nil {
		return nil, err
	}
	if err := connection.connect(address); err != nil {
		return nil, err
	}
	return conn, nil
}

func (c Client) dialUDP(network, address string) (*UDPConnection, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, errors.New("can't connect to socks5 server: " + err.Error())
	}
	connection := &Connection{
		conn: conn,
		client: &c,
	}
	if err := connection.authenticate(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	bindAddr, err := connection.udpAssociate(nil)
	if err != nil {
		return nil, err
	}
	f, err := net.DialUDP(bindAddr.Network(), nil, bindAddr)
	if err != nil {
		return nil, err
	}
	return &UDPConnection{
		targetAddr:	 address,
		forward:     f,
		controlconn: conn,
		debug:       c.Debug,
	}, nil
}

func (c *Client) Dial(network, address string) (net.Conn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		return c.dialUDP(network, address)
	case "tcp", "tcp4", "tcp6":
		return c.dialTCP(network, address)
	default:
		return nil, errors.New("unsupported network")
	}
}

func (c *Client) DialTCP(network string, laddr, raddr *net.TCPAddr) (*net.TCPConn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		break
	default:
		return nil, errors.New("wrong network type: " + network)
	}
	if raddr == nil {
		return nil, errors.New("missing remote address")
	}
	return c.dialTCP(network, raddr.String())
}

func (c *Client) DialUDP(network string, laddr, raddr *net.UDPAddr) (*UDPConnection, error) {
	switch network {
	case "udp", "udp4", "udp6":
		break
	default:
		return nil, errors.New("wrong network type: " + network)
	}
	if raddr == nil {
		return nil, errors.New("missing remote address")
	}
	return c.dialUDP(network, raddr.String())
}

func (c *Client) Listen(network, address string) (net.Listener, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error) {
	return nil, errors.New("not implemented")
}

func (c Client) ListenUDP(network string, laddr *net.UDPAddr) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		break
	default:
		return nil, errors.New(fmt.Sprintf("wrong network type: %s", network))
	}

	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, errors.New("can't connect to socks5 server: " + err.Error())
	}

	if laddr == nil {
		t := conn.LocalAddr().(*net.TCPAddr)
		laddr = &net.UDPAddr{
			IP: t.IP,
			Port: 0,
		}
	}

	connection := &Connection{
		client: &c,
		conn: conn,
	}

	if err := connection.authenticate(); err != nil {
		return nil, err
	}

	bindAddr, err := connection.udpAssociate(laddr)
	if err != nil {
		return nil, err
	}

	f, err := net.DialUDP(bindAddr.Network(), laddr, bindAddr)
	if err != nil {
		return nil, err
	}

	return &UDPConnection{
		controlconn: conn,
		debug:       c.Debug,
		forward:     f,
	}, nil
}
