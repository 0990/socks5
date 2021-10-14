package main

import (
	"flag"
	"fmt"
	"github.com/0990/socks5/logconfig"
	"os"
	"os/signal"
	"strconv"

	"github.com/0990/socks5"
	"github.com/sirupsen/logrus"
)

var confFile = flag.String("c", "ss5.json", "config file")

func main() {

	flag.Parse()

	cfg, err := socks5.ReadServerCfg(*confFile)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Infof("config file:%s not exist,use default config", *confFile)
			cfg = parseOSEnvCfg()
		} else {
			logrus.Fatal(err)
		}
	}

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Fatal(err)
	}

	logconfig.InitLogrus("ss5", 10, logLevel)

	logrus.Infof("Server Config:%+v", *cfg)

	server := socks5.NewServer(*cfg)
	err = server.Run()
	if err != nil {
		logrus.Fatalln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("quit,Got signal:", s)

}

//从环境变量中取配置，优先使用环境变量中的值
func parseOSEnvCfg() *socks5.ServerCfg {
	cfg := socks5.DefaultServerConfig

	username := os.Getenv("PROXY_USER")
	password := os.Getenv("PROXY_PASSWORD")
	udpTimeout := os.Getenv("PROXY_UDP_TIMEOUT")
	tcpTimeout := os.Getenv("PROXY_TCP_TIMEOUT")
	port := os.Getenv("PROXY_PORT")
	udpAddr := os.Getenv("PROXY_UDP_ADDR")

	if udpAddr != "" {
		cfg.UDPAddr = udpAddr
	}

	if username != "" {
		cfg.UserName = username
	}

	if password != "" {
		cfg.Password = password
	}

	if v, err := parseStr2Int(udpTimeout); err == nil {
		cfg.UDPTimout = v
	}

	if v, err := parseStr2Int(tcpTimeout); err == nil {
		cfg.TCPTimeout = v
	}

	if v, err := parseStr2Int(port); err == nil {
		cfg.ListenPort = v
	}

	return &cfg
}

func parseStr2Int(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	return int(v), err
}
