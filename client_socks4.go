package socks5

import (
	"fmt"
	"net"
)

type socks4Client struct {
	cfg                   ClientCfg
	tcpConn               *net.TCPConn
	handshakeSuccCallback func(conn *net.TCPConn)
}

func NewSocks4Client(cfg ClientCfg) *socks4Client {
	return &socks4Client{
		cfg:                   cfg,
		tcpConn:               nil,
		handshakeSuccCallback: nil,
	}
}

func (p *socks4Client) SetHandshakeSuccCallback(cb func(c *net.TCPConn)) {
	p.handshakeSuccCallback = cb
}

func (p *socks4Client) Dial(network string, addr string) (net.Conn, error) {
	if network != "tcp" {
		return nil, fmt.Errorf("not support network:%s", network)
	}

	conn, err := net.Dial("tcp", p.cfg.ServerAddr)
	if err != nil {
		return nil, err
	}

	p.tcpConn = conn.(*net.TCPConn)

	_, err = p.request(CmdConnect, addr)
	if err != nil {
		return nil, err
	}

	if p.handshakeSuccCallback != nil {
		p.handshakeSuccCallback(p.tcpConn)
	}

	return p.tcpConn, nil
}

func (p *socks4Client) request(cmd byte, addr string) (*ReplySocks4, error) {
	req, err := NewReqSocks4(cmd, addr)
	if err != nil {
		return nil, err
	}
	_, err = p.tcpConn.Write(req.ToBytes())
	if err != nil {
		return nil, err
	}

	reply, err := NewReplySocks4From(p.tcpConn)
	if err != nil {
		return nil, err
	}

	if reply.CD != RepSocks4Granted {
		return nil, fmt.Errorf("reply failure:%d", reply.VN)
	}

	return reply, nil
}
