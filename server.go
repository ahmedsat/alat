package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"

	"github.com/ahmedsat/alat/alat"
)

type Server struct {
}

func (s *Server) Close(code int, reply *string) error {
	fmt.Println("Server is closing")
	os.Exit(code)
	return nil
}

func (s *Server) Run() {
	wc := alat.NewWindowCreator()
	rpc.Register(wc)
	rpc.Register(&Server{})
	listener, err := net.Listen("tcp", ":8080") // Listen on port 8080
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080...")
	go rpc.Accept(listener) // Accept incoming RPC connections

	wc.Show()
}
