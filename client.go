package go_socks5

import (
	"errors"
	"fmt"
	"github.com/anacrolix/missinggo"
	"net"
)

const MaxProtoSize = 262

type Client struct {
	Addr     string
	Username string
	Password string
	PublicIP net.IP
	Debug    bool
}

func (c *Client) connect() (*Connection, error) {
	controlconn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	conn := &Connection{
		client: c,
		conn:   controlconn,
	}
	if c.Username == "" && c.Password == "" {
		return conn, nil
	}

	err = conn.authenticate()
	if err != nil {
		defer controlconn.Close()
		return nil, err
	}

	return conn, nil
}

func (c *Client) dialTCP(network, address string) (*TCPConnection, error) {
	controlconn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, errors.New("can't connect to socks5 server: " + err.Error())
	}
	connection := &Connection{
		conn:   controlconn,
		client: c,
	}
	if err := connection.authenticate(); err != nil {
		return nil, err
	}
	//TODO resolve using socks5
	tcpAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}
	localAddr, err := connection.connect(tcpAddr.String())
	if err != nil {
		return nil, err
	}
	return &TCPConnection{
		debug:      c.Debug,
		localAddr:  localAddr,
		remoteAddr: tcpAddr,
		forward:    controlconn,
	}, nil
}

func (c *Client) dialUDP(network, address string) (*UDPConnection, error) {
	connection, err := c.connect()
	if err != nil {
		return nil, err
	}

	//TODO resolve using proxy
	remoteAddr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}

	relayAddr, err := connection.udpAssociate(nil)
	if err != nil {
		return nil, err
	}

	f, err := net.DialUDP(relayAddr.Network(), nil, relayAddr)
	if err != nil {
		return nil, err
	}

	return &UDPConnection{
		forward:     f,
		controlconn: connection,
		debug:       c.Debug,
		remoteAddr:  remoteAddr,
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

func (c *Client) DialTCP(network string, laddr, raddr *net.TCPAddr) (*TCPConnection, error) {
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
	tcpAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}
	return c.ListenTCP(network, tcpAddr)
}

func (c *Client) ListenTCP(network string, laddr *net.TCPAddr) (net.Listener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		break
	default:
		return nil, errors.New(fmt.Sprintf("wrong network type: %s", network))
	}

	l, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, errors.New("can't listen" + err.Error())
	}

	connection, err := c.connect()
	if err != nil {
		return nil, errors.New("can't connect to socks5 server: " + err.Error())
	}

	addr := laddr
	if c.PublicIP != nil {
		addr = &net.TCPAddr{
			IP:   c.PublicIP,
			Port: addr.Port,
		}
	}
	bindAddr, err := connection.bind(addr)
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		controlconn:    connection.conn,
		remoteBindAddr: bindAddr,
		debug:          c.Debug,
		localListener:  l,
	}, nil
}

func (c *Client) ListenPacket(network string, address string) (net.PacketConn, error) {
	host, port, err := missinggo.ParseHostPort(address)
	if err != nil {
		return nil, err
	}
	addr := &net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}
	return c.ListenUDP(network, addr)
}

func (c *Client) ListenUDP(network string, laddr *net.UDPAddr) (net.PacketConn, error) {
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
			IP:   t.IP,
			Port: 0,
		}
	}

	connection, err := c.connect()
	if err != nil {
		return nil, err
	}

	bindAddr, err := connection.udpAssociate(laddr)
	if err != nil {
		return nil, err
	}

	f, err := net.ListenUDP(bindAddr.Network(), laddr)
	if err != nil {
		return nil, err
	}

	return &UDPConnection{
		controlconn: connection,
		debug:       c.Debug,
		forward:     f,
	}, nil
}
