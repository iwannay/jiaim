package main

var globalHub *hub
var globalConfig *config

func main() {
	globalConfig = newConfig()

	globalHub = newhub()
	// 监听tcp连接
	go listentAndServerTCP()
	// 监听websocket连接
	listenAndServerWs()
}
