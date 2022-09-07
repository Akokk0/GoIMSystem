package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// Start server
func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("listen is err:", err)
	}

	// Close listener server
	defer listener.Close()

	// Accept
	for {
		conn, err := listener.Accept()
		fmt.Println("accept is connected")
		if err != nil {
			fmt.Println("accept is err:", err)
			// Do not run this code blow and keep listen
			continue
		}

		// Handler
		go s.Handler(conn)
	}
}

// Handler function
func (s *Server) Handler(conn net.Conn) {
	fmt.Println("connection is connected")
}

// Create a new server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}

	return server
}
