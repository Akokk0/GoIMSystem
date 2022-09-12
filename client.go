package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string   // 服务器IP
	ServerPort int      // 服务器Port
	Name       string   // 用户名
	conn       net.Conn // 连接具柄
	flag       int      // 用户使用的模式
}

func (c *Client) Menu() bool {
	// 模式判断数值
	var flag int
	// 打印提示
	fmt.Println("1、公聊模式")
	fmt.Println("2、私聊模式")
	fmt.Println("3、更新用户名")
	fmt.Println("0、退出")
	// 用户输入
	fmt.Scanln(&flag)
	// 判断用户输入的值是否合法
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println("请输入合法数值！")
		return false
	}
}

func (c *Client) Run() {
	for c.flag != 0 {
		// 只要返回值为false则一直循环
		for c.Menu() == false {
		}
		// 根据当前选择执行对应模式
		switch c.flag {
		case 1:
			c.publicChat()
		case 2:
			fmt.Println("私聊模式...")
		case 3:
			c.updateName()
		}
	}
}

func (c *Client) updateName() bool {
	// 提示用户输入用户名
	fmt.Println("请输入新用户名！")
	// 用户输入
	fmt.Scanln(&c.Name)
	// 发送的消息
	sendMsg := "/rename " + c.Name + "\n"
	// 获取发送消息的返回值
	_, err := c.conn.Write([]byte(sendMsg))
	// 判断发送是否失败
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}
	// 返回发送成功flag
	return true
}

// 公聊模式
func (c *Client) publicChat() {
	var chatMsg string
	fmt.Println("请输入聊天内容，输入exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" && len(chatMsg) != 0 {
		sendMsg := chatMsg + "\n"
		_, err := c.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn.Write error:", err)
			break
		}
		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}
}

// 回调处理函数，处理服务器返回的信息
func (c *Client) DealResponse() {
	// 简写
	io.Copy(os.Stdout, c.conn)
	// 原来的写法
	/*buff := make([]byte, 4096)
	c.conn.Read(buff)
	fmt.Println(buff)*/
}

// 创建一个新的客户端
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

// 保存服务器IP和端口
var serverIP string
var serverPort int

// ./client -ip 127.0.0.1 -p 8080
func init() {
	// 要解析命令行，要使用flag库
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址(默认值为127.0.0.1)")
	flag.IntVar(&serverPort, "p", 8080, "设置服务器端口(默认值为8080)")
	// 命令行解析
	flag.Parse()
}

func main() {
	// 创建新连接
	client := NewClient(serverIP, serverPort)
	// 判断是否连接成功
	if client == nil {
		fmt.Println("服务器连接失败")
		return
	}
	// 连接成功返回成功消息
	fmt.Println(">>>>>>>>>> 服务器连接成功！")
	// 单独开一个goroutine监听服务器返回的消息
	go client.DealResponse()
	// 执行客户端操作
	client.Run()
}
