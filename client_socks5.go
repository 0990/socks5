package socks5

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type socks5client struct {
	cfg ClientCfg

	handShakeCallback func(cmd byte, reply *Reply)
}

func NewSocks5Client(cfg ClientCfg) *socks5client {
	return &socks5client{
		cfg: cfg,
	}
}

func (p *socks5client) SetHandShakeCallback(callback func(cmd byte, reply *Reply)) {
	p.handShakeCallback = callback
}

func (p *socks5client) Dial(network, addr string) (net.Conn, error) {
	return p.DialTimeout(network, addr, 0)
}

func (p *socks5client) DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
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

	var conn net.Conn
	if timeout > 0 {
		c, err := net.DialTimeout("tcp", p.cfg.ServerAddr, timeout)
		if err != nil {
			return nil, err
		}
		conn = c
	} else {
		c, err := net.Dial("tcp", p.cfg.ServerAddr)
		if err != nil {
			return nil, err
		}
		conn = c
	}

	tcpConn := conn.(*net.TCPConn)

	method, err := p.selectAuthMethod(tcpConn)
	if err != nil {
		return nil, err
	}

	err = p.authMethod(tcpConn, method)
	if err != nil {
		return nil, err
	}

	//TODO support local udp addr
	var dstAddr AddrByte
	if cmd == CmdConnect {
		dstAddr = bRemoteAddr
	}

	reply, err := p.request(tcpConn, cmd, dstAddr)
	if err != nil {
		return nil, err
	}

	if p.handShakeCallback != nil {
		p.handShakeCallback(cmd, reply)
	}

	if cmd == CmdConnect {
		return tcpConn, nil
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

func (p *socks5client) selectAuthMethod(conn *net.TCPConn) (byte, error) {
	methods := []byte{MethodNone}
	if p.cfg.UserName != "" && p.cfg.Password != "" {
		methods = append(methods, MethodUserPass)
	}

	_, err := conn.Write(NewMethodSelectReq(methods).ToBytes())
	if err != nil {
		return 0, err
	}

	reply, err := NewMethodSelectReplyFrom(conn)
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

func (p *socks5client) authMethod(conn *net.TCPConn, method byte) error {
	switch method {
	case MethodNone:
		return nil
	case MethodUserPass:
		_, err := conn.Write(NewUserPassAuthReq([]byte(p.cfg.UserName), []byte(p.cfg.Password)).ToBytes())
		if err != nil {
			return err
		}
		reply, err := NewUserPassAuthReplyFrom(conn)
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

func (p *socks5client) request(conn *net.TCPConn, cmd byte, addrByte AddrByte) (*Reply, error) {
	_, err := conn.Write(NewRequest(cmd, addrByte).ToBytes())
	if err != nil {
		return nil, err
	}

	reply, err := NewReplyFrom(conn)
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
