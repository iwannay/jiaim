package main

import (
	"app/pkg/proto"
	"encoding/json"
	"log"
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
		log.Println(r.Host, r.RemoteAddr, r.RequestURI, r.UserAgent())
		fn(w, r)
	}
}

func serverWs(w http.ResponseWriter, r *http.Request) {

	var (
		ch   = NewChannel(100, 100)
		msg  *proto.Msg
		resp *proto.Auth
		err  error
		b    *Bucket
	)

	if r.Header.Get("Sec-Websocket-Version") == "" {
		log.Println("Error:", "Sec-Websocket-Version == ''")
		return
	}

	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	// 认证
	if msg, err = ch.Cache.Set(); err == nil {
		if resp, err = authWebsocket(msg, conn); err == nil {
			b = globalHub.Bucket(resp.SessionId)
			err = b.Put(resp.SessionId, resp.GroupId, ch)
		}
	}

	if err != nil {
		if Debug {
			log.Println("认证：handshark error", err)
		}
		conn.Close()
		return
	}

	session := &session{
		id:   resp.SessionId,
		hub:  globalHub,
		conn: conn,
		ch:   ch,
		b:    b,
		send: make(chan proto.Msg, 256),
	}
	session.hub.register <- session

	go session.writePump()
	session.readPump()
}

func httpPushMsg(w http.ResponseWriter, r *http.Request) {
	gId := r.FormValue("gid")
	body := r.FormValue("body")
	bts, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
	}
	var message = proto.Msg{
		MsgId: 100,
		Op:    proto.ClientSendMsg,
		Body:  bts,
	}

	if gId == "" {
		for _, v := range globalHub.Buckets {
			v.Broadcast(&message)
		}
	} else {
		for _, v := range globalHub.Buckets {
			v.BroadcastGroup(&proto.BoardcastGroupArg{
				GId: gId,
				M:   message,
			})
		}
	}

	w.Write([]byte("1"))
}

func route(mux *http.ServeMux) {
	mux.HandleFunc("/push", httpPushMsg)
	mux.HandleFunc("/im", wrapHandler(serverWs))
}
