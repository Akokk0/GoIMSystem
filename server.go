package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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
	user := CreateNewUser(conn, s)

	/*// 用户上线广播通知
	s.mapLock.Lock() // 加锁，map是线程不安全的数据类型
	// 将新用户添加到UserMap中
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock() // 添加完数据，解锁

	// 向全服用户广播上线通知
	s.BroadCast(user, "已上线")*/

	// 用户上线
	user.Online()

	// 监听用户是否活跃的channel
	isAlive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		// 定义用户输入消息的接收容器
		buff := make([]byte, 4096)
		for {
			// 读取消息
			n, err := conn.Read(buff)
			// 用户下线
			if n == 0 {
				// s.BroadCast(user, "已下线")
				// 用户下线
				user.Offline()
				return
			}
			// 未成功读取
			if err != nil && err != io.EOF {
				fmt.Println("Conn read err:", err)
			}
			// 提取用户消息
			msg := string(buff[:n-1])
			// 判断用户输入的是指令还是消息
			if user.MsgType(msg) {
				continue
			}
			// 将得到的消息进行广播
			s.BroadCast(user, msg)
			// 只要进入这里面就代表用户活跃
			isAlive <- true

		}
	}()

	// 用户活跃检测
	for {
		select {
		case <-isAlive:
			// 说明当前用户是活跃的，需要更新定时器
			// 不做任何事情，为了激活select，更新定时器
		// 定时器记录用户 xx S不发言视作不活跃被强踢
		case <-time.After(time.Second * 300):
			// 向用户发送消息
			user.SendMsg("因您长时间未活动，您已被踢出该服务器！")
			// 销毁用户使用的资源
			close(user.C)
			// 关闭连接
			conn.Close()
			// 退出当前Handler
			runtime.Goexit()

		}
	}
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
