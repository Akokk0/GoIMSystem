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
	// 获取用户输入的指令
	command := strings.Split(cmd, " ")[0]
	// 判断是哪一个指令
	switch command {
	// 查询当前所有在线用户
	case "/who":
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			msg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.SendMsg(msg)
		}
		u.server.mapLock.Unlock()
	// 重命名
	case "/rename":
		name := strings.Split(cmd, " ")[1]
		// 判断之前是否存在该值
		if _, ok := u.server.OnlineMap[name]; ok {
			u.SendMsg("当前用户名被使用！\n")
		} else {
			// 不存在，可以使用该用户名
			u.server.mapLock.Lock()
			// 删除以前的值
			delete(u.server.OnlineMap, u.Name)
			// 把修改的用户名写入OnlineMap
			u.server.OnlineMap[name] = u
			u.server.mapLock.Unlock()
			// 修改自己的用户名
			u.Name = name
			// 告诉客服端已更新用户名
			u.SendMsg("您已更新用户名：" + name + "\n")

		}
	case "/to":
		// 把输入的内容分段
		msgList := strings.Split(cmd, " ")
		// 获取发送人
		who := msgList[1]
		// 获取发送内容
		content := msgList[2]
		// 查询发送人是否存在或在线
		receiver, ok := u.server.OnlineMap[who] // 获取发送人
		if !ok {
			// 发送人不存在或不在线
			u.SendMsg("该用户不存在或不在线\n")
			return
		}
		// 判断发送内容是否为空
		if content == "" {
			u.SendMsg("您发送的消息为空，请重新发送")
			return
		}
		// 发送消息
		receiver.SendMsg(u.Name + "对您说：" + content + "\n")
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
