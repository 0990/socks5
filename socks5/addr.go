package socks5

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

const MaxAddrLen = 1 + 1 + 255 + 2
const PortLen = 2

type AddrByte []byte

func (a AddrByte) String() string {
	var host, port string

	switch a[0] { // address type
	case ATypDomainname:
		host = string(a[2 : 2+int(a[1])])
		port = strconv.Itoa((int(a[2+int(a[1])]) << 8) | int(a[2+int(a[1])+1]))
	case ATypIPV4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case ATypIPV6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}

	return net.JoinHostPort(host, port)
}

func (a AddrByte) Split() (aType byte, addr []byte, port []byte) {
	aType = ATypIPV4
	addr = []byte{0, 0, 0, 0}
	port = []byte{0, 0}

	if a != nil {
		aType = a[0]
		addr = a[1 : len(a)-2]
		port = a[len(a)-2:]
	}
	return
}

func NewAddrByteFromString(s string) (AddrByte, error) {
	var addr []byte

	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil, fmt.Errorf("addr:%s SplitHostPort %v", s, err)
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = ATypIPV4
			copy(addr[1:], ip4)
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = ATypIPV6
			copy(addr[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return nil, fmt.Errorf("host:%s too long", host)
		}

		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = ATypDomainname
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("port:%s ParseUint %v", port, err)
	}

	addr[len(addr)-2], addr[len(addr)-1] = byte(portNum>>8), byte(portNum)
	return addr, nil
}

func NewAddrByteFrom(r io.Reader) (AddrByte, error) {
	b := make([]byte, MaxAddrLen)

	_, err := io.ReadFull(r, b[:1])
	if err != nil {
		return nil, err
	}

	var startPos int = 1
	var addrLen int
	switch b[0] {
	case ATypDomainname:
		_, err := io.ReadFull(r, b[1:2])
		if err != nil {
			return nil, err
		}
		startPos++
		addrLen = int(b[1])
	case ATypIPV4:
		addrLen = net.IPv4len
	case ATypIPV6:
		addrLen = net.IPv6len
	default:
		return nil, ErrAddrType
	}

	endPos := startPos + addrLen + PortLen

	_, err = io.ReadFull(r, b[startPos:endPos])
	return b[:endPos], err
}

func NewAddrByteFromByte(b []byte) (AddrByte, error) {
	if len(b) < 1 {
		return nil, ErrBadRequest
	}
	var startPos int = 1
	var addrLen int
	switch b[0] {
	case ATypDomainname:
		if len(b) < 2 {
			return nil, ErrBadRequest
		}
		startPos++
		addrLen = int(b[1])
	case ATypIPV4:
		addrLen = net.IPv4len
	case ATypIPV6:
		addrLen = net.IPv6len
	default:
		return nil, ErrAddrType
	}

	endPos := startPos + addrLen + PortLen

	if len(b) < endPos {
		return nil, ErrBadRequest
	}
	return b[:endPos], nil
}
