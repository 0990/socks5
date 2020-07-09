package socks5

import (
	"github.com/0990/socks5/config"
	"net"
	"testing"
)

func TestServer_NoAuth(t *testing.T) {
	ServerTest(config.Server{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
	}, config.Client{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func TestServer_UserPassAuth(t *testing.T) {
	ServerTest(config.Server{
		ListenPort: 1080,
		UserName:   "0990",
		Password:   "123456",
	}, config.Client{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "0990",
		Password:   "123456",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func TestServer_UDP(t *testing.T) {
	ServerUDPTest(config.Server{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, config.Client{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func ServerTest(s config.Server, c config.Client, t *testing.T) {
	ss := NewServer(s)
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTest(c, nil, t)
}

func ServerUDPTest(s config.Server, c config.Client, t *testing.T) {
	ss := NewServer(s)
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTestUDP(c, nil, t)
}

func TestServer_UDP_TcpDisconnect(t *testing.T) {
	ss := NewServer(config.Server{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
	})
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTestUDP_TCPDisconnect(config.Client{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, func(conn *net.TCPConn) {
		conn.Close()
	}, t)
}
