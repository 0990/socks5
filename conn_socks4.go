package socks5

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type Socks4Conn struct {
	conn net.Conn
	cfg  ServerCfg
}

func (p *Socks4Conn) Handle() error {
	req, err := p.readRequest()
	if err != nil {
		return fmt.Errorf("readRequest:%w", err)
	}

	return p.handleRequest(req)
}

func (p *Socks4Conn) readRequest() (*ReqSocks4, error) {
	req, err := NewReqSocks4From(p.conn)
	if err != nil {
		return nil, fmt.Errorf("readRequest:%w", err)
	}
	return req, nil
}

func (p *Socks4Conn) handleRequest(req *ReqSocks4) error {
	switch req.CD {
	case CmdConnect:
		return p.handleConnect(req)
	default:
		p.conn.Write(NewReplySocks4(RepSocks4Rejected, nil).ToBytes())
		return ErrCmdNotSupport
	}
}

func (p *Socks4Conn) handleConnect(req *ReqSocks4) error {
	addr := req.Address()
	logrus.Debug("tcp req:", addr)
	s, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		p.conn.Write(NewReplySocks4(RepSocks4Rejected, nil).ToBytes())
		logrus.WithError(err).Debugf("connect to %v failed", req.Address())
		return nil
	}
	defer s.Close()

	_, err = p.conn.Write(NewReplySocks4(RepSocks4Granted, req.PortIPBytes()).ToBytes())
	if err != nil {
		return fmt.Errorf("reply:%w", err)
	}

	timeout := time.Duration(p.cfg.TCPTimeout) * time.Second
	go func() {
		copyWithTimeout(p.conn, s, timeout)
	}()

	copyWithTimeout(s, p.conn, timeout)
	return nil
}
