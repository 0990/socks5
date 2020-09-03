package socks5

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

var DefaultServerConfig = ServerCfg{
	ListenPort: 1080,
	UserName:   "",
	Password:   "",
	UDPTimout:  60,
	TCPTimeout: 60,
	UDPAddr:    "",
}

type ServerCfg struct {
	ListenPort int
	UserName   string
	Password   string
	UDPTimout  int
	TCPTimeout int
	UDPAddr    string
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
