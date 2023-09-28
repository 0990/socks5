package socks5

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Server interface {
	Run() error
	SetCustomTcpConnHandler(handler func(conn *net.TCPConn))
}

func NewServer(cfg ServerCfg) (Server, error) {
	return newServer(cfg)
}

type server struct {
	listener net.Listener
	cfg      ServerCfg

	tcpListenAddr *net.TCPAddr
	udpListenAddr *net.UDPAddr

	customTcpConnHandler func(conn *net.TCPConn)
}

func newServer(cfg ServerCfg) (*server, error) {
	listenAddr := fmt.Sprintf(":%d", cfg.ListenPort)

	tcpAddress := listenAddr
	if len(cfg.TCPListen) > 0 {
		tcpAddress = cfg.TCPListen
	}

	taddr, err := net.ResolveTCPAddr("tcp", tcpAddress)
	if err != nil {
		return nil, err
	}

	udpAddress := listenAddr
	if len(cfg.UDPListen) > 0 {
		udpAddress = cfg.UDPListen
	}

	uaddr, err := net.ResolveUDPAddr("udp", udpAddress)
	if err != nil {
		return nil, err
	}

	p := &server{
		cfg:           cfg,
		tcpListenAddr: taddr,
		udpListenAddr: uaddr,
	}
	return p, nil
}

func (p *server) Run() error {
	err := p.listen()
	if err != nil {
		return err
	}
	go p.serve()
	go runUDPRelayServer(p.udpListenAddr, time.Duration(p.cfg.UDPTimout)*time.Second)
	return nil
}

func (p *server) listen() error {
	l, err := net.ListenTCP("tcp", p.tcpListenAddr)
	if err != nil {
		return err
	}
	p.listener = l
	return nil
}

func (p *server) serve() {
	var tempDelay time.Duration

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			if ne, ok := err.(*net.OpError); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logrus.Errorf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}

		go p.tcpConnHandler(conn)
	}
}

func (p *server) tcpConnHandler(conn net.Conn) {
	if p.customTcpConnHandler != nil {
		p.customTcpConnHandler(conn.(*net.TCPConn))
		return
	}

	p.defaultTcpConnHandler(conn)
}

func (p *server) defaultTcpConnHandler(conn net.Conn) {
	c := &Conn{
		conn: conn,
		cfg: ConnCfg{
			UserName:          p.cfg.UserName,
			Password:          p.cfg.Password,
			TCPTimeout:        int32(p.cfg.TCPTimeout),
			UDPAdvertisedIP:   p.cfg.UDPAdvertisedIP,
			UDPAdvertisedPort: p.udpListenAddr.Port,
		},
	}

	err := c.Handle()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			logrus.WithError(err).Debug("conn handle")
		}
	}
}

func (p *server) SetCustomTcpConnHandler(handler func(conn *net.TCPConn)) {
	p.customTcpConnHandler = handler
}
