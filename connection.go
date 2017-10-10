package go_socks5

import (
	"net"
	"errors"
	"fmt"
	"log"
	"strconv"
	"encoding/binary"
)

type Connection struct {
	client   *Client
	conn	 net.Conn
}

func (c Connection) writePacket(buf []byte) error {
	rc, err := c.conn.Write(buf)
	if err != nil {
		return err
	}
	if rc != len(buf) {
		return errors.New(fmt.Sprintf("proxy: couldn't write all data: %d/%d", rc, len(buf)))
	}
	if c.client.Debug {
		log.Println("W", buf)
	}
	return nil
}

func (c Connection) readPacket(buf []byte) (int, error) {
	rc, err := c.conn.Read(buf)
	if err != nil {
		return rc, err
	}
	if c.client.Debug {
		log.Println("R", buf[:rc])
	}
	return rc, nil
}

func (c Connection) authenticate() error {
	//TODO support other method than username/password
	buf := make([]byte, MaxProtoSize)
	buf[0] = 5 // Socks 5
	buf[1] = 1 // 1 Method
	buf[2] = 2 // username/password
	if err := c.writePacket(buf[0:3]); err != nil {
		return err
	}
	if rc, err := c.readPacket(buf); err != nil {
		return err
	} else if rc != 2 {
		return errors.New(fmt.Sprintf("proxy: unexpected response packet size: %d", rc))
	} else if buf[1] == 0xff {
		return errors.New("proxy: no acceptable authentication method")
	}

	username_len := len(c.client.Username)
	password_len := len(c.client.Password)

	buf[0] = 1
	buf[1] = byte(username_len)
	copy(buf[2:], c.client.Username)
	buf[2+username_len] = byte(password_len)
	copy(buf[3+username_len:], c.client.Password)
	if err := c.writePacket(buf[0:3+username_len+password_len]); err != nil {
		return err
	}
	if rc, err := c.readPacket(buf); err != nil {
		return err
	} else if rc != 2 {
		return errors.New(fmt.Sprintf("proxy: unexpected response packet size: %d", rc))
	} else if buf[1] != 0 {
		return errors.New(fmt.Sprintf("proxy: authentication failed: status=%x", buf[1]))
	}

	return nil
}

func (c Connection) udpAssociate(laddr *net.UDPAddr) (*net.UDPAddr, error) {
	buf := make([]byte, MaxProtoSize)
	buf[0] = 5 // Ver
	buf[1] = 3 // UDP Associate
	buf[2] = 0 // Reserved
	buf[3] = 1 // IPv4
	//All blank if local network
	for k := 4; k < 10; k++ {
		buf[k] = 0
	}
	//TODO support ipv6, domain
	if err := c.writePacket(buf[0:10]); err != nil {
		return nil, err
	}

	if rc, err := c.readPacket(buf); err != nil {
		return nil, err
	} else if rc != 10 {
		return nil, errors.New(fmt.Sprintf("proxy: unexpected response packet size: %d", rc))
	} else if buf[1] != 0 {
		return nil, errors.New(fmt.Sprintf("proxy: udp associate failed: status=%x", buf[1]))
	}

	return &net.UDPAddr{
		IP:   buf[4:8],
		Port: int(binary.BigEndian.Uint16(buf[8:10])),
	}, nil
}

func (c Connection) connect(address string) error {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("proxy: failed to parse port number: " + portStr)
	}
	if port < 1 || port > 0xffff {
		return errors.New("proxy: port number out of range: " + portStr)
	}

	buf := make([]byte, MaxProtoSize)
	buf[0] = 5 // Ver
	buf[1] = 1 // Connect
	buf[2] = 0 // Reserved

	size := 4
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buf[3] = 1 // Ipv4
			size += copy(buf[size:], ip4)
		} else {
			buf[3] = 4 // Ipv6
			size += copy(buf[size:], ip)
		}
	} else {
		host_len := len(host)
		if host_len > 255 {
			return errors.New("proxy: hostname too long: " + host)
		}
		buf[3] = 3 // domain
		buf[size] = byte(host_len)
		size++
		size += copy(buf[size:], host)
	}
	binary.BigEndian.PutUint16(buf[size:], uint16(port))
	size += 2

	if err := c.writePacket(buf[:size]); err != nil {
		return err
	}

	if rc, err := c.readPacket(buf); err != nil {
		return err
	} else if rc < 10 {
		return errors.New("unexpected response")
	} else if buf[1] != 0 {
		return errors.New(fmt.Sprintf("connect failed: %d", buf[1]))
	}

	//TODO should we use the returned bind address as LocalAddr???

	return nil
}
