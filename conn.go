package socks5

import (
	"errors"
	"io"
)

const (
	VerSocks4 = 0x04
	VerSocks5 = 0x05
)

type ConnCfg struct {
	UserName   string
	Password   string
	TCPTimeout int32

	UDPAdvertisedIP   string
	UDPAdvertisedPort int
}

type Conn struct {
	conn Stream
	cfg  ConnCfg

	customDialTarget func(addr string) (Stream, byte, string, error)
}

func NewConn(conn Stream, cfg ConnCfg) *Conn {
	return &Conn{
		conn: conn,
		cfg:  cfg,
	}
}

func (p *Conn) SetCustomDialTarget(f func(addr string) (Stream, byte, string, error)) {
	p.customDialTarget = f
}

func (p *Conn) Handle() error {
	defer p.conn.Close()

	ver := make([]byte, 1)
	_, err := io.ReadFull(p.conn, ver)
	if err != nil {
		return err
	}

	switch ver[0] {
	case VerSocks4:
		c := &Socks4Conn{
			conn: p.conn,
			cfg:  p.cfg,
		}
		return c.Handle()
	case VerSocks5:
		c := &Socks5Conn{
			conn: p.conn,
			cfg:  p.cfg,
		}
		c.SetCustomDialTarget(p.customDialTarget)
		return c.Handle()
	default:
		return errors.New("unsupport socks version")
	}
}
