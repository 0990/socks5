package socks5

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type Server interface {
	Run() error
}

func NewServer(cfg ServerCfg) (Server, error) {
	return newServer(cfg)
}

type server struct {
	listener net.Listener
	cfg      ServerCfg

	tcpListenAddr *net.TCPAddr
	udpListenAddr *net.UDPAddr
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
		go p.connHandler(conn)
	}
}

func (p *server) connHandler(conn net.Conn) {
	c := &Conn{
		conn:          conn,
		cfg:           p.cfg,
		udpListenAddr: p.udpListenAddr,
	}

	err := c.Handle()
	if err != nil {
		logrus.WithError(err).Error("conn handle")
	}
}
