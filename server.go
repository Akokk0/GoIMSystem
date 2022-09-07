package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	Message   chan string
	mapLock   sync.Mutex
}

// Create a new server
func CreateNewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
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
		// fmt.Println("accept is connected")
		if err != nil {
			fmt.Println("accept is err:", err)
			// Do not run this code blow and keep listen
			continue
		}

		// Handler
		go s.Handler(conn)

		// 开启消息监听
		go s.MessageListener()
	}
}

// Handler function
func (s *Server) Handler(conn net.Conn) {
	// fmt.Println("connection is connected")

	// 得到新用户
	user := CreateNewUser(conn)

	// 用户上线广播通知
	s.mapLock.Lock() // 加锁，map是线程不安全的数据类型
	// 将新用户添加到UserMap中
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock() // 添加完数据，解锁

	// 向全服用户广播上线通知
	s.BroadCast(user, "已上线")

	// 接收客户端发送的消息
	go func() {
		buff := make([]byte, 4096)
		for {
			// 读取消息
			n, err := conn.Read(buff)

			// 排除错误
			if n == 0 {
				s.BroadCast(user, "已下线")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn read err:", err)
			}

			// 提取用户消息
			msg := string(buff[:n-1])

			// 将得到的消息进行广播
			s.BroadCast(user, msg)

		}
	}()

	// 保持handler不死
	select {}

}

func (s *Server) BroadCast(user *User, msg string) {
	// 拼接要发送的字符串
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	// 将要发送的字符串添加到message管道
	s.Message <- sendMsg
}

func (s *Server) MessageListener() {
	for {
		msg := <-s.Message
		// 加锁
		s.mapLock.Lock()
		// 遍历得到所有在线用户
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		// 闭锁
		s.mapLock.Unlock()
	}
}
