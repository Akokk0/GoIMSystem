package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

// 客服端结构
type Client struct {
	ServerIp   string   // 服务器IP
	ServerPort int      // 服务器Port
	Name       string   // 用户名
	conn       net.Conn // 连接具柄
	flag       int      // 用户使用的模式
}

// 菜单
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

// 客户端主要运行方法
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
			c.privateChat()
		case 3:
			c.updateName()
		}
	}
}

// 更新用户名
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
	// 定义聊天内容
	var chatMsg string
	// 打印提示
	fmt.Println("请输入聊天内容，输入exit退出")
	// 接收用户发送消息
	fmt.Scanln(&chatMsg)
	// 判断用户输入的是否是exit或输入的是空字符串
	for chatMsg != "exit" && len(chatMsg) != 0 {
		// 定义要发送的消息
		sendMsg := chatMsg + "\n"
		// 发送消息
		_, err := c.conn.Write([]byte(sendMsg))
		// 判断是否发送失败
		if err != nil {
			fmt.Println("conn.Write error:", err)
			break
		}
		// 重置消息
		chatMsg = ""
		// 再次接收用户要发送的消息
		fmt.Scanln(&chatMsg)
	}
}

// 查找在线用户
func (c *Client) onlineUsers() {
	// 定义指令
	sendMsg := "/who\n"
	// 发送消息
	_, err := c.conn.Write([]byte(sendMsg))
	// 判断消息是否发送失败
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return
	}
}

// 私聊模式
func (c *Client) privateChat() {
	// 定义要发送消息的用户
	var userName string
	// 定义要发送的消息
	var chatMsg string

	// 查询当前在线用户
	c.onlineUsers()
	// 打印提示消息
	fmt.Println("请输入您要发送的用户名，输入exit退出")
	// 接收用户输入的用户名
	fmt.Scanln(&userName)

	// 判断用户输入的用户名是否合法
	for userName != "exit" && len(userName) != 0 {
		// 打印提示消息
		fmt.Println("请输入您要发送的内容，输入exit退出")
		fmt.Scanln(&chatMsg)

		// 判断用户输入的消息内容是否合法
		for chatMsg != "exit" {
			// 判断用户是否输入空字符串
			if len(chatMsg) != 0 {
				// 拼接发送消息
				sendMsg := "/to " + userName + " " + chatMsg + "\n"
				// 发送消息
				_, err := c.conn.Write([]byte(sendMsg))
				// 判断消息是否发送成功
				if err != nil {
					fmt.Println("conn.Write error:", err)
					break
				}
			} else {
				fmt.Println("您什么也没输入，请重新输入要发送的内容！")
			}

			// 重置要发送的消息
			chatMsg = ""
			// 再次接收消息
			fmt.Scanln(&chatMsg)

		}
		// 查询当前在线用户
		c.onlineUsers()
		// 打印提示消息
		fmt.Println("请输入您要发送的用户名，输入exit退出")
		// 接收用户输入的用户名
		fmt.Scanln(&userName)
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
	// 创建新客服端
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

// 初始化方法 ./client -ip 127.0.0.1 -p 8080
func init() {
	// 要解析命令行，要使用flag库
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址(默认值为127.0.0.1)")
	flag.IntVar(&serverPort, "p", 8080, "设置服务器端口(默认值为8080)")
	// 命令行解析
	flag.Parse()
}

// 主函数
func main() {
	// 创建新连接
	client := NewClient(serverIP, serverPort)
	// 判断是否连接成功
	if client == nil {
		fmt.Println(">>>>>>>>>> 服务器连接失败")
		return
	}
	// 连接成功返回成功消息
	fmt.Println(">>>>>>>>>> 服务器连接成功！")
	// 单独开一个goroutine监听服务器返回的消息
	go client.DealResponse()
	// 执行客户端操作
	client.Run()
}
