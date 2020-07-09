package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type Server struct {
	ListenPort int32
	UserName   string
	Password   string
	UDPTimout  int
	TCPTimeout int
	LogLevel   string
}

func ReadServerCfg(path string) (*Server, error) {
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

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := Server{}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func CreateServerCfg(path string) error {
	cfg := Server{
		ListenPort: 1080,
		UserName:   "",
		Password:   "",
		UDPTimout:  60,
		TCPTimeout: 60,
		LogLevel:   "debug",
	}
	c, _ := json.MarshalIndent(cfg, "", "    ")
	return ioutil.WriteFile(path, c, 0644)
}
