package socks5

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

//you should start a socks5 server before test
//for me,use ss5 because it support udp

const TEST_SERVER_ADDR = "127.0.0.1:1086"

func TestClient_NoAuth(t *testing.T) {
	ClientTest(ClientCfg{
		ServerAddr: TEST_SERVER_ADDR,
		UserName:   "",
		Password:   "",
	}, t)
}

func TestClient_UserPassAuth(t *testing.T) {
	ClientTest(ClientCfg{
		ServerAddr: TEST_SERVER_ADDR,
		UserName:   "0990",
		Password:   "123456",
	}, t)
}

func ClientTest(cfg ClientCfg, t *testing.T) {
	sc := NewSocks5Client(cfg)

	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sc.Dial(network, addr)
			},
		},
	}
	resp, err := hc.Get("http://whatismyip.akamai.com/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.FailNow()
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}

func Socks4ClientTest(cfg ClientCfg, t *testing.T) {
	sc := NewSocks4Client(cfg)

	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sc.Dial(network, addr)
			},
		},
	}
	resp, err := hc.Get("http://whatismyip.akamai.com/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.FailNow()
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}

func TestClient_UDP(t *testing.T) {
	ClientTestUDP(ClientCfg{
		ServerAddr: TEST_SERVER_ADDR,
		UserName:   "",
		Password:   "",
		UDPTimout:  60,
	}, t)
}

func ClientTestUDP(cfg ClientCfg, t *testing.T) {
	sc := NewSocks5Client(cfg)
	conn, err := sc.Dial("udp", "8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}

	//one dns query packet
	b, err := hex.DecodeString("0001010000010000000000000a74787468696e6b696e6703636f6d0000010001")
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write(b); err != nil {
		t.Fatal(err)
	}

	b = make([]byte, 2048)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	b = b[:n]
	b = b[len(b)-4:]
	fmt.Println(net.IPv4(b[0], b[1], b[2], b[3]))
}

func TestClient_UDP_TcpDisconnect(t *testing.T) {
	ClientTestUDP_TCPDisconnect(ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, func(conn *net.TCPConn) {
		conn.Close()
	}, t)
}

func ClientTestUDP_TCPDisconnect(cfg ClientCfg, handshakeCB func(conn *net.TCPConn), t *testing.T) {
	sc := NewSocks5Client(cfg)
	sc.SetHandshakeSuccCallback(handshakeCB)

	conn, err := sc.Dial("udp", "8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}

	//one dns query packet
	b, err := hex.DecodeString("0001010000010000000000000a74787468696e6b696e6703636f6d0000010001")
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write(b); err != nil {
		t.Fatal(err)
	}

	b = make([]byte, 2048)
	_, err = conn.Read(b)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "i/o timeout" {
			return
		}
	}

	t.FailNow()
}
