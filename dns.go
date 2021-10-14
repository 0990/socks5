package socks5

import (
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

func PrintDNS(title string, data []byte) {
	var a dns.Msg
	err := a.Unpack(data)
	if err != nil {
		return
	}

	logrus.Info(title)
	logrus.Info(a.String())
}
