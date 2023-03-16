package socks5

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Socks5Conn struct {
	conn net.Conn
	cfg  ServerCfg

	udpListenAddr *net.UDPAddr
}

func (p *Socks5Conn) Handle() error {
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

// TODO suport bind
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
	addrAdv := p.getUDPAdvAddr()
	bAddr, err := NewAddrByteFromString(addrAdv)
	if err != nil {
		p.conn.Write(NewReply(RepServerFailure, nil).ToBytes())
		return err
	}
	_, err = p.conn.Write(NewReply(RepSuccess, bAddr).ToBytes())
	if err != nil {
		return err
	}

	buf := make([]byte, 32)
	for {
		//p.conn.SetDeadline(time.Time{})
		_, err := p.conn.Read(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Socks5Conn) getUDPAdvAddr() string {
	port := p.udpListenAddr.Port

	//docker等环境中获取不了本机正确ip,这时需要从事先设置的配置或环境变量中获取
	if len(p.cfg.UDPAdvertisedIP) > 0 {
		return net.JoinHostPort(p.cfg.UDPAdvertisedIP, strconv.FormatInt(int64(port), 10))
	}

	localAddr := p.conn.LocalAddr().(*net.TCPAddr)
	addr := net.UDPAddr{
		IP:   localAddr.IP,
		Zone: localAddr.Zone,
		Port: port,
	}

	return addr.String()
}

func (p *Socks5Conn) handleConnect(req *Request) error {
	addr := req.Address()
	logrus.Debug("tcp req:", addr)
	s, err := net.DialTimeout("tcp", addr, time.Second*3)
	if err != nil {
		msg := err.Error()
		var rep byte = RepHostUnreachable
		if strings.Contains(msg, "refused") {
			rep = RepConnectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			rep = RepNetworkUnreachable
		}
		p.conn.Write(NewReply(rep, nil).ToBytes())
		logrus.WithError(err).Debugf("connect to %v failed", req.Address())
		return nil
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

func (p *Socks5Conn) readRequest() (*Request, error) {
	req, err := NewRequestFrom(p.conn)
	if err != nil {
		return nil, fmt.Errorf("readRequest:%w", err)
	}
	return req, nil
}
