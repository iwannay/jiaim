package main

import (
	"log"
	"net"
	"net/http"
	"runtime/debug"
)

func wrapHandler(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				http.Error(w, "server error!", http.StatusInternalServerError)
				log.Println("Error: panic in", e, string(debug.Stack()))

			}
		}()
		log.Println(r.Host, r.RemoteAddr)

		fn(w, r)
	}
}

func serverWs(w http.ResponseWriter, r *http.Request) {

	var uin string
	if r.Header.Get("Sec-Websocket-Version") == "" {
		log.Println("Error:", "Sec-Websocket-Version == ''")
		return
	}
	conn, err := upgrade.Upgrade(w, r, nil)
	defer conn.Close()

	if err != nil {
		log.Println("Error:", err)
		return
	}
	session := &session{
		uin:  uin,
		hub:  globalHub,
		conn: conn,
		send: make(chan msg, 256),
	}
	session.hub.register <- session
	go session.writePump()
	session.readPump()
}

func listenAndServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/sub", wrapHandler(serverWs))
	log.Fatal(http.ListenAndServe(":8080", mux))
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
	var uin string
	log.Printf("Tcp: accept %s->%s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	session := &session{
		uin:     uin,
		hub:     globalHub,
		tcpConn: conn,
		send:    make(chan msg, 256),
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
