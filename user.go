package main

import "net"

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
