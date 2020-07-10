package socks5

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type client struct {
	cfg                   ClientCfg
	tcpConn               *net.TCPConn
	handshakeSuccCallback func(conn *net.TCPConn)
}

func NewClient(cfg ClientCfg) *client {
	return &client{
		cfg:                   cfg,
		tcpConn:               nil,
		handshakeSuccCallback: nil,
	}
}

func (p *client) SetHandshakeSuccCallback(cb func(c *net.TCPConn)) {
	p.handshakeSuccCallback = cb
}

func (p *client) Dial(network, addr string) (net.Conn, error) {
	var cmd byte
	switch network {
	case "tcp":
		cmd = CmdConnect
	case "udp":
		cmd = CmdUDP
	default:
		return nil, errors.New("network error")
	}

	bRemoteAddr, err := NewAddrByteFromString(addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", p.cfg.ServerAddr)
	if err != nil {
		return nil, err
	}

	p.tcpConn = conn.(*net.TCPConn)

	method, err := p.selectAuthMethod()
	if err != nil {
		return nil, err
	}

	err = p.authMethod(method)
	if err != nil {
		return nil, err
	}

	//TODO support local udp addr
	var dstAddr AddrByte
	if cmd == CmdConnect {
		dstAddr = bRemoteAddr
	}

	reply, err := p.request(cmd, dstAddr)
	if err != nil {
		return nil, err
	}

	if p.handshakeSuccCallback != nil {
		p.handshakeSuccCallback(p.tcpConn)
	}

	if cmd == CmdConnect {
		return p.tcpConn, nil
	} else {
		udpConn, err := net.Dial("udp", reply.Address())
		if err != nil {
			return nil, err
		}
		return &SocksUDPConn{
			UDPConn: udpConn.(*net.UDPConn),
			dstAddr: bRemoteAddr,
			timeout: time.Duration(p.cfg.UDPTimout) * time.Second,
		}, nil
	}
}

func (p *client) selectAuthMethod() (byte, error) {
	methods := []byte{MethodNone}
	if p.cfg.UserName != "" && p.cfg.Password != "" {
		methods = append(methods, MethodUserPass)
	}

	_, err := p.tcpConn.Write(NewMethodSelectReq(methods).ToBytes())
	if err != nil {
		return 0, err
	}

	reply, err := NewMethodSelectReplyFrom(p.tcpConn)
	if err != nil {
		return 0, err
	}
	if reply.Ver != VerSocks5 {
		return 0, ErrSocksVersion
	}
	if reply.Method == MethodNoAcceptable {
		return 0, ErrMethodNoAcceptable
	}

	return reply.Method, nil
}

func (p *client) authMethod(method byte) error {
	switch method {
	case MethodNone:
		return nil
	case MethodUserPass:
		_, err := p.tcpConn.Write(NewUserPassAuthReq([]byte(p.cfg.UserName), []byte(p.cfg.Password)).ToBytes())
		if err != nil {
			return err
		}
		reply, err := NewUserPassAuthReplyFrom(p.tcpConn)
		if err != nil {
			return err
		}
		if reply.Ver != VerAuthUserPass {
			return errors.New("userPassAuthVer!=1")
		}
		if reply.Status != AuthStatusSuccess {
			return ErrAuthFailed
		}
		return nil
	default:
		return ErrMethod
	}
}

func (p *client) request(cmd byte, addrByte AddrByte) (*Reply, error) {
	_, err := p.tcpConn.Write(NewRequest(cmd, addrByte).ToBytes())
	if err != nil {
		return nil, err
	}

	reply, err := NewReplyFrom(p.tcpConn)
	if err != nil {
		return nil, err
	}

	if reply.Ver != VerSocks5 {
		return nil, ErrSocksVersion
	}

	if reply.Rep != RepSuccess {
		return nil, fmt.Errorf("reply failure:%d", reply.Rep)
	}

	return reply, nil
}
