package main

import (
	"flag"
	"fmt"
	"github.com/0990/socks5/config"
	"github.com/0990/socks5/logconfig"
	"github.com/0990/socks5/socks5"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var confFile = flag.String("c", "ss5.json", "config file")

func main() {
	flag.Parse()

	cfg, err := config.ReadServerCfg(*confFile)
	if err != nil {
		logrus.Fatalln(err)
	}

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Fatalln(err)
	}
	logconfig.InitLogrus("ss5", 10, level)

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
