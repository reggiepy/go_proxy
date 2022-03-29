package main

import (
	"go_proxy/server"
	"log"
)

func main() {
	ser := server.NewServer(
		"",
		"3000",
		make([]server.HandleFunc, 0),
	)
	ctx, err := ser.Start()
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
}
