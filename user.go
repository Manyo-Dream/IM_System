package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

// 用户上线功能
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "已上线")
}

// 用户下线功能
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "已下线")
}

// 发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// 处理用户消息
func (u *User) DoMessage(msg string) {
	if msg == "who" {
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "\n"
			u.SendMsg(onlineMsg)
		}
		//修改用户名
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名已被使用\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("您已修改用户名" + newName + "\n")
		}
		//发送消息
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 1 获取用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式错误；消息格式：to|张三|消息内容\n")
			return
		}

		// 2 根据用户名获取User对象
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户不存在/未上线\n")
			return
		}

		// 3 将消息通过对方的User对象发送过去
		context := strings.Split(msg, "|")[2]
		if context == "" {
			u.SendMsg("消息不能为空；消息格式：to|张三|消息内容\n")
			return
		}
		remoteUser.SendMsg(u.Name + "：" + context + "\n")
	} else {
		u.server.BroadCast(u, msg)
	}
}

// 监听当前User Channel的方法，一旦有消息，就发送给客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
