package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// 定义server
type Server struct {
	IP   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息管道
	Message chan string
}

// 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听消息管道
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + ":" + msg
	s.Message <- sendMsg
}

// 处理用户
func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)
	user.Online()

	isLive := make(chan bool)
	defer close(isLive) // 确保退出时关闭 channel

	go func() {
		defer func() {
			user.Offline()
			conn.Close() // 确保连接关闭
		}()

		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 || err != nil {
				if err != io.EOF {
					fmt.Println("Conn Read err:", err)
				}
				return
			}

			msg := strings.TrimSpace(string(buf[:n])) // 安全处理消息
			user.DoMessage(msg)

			select {
			case isLive <- true: // 避免阻塞（如果主 goroutine 已退出）
			default:
			}
		}
	}()

	timeout := time.NewTimer(300 * time.Second)
	defer timeout.Stop() // 避免 Timer 泄漏

	for {
		select {
		case <-isLive:
			if !timeout.Stop() {
				<-timeout.C // 排空 channel
			}
			timeout.Reset(300 * time.Second) // 重置超时
		case <-timeout.C:
			user.SendMsg("你已掉线")
			return // defer 会处理资源释放
		}
	}
}

func (s *Server) Start() {
	//创建监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}

	//关闭监听
	defer listener.Close()

	//监听Message
	go s.ListenMessage()

	//具体操作
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		go s.Handler(conn)
	}
}
