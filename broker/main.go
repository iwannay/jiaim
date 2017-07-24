package main

var globalHub *hub
var globalConfig *config

func main() {
	globalConfig = newConfig()
	globalHub = newhub()
	go listentAndServerTcp()
	listenAndServer()
}
