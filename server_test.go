package socks5

import (
	"net"
	"testing"
)

func TestServer_CreateConfig(t *testing.T) {
	CreateServerCfg("ss5.json")
}

func TestServer_NoAuth(t *testing.T) {
	ServerTest(ServerCfg{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func TestServer_UserPassAuth(t *testing.T) {
	ServerTest(ServerCfg{
		ListenPort: 1080,
		UserName:   "0990",
		Password:   "123456",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "0990",
		Password:   "123456",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func TestServer_UDP(t *testing.T) {
	ServerUDPTest(ServerCfg{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func ServerTest(s ServerCfg, c ClientCfg, t *testing.T) {
	ss := NewServer(s)
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTest(c, t)
}

func ServerUDPTest(s ServerCfg, c ClientCfg, t *testing.T) {
	ss := NewServer(s)
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTestUDP(c, t)
}

func TestServer_UDP_TcpDisconnect(t *testing.T) {
	ss := NewServer(ServerCfg{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
	})
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTestUDP_TCPDisconnect(ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, func(conn *net.TCPConn) {
		conn.Close()
	}, t)
}

func TestServer_Socks4(t *testing.T) {
	Socks4ServerTest(ServerCfg{
		ListenPort: 1089,
		UserName:   "",
		Password:   "",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func Socks4ServerTest(s ServerCfg, c ClientCfg, t *testing.T) {
	ss := NewServer(s)
	err := ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	Socks4ClientTest(c, t)
}
