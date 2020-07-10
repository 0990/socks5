package socks5

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)

type Server interface {
	Run() error
}

func NewServer(cfg ServerCfg) Server {
	return newServer(cfg)
}

type server struct {
	listener   net.Listener
	cfg        ServerCfg
	listenAddr string

	envAddr string //环境变量中的服务器IP（docker专用）
}

func newServer(cfg ServerCfg) *server {
	p := &server{
		cfg:        cfg,
		listenAddr: fmt.Sprintf(":%d", cfg.ListenPort),
		envAddr:    os.Getenv("PROXY_ADDR"),
	}
	return p
}

func (p *server) Run() error {
	err := p.listen()
	if err != nil {
		return err
	}
	go p.serve()
	go runUDPRelayServer(p.listenAddr, time.Duration(p.cfg.UDPTimout)*time.Second)
	return nil
}

func (p *server) listen() error {
	l, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}
	p.listener = l
	return nil
}

func (p *server) serve() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			return
		}
		go p.connHandler(conn)
	}
}

func (p *server) connHandler(conn net.Conn) {
	c := &Socks5Conn{
		conn:    conn,
		cfg:     p.cfg,
		envAddr: p.envAddr,
	}
	err := c.Handle()
	if err != nil {
		logrus.WithError(err).Error("socks5 conn handle")
	}
}
