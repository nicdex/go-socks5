package go_socks5

import (
	"net"
	"errors"
	"log"
	"encoding/binary"
	"time"
	"fmt"
	"strconv"
)

//TODO packet size / fragmentation

type UDPConnection struct {
	targetAddr	string
	controlconn net.Conn
	forward     net.Conn
	debug       bool
}

func (u *UDPConnection) Read(b []byte) (int, error) {
	if n, addr, err := u.ReadFrom(b); err != nil {
		return n, nil
	} else if addr.String() != u.targetAddr {
		//TODO handle wrong addr
		return n, nil
	} else {
		return n, nil
	}
}

func (u *UDPConnection) ReadFrom(b []byte) (int, net.Addr, error) {
	if len(b) == 0 {
		return 0, nil, nil
	}

	buf := make([]byte, len(b) + MaxProtoSize)
	rc, err := u.forward.Read(buf)
	if err != nil {
		return 0, nil, err
	}
	//TODO support ipv6 and domain
	var from_addr *net.UDPAddr
	var n int
	switch buf[3] {
	case 1: // IPv4
		n = rc - 10
		from_addr = &net.UDPAddr{
			IP: buf[4:8],
			Port: int(binary.BigEndian.Uint16(buf[8:])),
		}
		copy(b, buf[10:rc])
		break
	default:
		return 0, nil, errors.New(fmt.Sprintf("unsupported atyp: %d", buf[3]))
	}
	if u.debug {
		log.Println("SR", from_addr, n, b)
	}
	return n, from_addr, nil
}

func (u *UDPConnection) Write(b []byte) (int, error) {
	host, portStr, err := net.SplitHostPort(u.targetAddr)
	if err != nil {
		return 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, errors.New("proxy: failed to parse port number: " + portStr)
	}
	if port < 1 || port > 0xffff {
		return 0, errors.New("proxy: port number out of range: " + portStr)
	}

	buf := make([]byte, MaxProtoSize + len(b))
	buf[0] = 0
	buf[1] = 0
	buf[2] = 0

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
			return 0, errors.New("proxy: hostname too long: " + host)
		}
		buf[3] = 3 // domain
		buf[size] = byte(host_len)
		size++
		size += copy(buf[size:], host)
	}
	binary.BigEndian.PutUint16(buf[size:], uint16(port))
	size += 2

	copy(buf[size:], b)
	size += len(b)

	if rc, err := u.forward.Write(buf[:size]); err != nil {
		return 0, err
	} else {
		if u.debug {
			log.Println("SW", u.forward.RemoteAddr().String(), rc, buf[:size])
		}
		return rc - 10, nil
	}
}

func (u *UDPConnection) WriteTo(b []byte, addr net.Addr) (int, error) {
	var ip []byte
	var ipLen int
	var port int
	switch addr.(type) {
	case *net.UDPAddr:
		udpAddr := addr.(*net.UDPAddr)
		ipLen = 4
		ip = udpAddr.IP.To4()
		port = udpAddr.Port
		break
	default:
		return 0, errors.New("unsupported address type")
	}

	buf := make([]byte, len(b) + 6 +ipLen)
	buf[0] = 0
	buf[1] = 0
	buf[2] = 0
	if ipLen == 4 {
		buf[3] = 1
	} else {
		return 0, errors.New(fmt.Sprintf("unsupported ip length: %d", ipLen))
	}
	copy(buf[4:], ip)
	binary.BigEndian.PutUint16(buf[8:], uint16(port))
	copy(buf[10:], b)

	if rc, err := u.forward.Write(buf); err != nil {
		return 0, err
	} else {
		if u.debug {
			log.Println("SW", u.forward.RemoteAddr().String(), rc, buf)
		}
		return rc - 10, nil
	}
}

func (u *UDPConnection) Close() error {
	u.controlconn.Close()
	return u.forward.Close()
}

func (u *UDPConnection) LocalAddr() net.Addr {
	return u.forward.LocalAddr()
}

func (u *UDPConnection) RemoteAddr() net.Addr {
	//TODO
	return nil
}

func (u *UDPConnection) SetDeadline(t time.Time) error {
	return u.forward.SetDeadline(t)
}

func (u *UDPConnection) SetReadDeadline(t time.Time) error {
	return u.forward.SetReadDeadline(t)
}

func (u *UDPConnection) SetWriteDeadline(t time.Time) error {
	return u.forward.SetWriteDeadline(t)
}