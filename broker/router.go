// TODO 整体迁出移到logic层
package main

import (
	"app/pkg/proto"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
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
			b = globalHub.Bucket(resp.Sid)
			for _, v := range resp.Gids {
				err = b.Put(resp.Sid, v, ch)
			}

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
		id:   resp.Sid,
		hub:  globalHub,
		conn: conn,
		ch:   ch,
		b:    b,
		send: make(chan proto.Msg, 256),
	}
	session.hub.register <- session

	if session.parseMsg(msg) {
		ch.Cache.SetAdv()
		ch.Signal()
	}

	go session.writePump()
	session.readPump()
}

func PushMsg(w http.ResponseWriter, r *http.Request) {

	var ret = "0"
	body := r.FormValue("body")
	sid := r.FormValue("sid")

	bts, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
	}

	var message = proto.Msg{
		Sid:  sid,
		Op:   proto.ServerReplyMsg,
		Body: bts,
	}

	globalHub.history.Push(&message)
	if b := globalHub.Bucket(sid); b != nil {
		if chs := b.Channel(sid); len(chs) != 0 {
			for ch := range chs {
				if err := ch.Push(&message); err == nil {
					ret = "1"
				}
			}
		}
	}

	w.Write([]byte(ret))
}

func PushAll(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")

	bts, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		w.Write([]byte("0"))
		return
	}

	var message = proto.Msg{
		Gid:  proto.WholeChannel,
		Op:   proto.ServerReplyMsg,
		Body: bts,
	}

	// 一定要放在上面，Push操作会生成MsgId
	globalHub.history.Push(&message)

	for _, v := range globalHub.Buckets {
		v.Broadcast(&message)
	}

	w.Write([]byte("1"))
}

func PushGroup(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	gid := r.FormValue("gid")
	bts, err := json.Marshal(body)

	if err != nil {
		log.Println(err)
		w.Write([]byte("0"))
		return
	}

	message := proto.Msg{
		Gid:  gid,
		Op:   proto.ServerReplyMsg,
		Body: bts,
	}
	globalHub.history.Push(&message)

	for _, v := range globalHub.Buckets {
		v.BroadcastGroup(&proto.BoardcastGroupArg{
			GId: gid,
			M:   message,
		})
	}

	w.Write([]byte("1"))
}

func PushMany(w http.ResponseWriter, r *http.Request) {
	var (
		sids    []string
		body    string
		err     error
		b       *Bucket
		ch      *Channel
		i       int
		v       string
		message proto.Msg
		bts     []byte
	)
	sids = strings.Split(r.FormValue("sids"), ",")
	body = r.FormValue("body")
	bts, err = json.Marshal(body)
	if err != nil {
		log.Println(err)
		w.Write([]byte("0"))
		return
	}

	message = proto.Msg{
		Op:   proto.ServerReplyMsg,
		Body: bts,
	}

	for _, v = range sids {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if Debug {
			log.Println("push many", v)
		}

		msg := message
		msg.Sid = v
		globalHub.history.Push(&msg)

		if b = globalHub.Bucket(v); b != nil {
			if chs := b.Channel(v); len(chs) != 0 {
				for ch = range chs {
					if err := ch.Push(&msg); err == nil {
						i++
					}
				}

			}
		}
	}

	w.Write([]byte(fmt.Sprint(i)))
}

func route(mux *http.ServeMux) {
	// 广播
	mux.HandleFunc("/push/all", wrapHandler(PushAll))
	// 组播
	mux.HandleFunc("/push/group", wrapHandler(PushGroup))
	// 群发
	mux.HandleFunc("/push/many", wrapHandler(PushMany))
	// 单发
	mux.HandleFunc("/push", wrapHandler(PushMsg))

	mux.HandleFunc("/im", wrapHandler(serverWs))
}
