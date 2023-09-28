package socks5

import (
	"errors"
	"github.com/0990/gotun/pkg/pool"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// send: client->relayer->sender->remote
// receive: client<-relayer<-sender<-remote
func runUDPRelayServer(listenAddr *net.UDPAddr, timeout time.Duration) {
	relayer, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return
	}
	defer relayer.Close()

	var senders SenderMap

	for {
		buf := pool.GetBuf(MaxSegmentSize)

		n, addr, err := relayer.ReadFrom(buf)
		if err != nil {
			continue
		}
		saddr := addr.String()
		sender, exist := senders.Get(saddr)
		if !exist {
			sender, err = net.ListenPacket("udp", "")
			if err != nil {
				continue
			}
			senders.Add(addr.String(), sender)

			go func() {
				relayToClient(sender, relayer, addr, timeout)
				if sender := senders.Del(saddr); sender != nil {
					sender.Close()
				}
			}()
		}

		err = relayToRemote(sender, buf[0:n])
		if err != nil {
			continue
		}
	}
}

func relayToRemote(sender net.PacketConn, datagram []byte) error {
	d, err := NewUDPDatagramFromBytes(datagram)
	if err != nil {
		return err
	}
	if d.Frag != 0x00 {
		return ErrUDPFrag
	}

	udpTargetAddr := d.Address()

	tgtUDPAddr, err := net.ResolveUDPAddr("udp", udpTargetAddr)
	if err != nil {
		return err
	}

	logrus.Debug("udp req:", udpTargetAddr)

	_, err = sender.WriteTo(d.Data, tgtUDPAddr)
	return err
}

func relayToClient(receiver net.PacketConn, relayer net.PacketConn, clientAddr net.Addr, timeout time.Duration) error {
	buf := pool.GetBuf(MaxSegmentSize)
	defer pool.PutBuf(buf)

	for {
		receiver.SetReadDeadline(time.Now().Add(timeout))
		n, addr, err := receiver.ReadFrom(buf)
		if err != nil {
			return err
		}

		bAddr, err := NewAddrByteFromString(addr.String())
		if err != nil {
			return err
		}

		_, err = relayer.WriteTo(NewUDPDatagram(bAddr, buf[:n]).ToBytes(), clientAddr)
		if err != nil {
			return err
		}
	}
}

type SenderMap struct {
	sync.Map
}

func (p *SenderMap) Add(key string, conn net.PacketConn) {
	p.Map.Store(key, conn)
}

func (p *SenderMap) Del(key string) net.PacketConn {
	if conn, exist := p.Get(key); exist {
		p.Map.Delete(key)
		return conn
	}

	return nil
}

func (p *SenderMap) Get(key string) (net.PacketConn, bool) {
	v, exist := p.Load(key)
	if !exist {
		return nil, false
	}

	return v.(net.PacketConn), true
}

type SocksUDPConn struct {
	*net.UDPConn
	dstAddr      AddrByte
	timeout      time.Duration
	readDeadline time.Time
}

func (p *SocksUDPConn) SetReadDeadline(t time.Time) error {
	p.readDeadline = t
	return nil
}

func (p *SocksUDPConn) Read(b []byte) (int, error) {
	if p.readDeadline.IsZero() {
		if p.timeout != 0 {
			p.UDPConn.SetReadDeadline(time.Now().Add(p.timeout))
		}
	} else {
		p.UDPConn.SetReadDeadline(p.readDeadline)
	}

	buf := pool.GetBuf(MaxSegmentSize)
	n, err := p.UDPConn.Read(buf)
	if err != nil {
		return 0, err
	}
	d, err := NewUDPDatagramFromBytes(buf[0:n])
	if err != nil {
		return 0, err
	}
	if len(b) < len(d.Data) {
		return 0, errors.New("buff too small")
	}
	n = copy(b, d.Data)
	return n, nil
}

func (p *SocksUDPConn) Write(b []byte) (int, error) {
	d := NewUDPDatagram(p.dstAddr, b)
	payload := d.ToBytes()
	n, err := p.UDPConn.Write(payload)
	if err != nil {
		return 0, err
	}
	if len(payload) != n {
		return 0, errors.New("not write full")
	}
	return len(b), nil
}
