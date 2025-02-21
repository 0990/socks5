package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/0990/socks5"
	"log"
	"net"
)

var addr = flag.String("addr", "127.0.0.1:1080", "addr")

func main() {
	flag.Parse()
	fmt.Println(*addr)
	sc := socks5.NewSocks5Client(socks5.ClientCfg{
		ServerAddr: *addr,
		UserName:   "",
		Password:   "",
		UDPTimout:  60,
	})
	conn, err := sc.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Fatal(err)
	}

	//one dns query packet
	b, err := hex.DecodeString("0001010000010000000000000a74787468696e6b696e6703636f6d0000010001")
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write(b); err != nil {
		log.Fatal(err)
	}

	b = make([]byte, 2048)
	n, err := conn.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	b = b[:n]
	b = b[len(b)-4:]
	fmt.Println(net.IPv4(b[0], b[1], b[2], b[3]))
}
