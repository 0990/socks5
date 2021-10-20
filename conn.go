package socks5

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	VerSocks4 = 0x04
	VerSocks5 = 0x05
)

type Conn struct {
	conn net.Conn
	cfg  ServerCfg
}

func (p *Conn) Handle() error {
	defer p.conn.Close()

	ver := make([]byte, 1)
	_, err := io.ReadFull(p.conn, ver)
	if err != nil {
		if err == io.EOF {
			return nil
		}
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
		return c.Handle()
	default:
		return errors.New("unsupport socks version")
	}
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
