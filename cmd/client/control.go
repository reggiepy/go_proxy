package main

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"go_proxy/models"
	"go_proxy/utils/conn"
	"go_proxy/utils/log"
)

var isHeartBeatContinue bool = true

func ControlProcess(cli *models.ProxyClient, wait *sync.WaitGroup) {
	defer wait.Done()
	c := loginServer(cli)
	if c == nil {
		log.Error("ProxyName [%s], connect to server failed!", cli.Name)
		return
	}
	defer c.Close()

	for {
		// ignore response content now
		_, err := c.ReadLine()
		if err == io.EOF {
			isHeartBeatContinue = false
			log.Debug("ProxyName [%s], server close this control conn", cli.Name)
			var sleepTime time.Duration = 1
			for {
				log.Debug("ProxyName [%s] try to reconnect to server [%s:%s]", cli.Name, ServerAddr, ServerPort)
				tmpConn := loginServer(cli)
				if tmpConn != nil {
					c.Close()
					c = tmpConn
					break
				}
				if sleepTime < 60 {
					sleepTime++
				}
				time.Sleep(sleepTime * time.Second)
			}
			continue
		} else if err != nil {
			log.Warn("ProxyName [%s], read from server error, %v", cli.Name, err)
			continue
		}

		err = cli.StartTunnel(ServerAddr, ServerPort)
		if err != nil {
			log.Error("StartTunnel error, %v", err)
		}
	}
}

func loginServer(cli *models.ProxyClient) (connect *conn.Conn) {
	c := &conn.Conn{}
	connect = nil
	for i := 0; i < 1; i++ {
		err := c.ConnectServer(ServerAddr, ServerPort)
		if err != nil {
			log.Error("ProxyName [%s], connect to server [%s:%d] error, %v", cli.Name, ServerAddr, ServerPort, err)
			break
		}

		req := &models.ClientCtlReq{
			Type:      models.ControlConn,
			ProxyName: cli.Name,
			Passwd:    cli.Passwd,
		}
		buf, _ := json.Marshal(req)
		err = c.Write(string(buf) + "\n")
		if err != nil {
			log.Error("ProxyName [%s], write to server error, %v", cli.Name, err)
			break
		}

		res, err := c.ReadLine()
		if err != nil {
			log.Error("ProxyName [%s], read from server error, %v", cli.Name, err)
			break
		}
		log.Debug("ProxyName [%s], read [%s]", cli.Name, res)

		clientCtlRes := &models.ClientCtlRes{}
		if err = json.Unmarshal([]byte(res), &clientCtlRes); err != nil {
			log.Error("ProxyName [%s], format server response error, %v", cli.Name, err)
			break
		}

		if clientCtlRes.Code != 0 {
			log.Error("ProxyName [%s], start proxy error, %s", cli.Name, clientCtlRes.Msg)
			break
		}
		connect = c
		go startHeartBeat(c)
		log.Debug("ProxyName [%s], connect to server[%s:%d] success!", cli.Name, ServerAddr, ServerPort)
	}
	if connect == nil {
		c.Close()
	}
	return
}

func startHeartBeat(conn *conn.Conn) {
	isHeartBeatContinue = true
	for {
		time.Sleep(time.Duration(HeartBeatInterval) * time.Second)
		if isHeartBeatContinue == true {
			err := conn.Write("\r\n")
			if err != nil {
				log.Error("send heart beat error, %v", err)
			}
		} else {
			break
		}
	}
}
