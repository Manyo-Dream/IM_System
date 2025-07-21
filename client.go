package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int

	Name string
	conn net.Conn
	flag int
}

func NewClient(serverIP string, serverPort int) *Client {
	// 1 创建客户端对象
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       999,
	}

	// 2 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = conn

	// 3 返回客户端对象
	return client
}

// 处理服务器端返回的消息
func (c *Client) DealResponse() {
	//一旦client.conn有数据，就直接Copy到stdout标准输出中，永久阻塞监听
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>> 请输入合法范围内的数字")
		return false
	}
}

// 更新用户名
func (c *Client) UpdateName() bool {
	fmt.Println(">>>> 请输入用户名:")
	fmt.Scanln(&c.Name)

	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 搜索用户
func (c *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

// 私聊模式
func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.SelectUsers()
	fmt.Println("请选择聊天对象,exit退出.")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("请输入聊天内容,exit退出.")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) > 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("请输入聊天内容,exit退出.")
			fmt.Scanln(&chatMsg)
		}

		c.SelectUsers()
		fmt.Println("请选择聊天对象,exit退出.")
		fmt.Scanln(&remoteName)
	}
}

// 公聊模式
func (c *Client) PublicChat() {
	//提示用户输入信息
	var chatMsg string

	fmt.Println(">>>> 请输入聊天内容,exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送到服务器

		if chatMsg != "" {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>> 请输入聊天内容,exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}

		//根据不同的模式处理业务
		switch c.flag {
		case 1:
			fmt.Println("公聊模式")
			c.PublicChat()
		case 2:
			fmt.Println("私聊模式")
			c.PrivateChat()
		case 3:
			fmt.Println("更新用户名")
			c.UpdateName()
		}
	}
}

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器地址，默认127.0.0.1")
	flag.IntVar(&serverPort, "port", 8080, "设置服务器端口，默认8080")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>>> 连接服务器失败")
		return
	}

	// 单独开启一个goroutine处理服务器返回的消息
	go client.DealResponse()

	fmt.Println(">>>> 链接服务器成功")

	// 启动客户端业务
	client.Run()
}
