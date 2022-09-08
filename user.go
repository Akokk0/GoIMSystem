package main

import (
	"net"
	"strings"
)

// Define user struct
type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// Create a new user
func CreateNewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// Start user's listen message function
	go user.ListenMessage()

	return user
}

func (u *User) MsgType(msg string) bool {
	if strings.HasPrefix(msg, "/") {
		u.CMListener(msg)
		return true
	} else {
		return false
	}
}

// User CLI
func (u *User) CMListener(cmd string) {
	if cmd == "/who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			msg := "[" + user.Addr + "]" + user.Name + ":" + "在线..."
			u.SendMsg(msg)
		}
		u.server.mapLock.Unlock()
	}
}

func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// Listen message and send message to conn
func (u *User) ListenMessage() {
	// 监听自己的管道是否有新消息，有则写给客户端
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线
func (u *User) Online() {
	u.server.mapLock.Lock() // 加锁，map是线程不安全的数据类型
	// 将新用户添加到UserMap中
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock() // 添加完数据，解锁

	// 向全服用户广播上线通知
	u.server.BroadCast(u, "已上线")
}

// 用户下线
func (u *User) Offline() {
	// 加锁，map是线程不安全的数据类型
	u.server.mapLock.Lock()
	// 将用户从UserMap中移除
	delete(u.server.OnlineMap, u.Name)
	// 添加完数据，解锁
	u.server.mapLock.Unlock()

	// 向全服用户广播上线通知
	u.server.BroadCast(u, "已下线")
}
