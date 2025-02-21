package socks5

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

const DefaultListenPort = 1080
const DefaultTcpTimeout = 300
const DefaultUdpTimeout = 90
const DefaultLogLevel = "error"

var DefaultServerConfig = ServerCfg{
	ListenPort:      DefaultListenPort,
	UserName:        "",
	Password:        "",
	UDPTimout:       DefaultTcpTimeout,
	TCPTimeout:      DefaultUdpTimeout,
	UDPAdvertisedIP: "",
	LogLevel:        DefaultLogLevel,
}

type ServerCfg struct {
	ListenPort      int    //tcp,udp监听端口，仅当TCPListen或UDPListen无值时有效，监听地址为 0.0.0.0:ListenPort
	TCPListen       string //tcp监听地址
	UDPListen       string //udp监听地址
	UDPAdvertisedIP string //udp的广告IP地址,告诉客户端将UDP数据发往这个ip,默认值为udp监听的本地ip地址

	UserName   string
	Password   string
	UDPTimout  int
	TCPTimeout int
	LogLevel   string
}

func ReadOrCreateServerCfg(path string) (*ServerCfg, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := CreateServerCfg(path)
			if err != nil {
				return nil, err
			}
			logrus.WithField("path", path).Info("create config file")
		} else {
			return nil, err
		}
	}
	return ReadServerCfg(path)
}

func ReadServerCfg(path string) (*ServerCfg, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := ServerCfg{}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func CheckServerCfgDefault(cfg *ServerCfg) {
	if cfg.ListenPort <= 0 {
		cfg.ListenPort = DefaultListenPort
	}

	if cfg.TCPTimeout <= 0 {
		cfg.TCPTimeout = DefaultTcpTimeout
	}

	if cfg.UDPTimout <= 0 {
		cfg.UDPTimout = DefaultUdpTimeout
	}

	if len(cfg.LogLevel) == 0 {
		cfg.LogLevel = DefaultLogLevel
	}
}

func CreateServerCfg(path string) error {
	c, _ := json.MarshalIndent(DefaultServerConfig, "", "    ")
	return ioutil.WriteFile(path, c, 0644)
}

type ClientCfg struct {
	ServerAddr string
	UserName   string
	Password   string
	UDPTimout  int
	TCPTimeout int
}

func ReadClientCfg(path string) (*ServerCfg, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := CreateClientCfg(path)
			if err != nil {
				return nil, err
			}
			logrus.WithField("path", path).Info("create config file")
		} else {
			return nil, err
		}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := ServerCfg{}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func CreateClientCfg(path string) error {
	cfg := ClientCfg{
		ServerAddr: "127.0.0.1:1080",
		UserName:   "",
		Password:   "",
		UDPTimout:  60,
		TCPTimeout: 60,
	}
	c, _ := json.MarshalIndent(cfg, "", "    ")
	return ioutil.WriteFile(path, c, 0644)
}
