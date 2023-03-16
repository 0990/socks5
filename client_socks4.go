package socks5

import (
	"fmt"
	"net"
)

type socks4Client struct {
	cfg ClientCfg
}

func NewSocks4Client(cfg ClientCfg) *socks4Client {
	return &socks4Client{
		cfg: cfg,
	}
}

func (p socks4Client) Dial(network string, addr string) (net.Conn, error) {
	if network != "tcp" {
		return nil, fmt.Errorf("not support network:%s", network)
	}

	conn, err := net.Dial("tcp", p.cfg.ServerAddr)
	if err != nil {
		return nil, err
	}

	tcpConn := conn.(*net.TCPConn)

	_, err = p.request(tcpConn, CmdConnect, addr)
	if err != nil {
		return nil, err
	}

	return tcpConn, nil
}

func (p socks4Client) request(conn *net.TCPConn, cmd byte, addr string) (*ReplySocks4, error) {
	req, err := NewReqSocks4(cmd, addr)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(req.ToBytes())
	if err != nil {
		return nil, err
	}

	reply, err := NewReplySocks4From(conn)
	if err != nil {
		return nil, err
	}

	if reply.CD != RepSocks4Granted {
		return nil, fmt.Errorf("reply failure:%d", reply.VN)
	}

	return reply, nil
}
