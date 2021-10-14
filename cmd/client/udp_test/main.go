package main

import (
	"encoding/hex"
	"fmt"
	"github.com/0990/socks5"
	"net"
)

func main() {
	for i := 0; i < 10; i++ {
		err := doDNS()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func doDNS() error {
	sc := socks5.NewClient(socks5.ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  120,
		TCPTimeout: 0,
	})

	conn, err := sc.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return err
	}

	//one dns query packet
	b, err := hex.DecodeString("0001010000010000000000000a74787468696e6b696e6703636f6d0000010001")
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write(b); err != nil {
		return err
	}

	b = make([]byte, 2048)
	n, err := conn.Read(b)
	if err != nil {
		return err
	}
	b = b[:n]
	b = b[len(b)-4:]
	fmt.Println(net.IPv4(b[0], b[1], b[2], b[3]))
	return nil
}
