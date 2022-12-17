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

func TestServer_NoAuth_TCPListen(t *testing.T) {
	ServerTest(ServerCfg{
		ListenPort: 1080,
		TCPListen:  "127.0.0.1:1081",
		UserName:   "",
		Password:   "",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:1081",
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

func TestServer_UDPListen(t *testing.T) {
	ServerUDPTest(ServerCfg{
		ListenPort: 1080,
		//TCPListen:       "127.0.0.1:1081",
		UDPListen:       "0.0.0.0:1083",
		UDPAdvertisedIP: "",
		UserName:        "",
		Password:        "",
		UDPTimout:       2,
	}, ClientCfg{
		ServerAddr: "10.229.1.56:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func TestServer_UDPAdverisedIP(t *testing.T) {
	ServerUDPTest(ServerCfg{
		ListenPort:      1080,
		UDPListen:       "0.0.0.0:1083",
		UDPAdvertisedIP: "127.0.0.1",
		UserName:        "",
		Password:        "",
		UDPTimout:       2,
	}, ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  0,
		TCPTimeout: 0,
	}, t)
}

func ServerTest(s ServerCfg, c ClientCfg, t *testing.T) {
	ss, err := NewServer(s)
	if err != nil {
		t.Fatal(err)
	}
	err = ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTest(c, t)
}

func ServerUDPTest(s ServerCfg, c ClientCfg, t *testing.T) {
	ss, err := NewServer(s)
	if err != nil {
		t.Fatal(err)
	}
	err = ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	ClientTestUDP(c, t)
}

func TestServer_UDP_TcpDisconnect(t *testing.T) {
	ss, err := NewServer(ServerCfg{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = ss.Run()
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
	ss, err := NewServer(s)
	if err != nil {
		t.Fatal(err)
	}
	err = ss.Run()
	if err != nil {
		t.Fatal(err)
	}

	Socks4ClientTest(c, t)
}
