package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"go_proxy/models"
	"go_proxy/utils/conn"
	"go_proxy/utils/log"
)

func ProcessControlConn(l *conn.Listener) {
	for {
		c := l.GetConn()
		log.Debug("Get one new conn, %v", c.GetRemoteAddr())
		go controlWorker(c)
	}
}

// control connection from every client and server
func controlWorker(c *conn.Conn) {
	defer c.Close()
	// the first message is from client to server
	// if error, close connection
	res, err := c.ReadLine()
	if err != nil {
		log.Warn("Read error, %v", err)
		return
	}
	log.Debug("get: %s", res)

	clientCtlReq := &models.ClientCtlReq{}
	clientCtlRes := &models.ClientCtlRes{}
	if err := json.Unmarshal([]byte(res), &clientCtlReq); err != nil {
		log.Warn("Parse err: %v : %s", err, res)
		return
	}

	// check
	succ, msg, needRes := checkProxy(clientCtlReq, c)
	if !succ {
		clientCtlRes.Code = 1
		clientCtlRes.Msg = msg
	}

	if needRes {
		buf, _ := json.Marshal(clientCtlRes)
		err = c.Write(string(buf) + "\n")
		if err != nil {
			log.Warn("Write error, %v", err)
			time.Sleep(1 * time.Second)
			return
		}
	} else {
		// work conn, just return
		return
	}

	// others is from server to client
	server, ok := ProxyServers[clientCtlReq.ProxyName]
	if !ok {
		log.Warn("ProxyName [%s] is not exist", clientCtlReq.ProxyName)
		return
	}

	go readControlMsgFromClient(server, c)

	serverCtlReq := &models.ClientCtlReq{}
	serverCtlReq.Type = models.WorkConn
	for {
		_, isStop := server.WaitUserConn()
		if isStop {
			break
		}
		buf, _ := json.Marshal(serverCtlReq)
		err = c.Write(string(buf) + "\n")
		if err != nil {
			log.Warn("ProxyName [%s], write to client error, proxy exit", server.Name)
			server.Close()
			return
		}

		log.Debug("ProxyName [%s], write to client to add work conn success", server.Name)
	}

	log.Error("ProxyName [%s], I'm dead!", server.Name)
	return
}

func checkProxy(req *models.ClientCtlReq, c *conn.Conn) (succ bool, msg string, needRes bool) {
	succ = false
	needRes = true
	// check if proxy name exist
	server, ok := ProxyServers[req.ProxyName]
	if !ok {
		msg = fmt.Sprintf("ProxyName [%s] is not exist", req.ProxyName)
		log.Warn(msg)
		return
	}

	// check password
	if server.Passwd != "" && req.Passwd != server.Passwd {
		msg = fmt.Sprintf("ProxyName [%s], password is not correct", req.ProxyName)
		log.Warn(msg)
		return
	}

	// control conn
	if req.Type == models.ControlConn {
		if server.Status != models.Idle {
			msg = fmt.Sprintf("ProxyName [%s], already in use", req.ProxyName)
			log.Warn(msg)
			return
		}

		// start proxy and listen for user conn, no block
		err := server.Start()
		if err != nil {
			msg = fmt.Sprintf("ProxyName [%s], start proxy error: %v", req.ProxyName, err.Error())
			log.Warn(msg)
			return
		}

		log.Info("ProxyName [%s], start proxy success", req.ProxyName)
	} else if req.Type == models.WorkConn {
		// work conn
		needRes = false
		if server.Status != models.Working {
			log.Warn("ProxyName [%s], is not working when it gets one new work conn", req.ProxyName)
			return
		}

		server.CliConnChan <- c
	} else {
		msg = fmt.Sprintf("ProxyName [%s], type [%d] unsupport", req.ProxyName)
		log.Warn(msg)
		return
	}

	succ = true
	return
}

func readControlMsgFromClient(server *models.ProxyServer, c *conn.Conn) {
	isContinueRead := true
	f := func() {
		isContinueRead = false
		server.StopWaitUserConn()
	}
	timer := time.AfterFunc(time.Duration(HeartBeatTimeout)*time.Second, f)
	defer timer.Stop()

	for isContinueRead {
		content, err := c.ReadLine()
		//log.Debug("Receive msg from client! content:%s", content)
		if err != nil {
			if err == io.EOF {
				log.Warn("Server detect client[%s] is dead!", server.Name)
				server.StopWaitUserConn()
				break
			}
			log.Error("ProxyName [%s], read error:%s", server.Name, err.Error())
			continue
		}

		if content == "\r\n" {
			log.Debug("receive hearbeat:%s", content)
			timer.Reset(10 * time.Second)
		}
	}
}
