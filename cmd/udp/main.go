package main

import (
	"encoding/binary"
	"errors"
	"github.com/nicdex/go-socks5"
	"golang.org/x/net/proxy"
	"log"
	"net"
)

func main() {
	//publicAddr, err := net.ResolveTCPAddr("tcp", "216.71.203.33:80")
	//if err != nil {
	//	log.Fatalln("could not resolve public ip", err)
	//}

	//client := &go_socks5.Client{
	//	Addr: "ca472.nordvpn.com:1080",
	//	Username: "rogerbob2@forgetaboutmail.com",
	//	Password: "34394520",
	//	PublicIP: publicAddr.IP,
	//	Debug: true,
	//}
	//log.Println(client)
	d, err := proxy.SOCKS5("tcp4", "ca472.nordvpn.com:1080", &proxy.Auth{
		User:     "rogerbob2@forgetaboutmail.com",
		Password: "34394520",
	}, &net.Dialer{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	testTCPDialDns(d)
	//testTCPDialIPv4(client)
	//testTCPDialIPv6(client)
	//testTCPListen(client, "0.0.0.0:6001")

	testUDPDialDns(d)
	//testUDPDialIPv4(client)
	//testUDPDialIPv6(client)

	//testUDPListenNil(client)
}

func testTCP(conn net.Conn) error {
	_, err := conn.Write([]byte("GET / HTTP/1.0\r\nHost: www.google.com\r\n\r\n"))
	if err != nil {
		return errors.New("can't tcp write: " + err.Error())
	}
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		return errors.New("Can't tcp read: " + err.Error())
	}
	return nil
}

func testUDP(conn net.Conn) error {
	buf := make([]byte, 1024)
	sent, err := conn.Write([]byte("HELLO"))
	if err != nil {
		return errors.New("Can't write packet via socks5: " + err.Error())
	}
	recv, err := conn.Read(buf)
	if err != nil {
		return errors.New("Can't read packet via socks5: " + err.Error())
	}
	if recv != sent {
		return errors.New("Recv count != sent count")
	}
	return nil
}

func testTCPDialDns(client proxy.Dialer) {
	log.Println("testTCPDialDns start")
	conn, err := client.Dial("tcp", "www.google.com:80")
	if err != nil {
		log.Fatalf("Can't tcp dial: %s", err.Error())
	}
	defer conn.Close()
	if err = testTCP(conn); err != nil {
		log.Fatalf("testTCPDialDns failed: %s", err.Error())
	}
	log.Println("testTCPDialDns OK")
}

func testTCPDialIPv4(client *go_socks5.Client) {
	log.Println("testTCPDialIPv4 start")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "www.google.com:80")
	if err != nil {
		log.Fatalf("resolve: %s", err.Error())
	}
	conn, err := client.Dial("tcp", tcpAddr.String())
	if err != nil {
		log.Fatalf("Can't tcp dial: %s", err.Error())
	}
	defer conn.Close()
	if err = testTCP(conn); err != nil {
		log.Fatalf("testTCPDialIPv4 failed: %s", err.Error())
	}
	log.Println("testTCPDialIPv4 OK")
}

func testTCPDialIPv6(client *go_socks5.Client) {
	log.Println("testTCPDialIPv6 start")
	conn, err := client.Dial("tcp6", "[2607:f8b0:400a:809::2004]:80")
	if err != nil {
		log.Fatalf("Can't tcp dial: %s", err.Error())
	}
	defer conn.Close()
	if err = testTCP(conn); err != nil {
		log.Fatalf("testTCPDialIPv6 failed: %s", err.Error())
	}
	log.Println("testTCPDialIPv6 OK")
}

func testUDPDialDns(client proxy.Dialer) {
	log.Println("testUDPDialDns start")
	conn, err := client.Dial("udp", "patrickstar.nicdex.com:6001")
	if err != nil {
		log.Fatalf("Can't udp listen: %s", err.Error())
	}
	defer conn.Close()
	if err = testUDP(conn); err != nil {
		log.Fatalf("testUDPDialDns failed: %s", err.Error())
	}
	log.Println("testUDPDialDns OK")
}

func testUDPDialIPv4(client *go_socks5.Client) {
	log.Println("testUDPDialIPv4 start")
	udpAddr, err := net.ResolveUDPAddr("udp4", "patrickstar.nicdex.com:6001")
	if err != nil {
		log.Fatalf("resolve: %s", err.Error())
	}
	conn, err := client.Dial("udp", udpAddr.String())
	if err != nil {
		log.Fatalf("Can't udp listen: %s", err.Error())
	}
	defer conn.Close()
	if err = testUDP(conn); err != nil {
		log.Fatalf("testUDPDialIPv4 failed: %s", err.Error())
	}
	log.Println("testUDPDialIPv4 OK")
}

func testUDPDialIPv6(client *go_socks5.Client) {
	udpAddr, err := net.ResolveUDPAddr("udp6", "tracker.coppersurfer.tk:80")
	if err != nil {
		log.Fatalf("resolve: %s", err.Error())
	}
	conn, err := client.Dial("udp", udpAddr.String())
	if err != nil {
		log.Fatalf("Can't udp listen: %s", err.Error())
	}
	defer conn.Close()
	if err = testUDP(conn); err != nil {
		log.Fatalf("testUDPDialIPv6 failed: %s", err.Error())
	}
	log.Println("testUDPDialIPv6 OK")
}

func testTCPListen(client *go_socks5.Client, address string) {
	log.Println("testTCPListen start")

	conn, err := client.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Can't tcp listen on %v: %v", address, err.Error())
	}
	defer conn.Close()
	log.Println("Listening on", conn.Addr())

	s, err := conn.Accept()
	if err != nil {
		log.Fatalf("Error accepting connection: %v", err.Error())
	}
	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil {
		log.Fatalf("Error reading: %v", err.Error())
	}
	// echo back to client
	n, err = s.Write(buf[:n])
	if err != nil {
		log.Fatalf("Error writing: %v", err.Error())
	}

	log.Println("testTCPListen OK", n)
}

func testUDPListenNil(client *go_socks5.Client) {
	laddr := &net.UDPAddr{
		IP:   net.IPv4(64, 46, 3, 19), //"64.46.3.19",
		Port: 62314,
	}
	conn, err := client.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalf("Can't udp listen: %s", err.Error())
	}
	defer conn.Close()
	log.Println("Local UDP Addr:", conn.LocalAddr())

	dstAddr, err := net.ResolveUDPAddr("udp4", "tracker.coppersurfer.tk:80")
	if err != nil {
		log.Fatalf("Can't resolve addr: %s", err.Error())
	}

	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:], 0x41727101980)
	binary.BigEndian.PutUint32(b[8:], 0)
	binary.BigEndian.PutUint32(b[12:], 123)
	_, err = conn.WriteTo(b, dstAddr)
	if err != nil {
		log.Fatalf("Can't write packet via socks5: %s", err.Error())
	}
	_, recvAddr, err := conn.ReadFrom(b)
	if err != nil {
		log.Fatalf("Can't read packet via socks5: %s", err.Error())
	}
	log.Println("testUDPListenNil: OK", recvAddr)
}
