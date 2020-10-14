package socks5

import (
	"fmt"
	"github.com/miekg/dns"
)

func PrintDNS(title string, data []byte) {
	var a dns.Msg
	err := a.Unpack(data)
	if err != nil {
		return
	}

	fmt.Println(title)
	fmt.Println(a.String())
}
