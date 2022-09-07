package main

import "net"

// Define user struct
type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// Create a new user
func CreateNewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// Start user's listen message function
	go user.ListenMessage()

	return user
}

// Listen message and send message to conn
func (u *User) ListenMessage() {
	// 监听自己的管道是否有新消息，有则写给客户端
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
