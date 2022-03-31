package models

import (
	"encoding/json"

	"go_proxy/utils/conn"
	"go_proxy/utils/log"
)

type ProxyClient struct {
	Name      string
	Passwd    string
	LocalHost string
	LocalPort int64
}

func (p *ProxyClient) GetLocalConn() (c *conn.Conn, err error) {
	c = &conn.Conn{}
	err = c.ConnectServer(p.LocalHost, p.LocalPort)
	if err != nil {
		log.Error("ProxyName [%s], connect to %v:%v error, %v", p.Name, p.LocalHost, p.LocalPort, err)
	}
	return
}

func (p *ProxyClient) GetRemoteConn(addr string, port int64) (c *conn.Conn, err error) {
	c = &conn.Conn{}
	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	err = c.ConnectServer(addr, port)
	if err != nil {
		log.Error("ProxyName [%s], connect to server [%s:%d] error, %v", p.Name, addr, port, err)
		return
	}

	req := &ClientCtlReq{
		Type:      WorkConn,
		ProxyName: p.Name,
		Passwd:    p.Passwd,
	}

	buf, _ := json.Marshal(req)
	err = c.Write(string(buf) + "\n")
	if err != nil {
		log.Error("ProxyName [%s], write to server error, %v", p.Name, err)
		return
	}

	err = nil
	return
}

func (p *ProxyClient) StartTunnel(serverAddr string, serverPort int64) (err error) {
	localConn, err := p.GetLocalConn()
	if err != nil {
		return err
	}
	remoteConn, err := p.GetRemoteConn(serverAddr, serverPort)
	if err != nil {
		return err
	}

	log.Debug("Join two conns, (l[%s] r[%s]) (l[%s] r[%s])", localConn.GetLocalAddr(), localConn.GetRemoteAddr(),
		remoteConn.GetLocalAddr(), remoteConn.GetRemoteAddr())
	go conn.Join(localConn, remoteConn)
	return nil
}
