package socks5

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestClient_NoAuth(t *testing.T) {
	startTestServer(t, ServerCfg{
		ListenPort: 13301,
	})

	ClientTest(ClientCfg{
		ServerAddr: "127.0.0.1:13301",
		UserName:   "",
		Password:   "",
	}, t)
}

func TestClient_UserPassAuth(t *testing.T) {
	startTestServer(t, ServerCfg{
		ListenPort: 13302,
		UserName:   "0990",
		Password:   "123456",
	})

	ClientTest(ClientCfg{
		ServerAddr: "127.0.0.1:13302",
		UserName:   "0990",
		Password:   "123456",
	}, t)
}

func ClientTest(cfg ClientCfg, t *testing.T) {
	sc := NewSocks5Client(cfg)
	httpGetViaProxy(sc.Dial, t)
}

func Socks4ClientTest(cfg ClientCfg, t *testing.T) {
	sc := NewSocks4Client(cfg)
	httpGetViaProxy(sc.Dial, t)
}

// httpGetViaProxy 通过代理访问本地http server,验证tcp connect转发是否正常
func httpGetViaProxy(dial func(network, addr string) (net.Conn, error), t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dial(network, addr)
			},
		},
	}
	resp, err := hc.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code:%d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "ok" {
		t.Fatalf("unexpected body:%s", string(b))
	}
}

func TestClient_UDP(t *testing.T) {
	startTestServer(t, ServerCfg{
		ListenPort: 13303,
		UDPTimout:  30,
	})

	ClientTestUDP(ClientCfg{
		ServerAddr: "127.0.0.1:13303",
		UDPTimout:  5,
	}, t)
}

// ClientTestUDP 通过本地udp echo server验证UDP ASSOCIATE转发是否正常
func ClientTestUDP(cfg ClientCfg, t *testing.T) {
	echoAddr, closeEcho := startUDPEchoServer(t)
	defer closeEcho()

	sc := NewSocks5Client(cfg)
	conn, err := sc.Dial("udp", echoAddr)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("hello socks5 udp")
	if _, err := conn.Write(msg); err != nil {
		t.Fatal(err)
	}

	b := make([]byte, 2048)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b[:n], msg) {
		t.Fatalf("unexpected udp response:%v", b[:n])
	}
}

// startUDPEchoServer 启动一个udp echo server,返回其地址和关闭函数
func startUDPEchoServer(t *testing.T) (string, func()) {
	t.Helper()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			conn.WriteTo(buf[:n], addr)
		}
	}()

	return conn.LocalAddr().String(), func() { conn.Close() }
}

// TODO UDP association在TCP控制连接断开后应终止,此功能还没有实现,测试暂时跳过
func TestClient_UDP_TcpDisconnect(t *testing.T) {
	t.Skip("UDP association teardown on TCP disconnect not implemented yet")

	startTestServer(t, ServerCfg{
		ListenPort: 13315,
		UDPTimout:  30,
	})

	ClientTestUDP_TCPDisconnect(ClientCfg{
		ServerAddr: "127.0.0.1:13315",
		UDPTimout:  2,
	}, func(conn *net.TCPConn) {
		conn.Close()
	}, t)
}

// TODO 这个测试用例待完善，还没有此功能
func ClientTestUDP_TCPDisconnect(cfg ClientCfg, handshakeCB func(conn *net.TCPConn), t *testing.T) {
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
	_, err = conn.Read(b)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "i/o timeout" {
			return
		}
	}

	t.FailNow()
}

func TestClient_TCPTimeout(t *testing.T) {
	echoServerAddr := "127.0.0.1:13322"
	echoCloser, err := StartTCPEchoServer(echoServerAddr, true)
	if err != nil {
		t.Fatal(err)
	}
	defer echoCloser.Close()

	startTestServer(t, ServerCfg{
		ListenPort: 13304,
		TCPTimeout: 2,
	})

	sc := NewSocks5Client(ClientCfg{
		ServerAddr: "127.0.0.1:13304",
	})
	conn, err := sc.Dial("tcp", echoServerAddr)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := conn.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}

	// echo server不回数据,连接进入空闲,服务器应在TCPTimeout后关闭连接
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, err = conn.Read(make([]byte, 1024))
	if err == nil {
		t.Fatal("expect server close the idle connection, but read succeed")
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		t.Fatal("server does not close the idle connection within TCPTimeout")
	}
}

// StartTCPEchoServer noEcho为true时只接收不回数据,返回listener用于关闭
func StartTCPEchoServer(addr string, noEcho bool) (io.Closer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func(conn net.Conn) {
				defer conn.Close()

				log := logrus.WithFields(logrus.Fields{
					"address": conn.RemoteAddr(),
				})

				for {
					buf := make([]byte, 65535)
					n, err := conn.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.WithError(err).Error("read from tcp")
						} else {
							log.WithError(err).Info("read from tcp")
						}
						return
					}

					log.WithField("data", string(buf[0:n])).Info("echoserver tcp receive")
					if !noEcho {
						log.WithField("data", string(buf[0:n])).Info("echoserver echo back")
						conn.Write(buf[0:n])
					}
				}

			}(conn)
		}
	}()
	return listener, nil
}
