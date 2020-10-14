package socks5

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)

type Socks5Conn struct {
	conn net.Conn
	cfg  ServerCfg
}

func (p *Socks5Conn) Handle() error {
	defer p.conn.Close()

	method, err := p.selectAuthMethod()
	if err != nil {
		return fmt.Errorf("selectAuthMethod:%w", err)
	}

	err = p.checkAuthMethod(method)
	if err != nil {
		return fmt.Errorf("checkAuthMethod:%w", err)
	}

	req, err := p.readRequest()
	if err != nil {
		return fmt.Errorf("readRequest:%w", err)
	}

	return p.handleRequest(req)
}

func (p *Socks5Conn) selectAuthMethod() (byte, error) {
	req, err := NewMethodSelectReqFrom(p.conn)
	if err != nil {
		return 0, fmt.Errorf("NewMethodSelectReqFrom:%w", err)
	}

	if req.Ver != VerSocks5 {
		return 0, ErrSocksVersion
	}

	var method byte = MethodNone
	if p.cfg.UserName != "" && p.cfg.Password != "" {
		method = MethodUserPass
	}

	var exist bool
	for _, v := range req.Methods {
		if method == v {
			exist = true
			break
		}
	}

	if !exist {
		method = MethodNoAcceptable
	}

	_, err = p.conn.Write(NewMethodSelectReply(method).ToBytes())
	if err != nil {
		return 0, fmt.Errorf("reply:%w", err)
	}

	if method == MethodNoAcceptable {
		return 0, ErrMethodNoAcceptable
	}

	return method, nil
}

func (p *Socks5Conn) checkAuthMethod(method byte) error {
	switch method {
	case MethodNone:
		return nil
	case MethodUserPass:
		req, err := NewUserPassAuthReqFrom(p.conn)
		if err != nil {
			return fmt.Errorf("NewUserPassAuthReqFrom:%w", err)
		}

		if req.Ver != VerAuthUserPass {
			return ErrAuthUserPassVer
		}

		var status byte = AuthStatusFailure
		if string(req.UserName) == p.cfg.UserName && string(req.Password) == p.cfg.Password {
			status = AuthStatusSuccess
		}

		_, err = p.conn.Write(NewUserPassAuthReply(status).ToBytes())
		if err != nil {
			return fmt.Errorf("reply:%w", err)
		}

		if status != AuthStatusSuccess {
			return ErrAuthFailed
		}
		return nil
	default:
		return ErrMethod
	}
}

//TODO suport bind
func (p *Socks5Conn) handleRequest(req *Request) error {
	switch req.Cmd {
	case CmdConnect:
		return p.handleConnect(req)
	case CmdUDP:
		return p.handleUDP(req)
	default:
		p.conn.Write(NewReply(RepCmdNotSupported, nil).ToBytes())
		return ErrCmdNotSupport
	}
}

func (p *Socks5Conn) handleUDP(req *Request) error {
	addr := p.conn.LocalAddr().(*net.TCPAddr)
	saddr := addr.String()
	//docker环境中获取不了本机正确ip,这时需要从事先设置的环境变量中获取
	if p.cfg.UDPAddr != "" {
		saddr = p.cfg.UDPAddr
	}
	bAddr, err := NewAddrByteFromString(saddr)
	if err != nil {
		p.conn.Write(NewReply(RepServerFailure, nil).ToBytes())
		return err
	}
	_, err = p.conn.Write(NewReply(RepSuccess, bAddr).ToBytes())
	if err != nil {
		return err
	}
	return nil
}

func (p *Socks5Conn) handleConnect(req *Request) error {
	addr := req.Address()
	logrus.Debug("tcp req:", addr)
	s, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		msg := err.Error()
		var rep byte = RepHostUnreachable
		if strings.Contains(msg, "refused") {
			rep = RepConnectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			rep = RepNetworkUnreachable
		}
		p.conn.Write(NewReply(rep, nil).ToBytes())
		return fmt.Errorf("Connect to %v failed: %w", req.Address(), err)
	}
	defer s.Close()

	bAddr, err := NewAddrByteFromString(s.LocalAddr().(*net.TCPAddr).String())
	if err != nil {
		p.conn.Write(NewReply(RepServerFailure, nil).ToBytes())
		return fmt.Errorf("NewAddrByteFromString:%w", err)
	}

	_, err = p.conn.Write(NewReply(RepSuccess, bAddr).ToBytes())
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

func copyWithTimeout(dst net.Conn, src net.Conn, timeout time.Duration) error {
	b := make([]byte, socketBufSize)
	for {
		if timeout != 0 {
			src.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err := src.Read(b)
		if err != nil {
			return fmt.Errorf("copy read:%w", err)
		}
		wn, err := dst.Write(b[0:n])
		if err != nil {
			return fmt.Errorf("copy write:%w", err)
		}
		if wn != n {
			return fmt.Errorf("copy write not full")
		}
	}
	return nil
}

func (p *Socks5Conn) readRequest() (*Request, error) {
	req, err := NewRequestFrom(p.conn)
	if err != nil {
		return nil, fmt.Errorf("readRequest:%w", err)
	}
	return req, nil
}
