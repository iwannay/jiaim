package main

import (
	"jiaim/pkg/proto"
	"log"
	"net"
	"net/http"
)

func listenAndServer() {
	mux := http.NewServeMux()

	route(mux)

	log.Println("websocket: listen", ":10000")
	log.Fatal(http.ListenAndServe(":10000", mux))
}

func listentAndServerTcp() {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP(globalConfig.tcpConfig.ip), globalConfig.tcpConfig.port, ""})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Tcp: listen %s:%d", globalConfig.ip, globalConfig.port)
	for {
		conn, err := l.Accept()
		if err != nil {
			conn.Close()
			log.Println("Error:", err)
			break
		}
		go wrapHandleTcp(conn)
	}
}

func handleTcp(conn net.Conn) {
	var id string
	log.Printf("Tcp: accept %s->%s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	session := &session{
		id:      id,
		hub:     globalHub,
		tcpConn: conn,
		send:    make(chan proto.Msg, 256),
	}
	session.hub.register <- session
	go session.writeTcpPump()
	session.readTcpPump()
}

func wrapHandleTcp(conn net.Conn) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
		}
	}()
	handleTcp(conn)
}
