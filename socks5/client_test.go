package socks5

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/0990/socks5/config"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

//you should start a socks5 server before test
//for me,use ss5 because it support udp

func TestClient_NoAuth(t *testing.T) {
	ClientTest(config.Client{
		ServerAddr: "10.225.137.202:1080",
		//ServerAddr: "127.0.0.1:1080",
		UserName: "",
		Password: "",
	}, nil, t)
}

func TestClient_UserPassAuth(t *testing.T) {
	ClientTest(config.Client{
		ServerAddr: "10.225.137.202:1080",
		UserName:   "xujialong",
		Password:   "123456",
	}, nil, t)
}

func ClientTest(cfg config.Client, dialcb func(conn *net.TCPConn), t *testing.T) {
	sc := NewClient(cfg, dialcb)

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
	ClientTestUDP(config.Client{
		ServerAddr: "10.225.137.202:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, nil, t)
}

func ClientTestUDP(cfg config.Client, dialcb func(conn *net.TCPConn), t *testing.T) {
	sc := NewClient(cfg, dialcb)
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
	ClientTestUDP_TCPDisconnect(config.Client{
		ServerAddr: "10.225.137.202:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  2,
	}, func(conn *net.TCPConn) {
		conn.Close()
	}, t)
}

func ClientTestUDP_TCPDisconnect(cfg config.Client, dialcb func(conn *net.TCPConn), t *testing.T) {
	sc := NewClient(cfg, dialcb)

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
