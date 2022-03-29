package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type HandleFunc struct {
	pattern    string
	handleFunc http.Handler
}

type Server struct {
	ctx        context.Context
	host       string
	port       string
	handleFunc []HandleFunc

	httpServer http.Server
}

func (s *Server) Start() (context.Context, error) {
	s.registerHandleFunc()
	ctx, cancel := context.WithCancel(s.ctx)
	s.httpServer.Addr = fmt.Sprintf("%s:%s", s.host, s.port)

	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()

	go func() {
		fmt.Printf("Server start at: %s Press any key to stop.\n", s.httpServer.Addr)
		var input string
		_, _ = fmt.Scanln(&input)
		_, _ = s.Stop(ctx)
		cancel()
	}()
	return ctx, nil
}

func (s *Server) Stop(ctx context.Context) (context.Context, error) {
	err := s.httpServer.Shutdown(ctx)
	return ctx, err
}

func (s *Server) registerHandleFunc() {
	for _, handleFunc := range s.handleFunc {
		http.Handle(handleFunc.pattern, handleFunc.handleFunc)
	}
}

func NewServer(host, port string, handleFunc []HandleFunc) *Server {
	return &Server{
		context.Background(),
		host,
		port,
		handleFunc,
		http.Server{},
	}
}
