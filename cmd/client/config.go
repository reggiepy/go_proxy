package main

import (
	"fmt"
	"strconv"

	"go_proxy/models"

	ini "github.com/vaughan0/go-ini"
)

// common config
var (
	ServerAddr        string = "0.0.0.0"
	ServerPort        int64  = 7000
	LogFile           string = "./client.log"
	LogLevel          string = "warn"
	LogWay            string = "file"
	Passwd            string = ""
	HeartBeatInterval int64  = 5
)

var ProxyClients map[string]*models.ProxyClient = make(map[string]*models.ProxyClient)

func LoadConf(confFile string) (err error) {
	var tmpStr string
	var ok bool

	conf, err := ini.LoadFile(confFile)
	if err != nil {
		return err
	}

	// common
	tmpStr, ok = conf.Get("common", "server_addr")
	if ok {
		ServerAddr = tmpStr
	}

	tmpStr, ok = conf.Get("common", "server_port")
	if ok {
		ServerPort, _ = strconv.ParseInt(tmpStr, 10, 64)
	}

	tmpStr, ok = conf.Get("common", "log_file")
	if ok {
		LogFile = tmpStr
	}

	tmpStr, ok = conf.Get("common", "log_level")
	if ok {
		LogLevel = tmpStr
	}

	tmpStr, ok = conf.Get("common", "log_way")
	if ok {
		LogWay = tmpStr
	}

	tmpStr, ok = conf.Get("common", "passwd")
	if ok {
		Passwd = tmpStr
	}
	tmpStr, ok = conf.Get("common", "heartbeat_interval")
	if ok {
		HeartBeatInterval, _ = strconv.ParseInt(tmpStr, 10, 64)
	}

	// servers
	for name, section := range conf {
		if name != "common" {
			proxyClient := &models.ProxyClient{}
			proxyClient.Name = name

			proxyClient.Passwd = Passwd

			proxyClient.LocalHost, ok = section["local_host"]
			if !ok {
				return fmt.Errorf("Parse ini file error: proxy [%s] no local_host found", proxyClient.Name)
			}

			portStr, ok := section["local_port"]
			if ok {
				proxyClient.LocalPort, err = strconv.ParseInt(portStr, 10, 64)
				if err != nil {
					return fmt.Errorf("Parse ini file error: proxy [%s] local_port error", proxyClient.Name)
				}
			} else {
				return fmt.Errorf("Parse ini file error: proxy [%s] local_port not found", proxyClient.Name)
			}

			ProxyClients[proxyClient.Name] = proxyClient
		}
	}

	if len(ProxyClients) == 0 {
		return fmt.Errorf("Parse ini file error: no proxy config found")
	}

	return nil
}
