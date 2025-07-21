package main

// 启动服务器
func main() {
	server := NewServer("127.0.0.1", 8080)
	server.Start()
}
