package main

import (
	"os"

	"go_proxy/utils/conn"
	"go_proxy/utils/log"
)

func main() {
	err := LoadConf("./server.ini")
	if err != nil {
		os.Exit(-1)
	}

	log.InitLog(LogWay, LogFile, LogLevel)

	l, err := conn.Listen(BindAddr, BindPort)
	if err != nil {
		log.Error("Create listener error, %v", err)
		os.Exit(-1)
	}

	log.Info("Start server success")
	ProcessControlConn(l)
}
