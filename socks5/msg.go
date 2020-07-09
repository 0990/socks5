package socks5

import (
	"errors"
	"fmt"
	"io"
)

const (
	VerSocks5 = 0x05

	MethodNone         = 0x00
	MethodUserPass     = 0x02
	MethodNoAcceptable = 0xff

	VerAuthUserPass   = 0x01
	AuthStatusSuccess = 0x00
	AuthStatusFailure = 0x01

	CmdConnect     = 0x01
	CmdBind        = 0x02
	CmdUDP         = 0x03
	ATypIPV4       = 0x01
	ATypDomainname = 0x03
	ATypIPV6       = 0x04

	RepSuccess              = 0x00
	RepServerFailure        = 0x01
	RepRuleFailure          = 0x02
	RepNetworkUnreachable   = 0x03
	RepHostUnreachable      = 0x04
	RepConnectionRefused    = 0x05
	RepTTLExpired           = 0x06
	RepCmdNotSupported      = 0x07
	RepAddrTypeNotSupported = 0x08
)

var (
	ErrMethodNoAcceptable = errors.New("no acceptable method")
	ErrAuthFailed         = errors.New("User authentication failed")
	NoSupportedAuth       = errors.New("no supported auth")
	ErrAuthUserPassVer    = errors.New("auth user pass version")
	ErrCmdNotSupport      = errors.New("cmd not support")

	ErrAddrType     = fmt.Errorf("Unrecognized address type")
	ErrSocksVersion = fmt.Errorf("not socks version 5")
	ErrMethod       = fmt.Errorf("Unsupport method")
	ErrBadRequest   = fmt.Errorf("bad request")
	ErrUDPFrag      = fmt.Errorf("Frag !=0 not supported")
)

type MethodSelectReq struct {
	Ver      byte
	NMethods byte
	Methods  []byte
}

func NewMethodSelectReq(methods []byte) *MethodSelectReq {
	return &MethodSelectReq{
		Ver:      VerSocks5,
		NMethods: byte(len(methods)),
		Methods:  methods,
	}
}

func NewMethodSelectReqFrom(r io.Reader) (*MethodSelectReq, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	nMethod := int(b[1])
	methods := make([]byte, nMethod)
	_, err = io.ReadFull(r, methods)
	if err != nil {
		return nil, err
	}
	return &MethodSelectReq{
		Ver:      b[0],
		NMethods: b[1],
		Methods:  methods,
	}, nil
}

func (p *MethodSelectReq) ToBytes() []byte {
	ret := []byte{p.Ver, p.NMethods}
	ret = append(ret, p.Methods...)
	return ret
}

type MethodSelectReply struct {
	Ver    byte
	Method byte
}

func NewMethodSelectReply(method byte) *MethodSelectReply {
	return &MethodSelectReply{
		Ver:    VerSocks5,
		Method: method,
	}
}

func NewMethodSelectReplyFrom(r io.Reader) (*MethodSelectReply, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return &MethodSelectReply{
		Ver:    b[0],
		Method: b[1],
	}, nil
}

func (p *MethodSelectReply) ToBytes() []byte {
	return []byte{p.Ver, p.Method}
}

type UserPassAuthReq struct {
	Ver      byte
	ULen     byte
	UserName []byte
	PLen     byte
	Password []byte
}

func NewUserPassAuthReq(username []byte, password []byte) *UserPassAuthReq {
	return &UserPassAuthReq{
		Ver:      VerAuthUserPass,
		ULen:     byte(len(username)),
		UserName: username,
		PLen:     byte(len(password)),
		Password: password,
	}
}

func NewUserPassAuthReqFrom(r io.Reader) (*UserPassAuthReq, error) {
	b := make([]byte, 1)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	ver := b[0]

	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	uLen := int(b[0])
	userName := make([]byte, uLen)
	_, err = io.ReadFull(r, userName)
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}

	pLen := int(b[0])
	password := make([]byte, pLen)
	_, err = io.ReadFull(r, password)
	if err != nil {
		return nil, err
	}

	return &UserPassAuthReq{
		Ver:      ver,
		ULen:     byte(uLen),
		UserName: userName,
		PLen:     byte(pLen),
		Password: password,
	}, nil
}

func (p *UserPassAuthReq) ToBytes() []byte {
	ret := []byte{p.Ver, p.ULen}
	ret = append(ret, p.UserName...)
	ret = append(ret, p.PLen)
	ret = append(ret, p.Password...)
	return ret
}

type UserPassAuthReply struct {
	Ver    byte
	Status byte
}

func (p *UserPassAuthReply) ToBytes() []byte {
	return []byte{p.Ver, p.Status}
}

func NewUserPassAuthReply(status byte) *UserPassAuthReply {
	return &UserPassAuthReply{
		Ver:    VerAuthUserPass,
		Status: status,
	}
}

func NewUserPassAuthReplyFrom(r io.Reader) (*UserPassAuthReply, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return &UserPassAuthReply{
		Ver:    b[0],
		Status: b[1],
	}, nil
}

type Request struct {
	Ver     byte
	Cmd     byte
	Rsv     byte //0x00
	Atyp    byte
	DstAddr []byte
	DstPort []byte //2 bytes
}

func NewRequestFrom(r io.Reader) (*Request, error) {
	b := []byte{0, 0, 0}
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}

	addrByte, err := NewAddrByteFrom(r)
	if err != nil {
		return nil, err
	}

	aType, addr, port := addrByte.Split()

	return &Request{
		Ver:     b[0],
		Cmd:     b[1],
		Rsv:     b[2],
		Atyp:    aType,
		DstAddr: addr,
		DstPort: port,
	}, nil
}

func NewRequest(cmd byte, addrByte AddrByte) *Request {
	aType, addr, port := addrByte.Split()
	return &Request{
		Ver:     VerSocks5,
		Cmd:     cmd,
		Rsv:     0,
		Atyp:    aType,
		DstAddr: addr,
		DstPort: port,
	}
}

func (p *Request) Address() string {
	var bAddr []byte
	bAddr = append(bAddr, p.Atyp)
	bAddr = append(bAddr, p.DstAddr...)
	bAddr = append(bAddr, p.DstPort...)
	return AddrByte(bAddr).String()
}

func (p *Request) ToBytes() []byte {
	ret := []byte{p.Ver, p.Cmd, p.Rsv, p.Atyp}
	ret = append(ret, p.DstAddr...)
	ret = append(ret, p.DstPort...)
	return ret
}

type Reply struct {
	Ver     byte
	Rep     byte
	Rsv     byte
	Atyp    byte
	BndAddr []byte
	BndPort []byte //2 bytes
}

func (p *Reply) Address() string {
	var bAddr []byte
	bAddr = append(bAddr, p.Atyp)
	bAddr = append(bAddr, p.BndAddr...)
	bAddr = append(bAddr, p.BndPort...)
	return AddrByte(bAddr).String()
}

func (p *Reply) ToBytes() []byte {
	ret := []byte{p.Ver, p.Rep, p.Rsv, p.Atyp}
	ret = append(ret, p.BndAddr...)
	ret = append(ret, p.BndPort...)
	return ret
}

func NewReply(rep byte, addrByte AddrByte) *Reply {
	aType, addr, port := addrByte.Split()
	return &Reply{
		Ver:     VerSocks5,
		Rep:     rep,
		Rsv:     0,
		Atyp:    aType,
		BndAddr: addr,
		BndPort: port,
	}
}

func NewReplyFrom(r io.Reader) (*Reply, error) {
	b := []byte{0, 0, 0}
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}

	addrByte, err := NewAddrByteFrom(r)
	if err != nil {
		return nil, err
	}

	aType, addr, port := addrByte.Split()

	return &Reply{
		Ver:     b[0],
		Rep:     b[1],
		Rsv:     b[2],
		Atyp:    aType,
		BndAddr: addr,
		BndPort: port,
	}, nil
}

type UDPDatagram struct {
	Rsv     []byte //0x00,0x00
	Frag    byte
	AType   byte
	DstAddr []byte
	DstPort []byte
	Data    []byte
}

func (p *UDPDatagram) ToBytes() []byte {
	b := []byte{}
	b = append(b, p.Rsv...)
	b = append(b, p.Frag)
	b = append(b, p.AType)
	b = append(b, p.DstAddr...)
	b = append(b, p.DstPort...)
	b = append(b, p.Data...)
	return b
}

func (p *UDPDatagram) Address() string {
	var bAddr []byte
	bAddr = append(bAddr, p.AType)
	bAddr = append(bAddr, p.DstAddr...)
	bAddr = append(bAddr, p.DstPort...)
	return AddrByte(bAddr).String()
}

func NewUDPDatagram(addrByte AddrByte, data []byte) *UDPDatagram {
	atype, addr, port := addrByte.Split()
	return &UDPDatagram{
		Rsv:     []byte{0, 0},
		Frag:    0,
		AType:   atype,
		DstAddr: addr,
		DstPort: port,
		Data:    data,
	}
}

func NewUDPDatagramFromBytes(b []byte) (*UDPDatagram, error) {
	if len(b) < 4 {
		return nil, ErrBadRequest
	}

	bAddr, err := NewAddrByteFromByte(b[3:])
	if err != nil {
		return nil, err
	}

	data := b[3+len(bAddr):]
	return NewUDPDatagram(bAddr, data), nil
}
