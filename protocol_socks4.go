package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	RepSocks4Granted     = 0x5a
	RepSocks4Rejected    = 0x5b
	RepSocks4NoIdentd    = 0x5c
	RepSocks4InvalidUser = 0x5d
)

//Ver|CD|DstPort|DstIP|USERID|0|HostName|0
type ReqSocks4 struct {
	Ver      byte
	CD       byte   //1 connect 2  bind
	DstPort  []byte //2 bytes
	DstIP    []byte
	UserId   []byte
	HostName []byte
}

func NewReqSocks4(cmd byte, addr string) (*ReqSocks4, error) {

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("addr:%s SplitHostPort %v", addr, err)
	}

	dstIP := make([]byte, 4)

	var hostname []byte
	if ip := net.ParseIP(host); ip != nil {
		dstIP = ip
	} else {
		dstIP[0] = 0
		dstIP[1] = 0
		dstIP[2] = 0
		dstIP[3] = 1

		hostname = []byte(host)
	}

	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("port:%s ParseUint %v", port, err)
	}

	dstPort := make([]byte, 2)
	dstPort[0] = byte(portNum >> 8)
	dstPort[1] = byte(portNum)

	return &ReqSocks4{
		Ver:      VerSocks4,
		CD:       cmd,
		DstPort:  dstPort,
		DstIP:    dstIP,
		UserId:   nil,
		HostName: hostname,
	}, nil
}

func NewReqSocks4From(r io.Reader) (*ReqSocks4, error) {
	b := make([]byte, 1)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	cd := b[0]

	var payload [6]byte

	_, err = io.ReadFull(r, payload[:])
	if err != nil {
		return nil, err
	}

	port := payload[:2]
	ip := payload[2:]

	userId, err := readUntilNull(r)
	if err != nil {
		return nil, err
	}

	var hostname []byte

	socks4a := (ip[0] == 0 && ip[1] == 0 && ip[2] == 0 && ip[3] != 0)

	if socks4a {
		hostname, err = readUntilNull(r)
		if err != nil {
			return nil, err
		}
	}

	return &ReqSocks4{
		Ver:      0x04,
		CD:       cd,
		DstPort:  port,
		DstIP:    ip,
		UserId:   userId,
		HostName: hostname,
	}, nil
}

func (p *ReqSocks4) Address() string {
	port := int(binary.BigEndian.Uint16(p.DstPort))

	if len(p.HostName) == 0 {
		return net.JoinHostPort(net.IP(p.DstIP).String(), strconv.Itoa(port))
	} else {
		return net.JoinHostPort(string(p.HostName), strconv.Itoa(port))
	}
}

func (p *ReqSocks4) PortIPBytes() []byte {
	ret := make([]byte, 0, 6)
	ret = append(ret, p.DstPort...)
	ret = append(ret, p.DstIP...)
	return ret
}

func (p *ReqSocks4) ToBytes() []byte {
	ret := make([]byte, 0, 20)
	ret = append(ret, p.Ver)
	ret = append(ret, p.CD)
	ret = append(ret, p.DstPort...)
	ret = append(ret, p.DstIP...)
	if len(p.UserId) > 0 {
		ret = append(ret, p.UserId...)
	}
	ret = append(ret, 0)

	if len(p.HostName) > 0 {
		ret = append(ret, p.HostName...)
		ret = append(ret, 0)
	}
	return ret
}

func readUntilNull(reader io.Reader) ([]byte, error) {
	var buf []byte
	var data [1]byte

	for {
		_, err := reader.Read(data[:])
		if err != nil {
			return nil, err
		}
		if data[0] == 0 {
			return buf, nil
		}
		buf = append(buf, data[0])
	}
}

type ReplySocks4 struct {
	VN      byte
	CD      byte
	DstPort []byte
	DstIP   []byte
}

func NewReplySocks4(cd byte, portIp []byte) *ReplySocks4 {
	var port, ip []byte
	if len(portIp) > 0 {
		port = portIp[:2]
		ip = portIp[2:]
	} else {
		port = make([]byte, 2)
		ip = make([]byte, 4)
	}
	return &ReplySocks4{
		VN:      0,
		CD:      cd,
		DstPort: port,
		DstIP:   ip,
	}
}

func NewReplySocks4From(r io.Reader) (*ReplySocks4, error) {
	b := make([]byte, 8)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}

	return &ReplySocks4{
		VN:      b[0],
		CD:      b[1],
		DstPort: b[2:4],
		DstIP:   b[4:],
	}, nil
}

func (p *ReplySocks4) ToBytes() []byte {
	ret := []byte{p.VN, p.CD}
	ret = append(ret, p.DstPort...)
	ret = append(ret, p.DstIP...)
	return ret
}
