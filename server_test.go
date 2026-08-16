package socks5

import (
	"net"
	"path/filepath"
	"testing"
)

// startTestServer 启动一个server,并在测试结束时自动关闭
func startTestServer(t *testing.T, cfg ServerCfg) {
	t.Helper()

	ss, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := ss.Run(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ss.Close()
	})
}

func TestServer_CreateConfig(t *testing.T) {
	if err := CreateServerCfg(filepath.Join(t.TempDir(), "ss5.json")); err != nil {
		t.Fatal(err)
	}
}

func TestServer_NoAuth(t *testing.T) {
	ServerTest(ServerCfg{
		ListenPort: 13305,
		UserName:   "",
		Password:   "",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13305",
		UserName:   "",
		Password:   "",
	}, t)
}

func TestServer_NoAuth_TCPListen(t *testing.T) {
	ServerTest(ServerCfg{
		ListenPort: 13306,
		TCPListen:  "127.0.0.1:13307",
		UserName:   "",
		Password:   "",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13307",
		UserName:   "",
		Password:   "",
	}, t)
}

func TestServer_UserPassAuth(t *testing.T) {
	ServerTest(ServerCfg{
		ListenPort: 13308,
		UserName:   "0990",
		Password:   "123456",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13308",
		UserName:   "0990",
		Password:   "123456",
	}, t)
}

func TestServer_UDP(t *testing.T) {
	ServerUDPTest(ServerCfg{
		ListenPort: 13309,
		UserName:   "",
		Password:   "",
		UDPTimout:  30,
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13309",
		UserName:   "",
		Password:   "",
		UDPTimout:  5,
	}, t)
}

func TestServer_UDPListen(t *testing.T) {
	ServerUDPTest(ServerCfg{
		ListenPort:      13310,
		UDPListen:       "0.0.0.0:13311",
		UDPAdvertisedIP: "",
		UserName:        "",
		Password:        "",
		UDPTimout:       30,
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13310",
		UserName:   "",
		Password:   "",
		UDPTimout:  5,
	}, t)
}

func TestServer_UDPAdverisedIP(t *testing.T) {
	ServerUDPTest(ServerCfg{
		ListenPort:      13312,
		UDPListen:       "0.0.0.0:13313",
		UDPAdvertisedIP: "127.0.0.1",
		UserName:        "",
		Password:        "",
		UDPTimout:       30,
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13312",
		UserName:   "",
		Password:   "",
		UDPTimout:  5,
	}, t)
}

func ServerTest(s ServerCfg, c ClientCfg, t *testing.T) {
	startTestServer(t, s)
	ClientTest(c, t)
}

func ServerUDPTest(s ServerCfg, c ClientCfg, t *testing.T) {
	startTestServer(t, s)
	ClientTestUDP(c, t)
}

// TODO UDP association在TCP控制连接断开后应终止,此功能还没有实现,测试暂时跳过
func TestServer_UDP_TcpDisconnect(t *testing.T) {
	t.Skip("UDP association teardown on TCP disconnect not implemented yet")

	startTestServer(t, ServerCfg{
		ListenPort: 13316,
		UserName:   "",
		Password:   "",
		UDPTimout:  30,
	})

	ClientTestUDP_TCPDisconnect(ClientCfg{
		ServerAddr: "127.0.0.1:13316",
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, func(conn *net.TCPConn) {
		conn.Close()
	}, t)
}

func TestServer_Socks4(t *testing.T) {
	Socks4ServerTest(ServerCfg{
		ListenPort: 13314,
		UserName:   "",
		Password:   "",
	}, ClientCfg{
		ServerAddr: "127.0.0.1:13314",
		UserName:   "",
		Password:   "",
	}, t)
}

func Socks4ServerTest(s ServerCfg, c ClientCfg, t *testing.T) {
	startTestServer(t, s)
	Socks4ClientTest(c, t)
}
