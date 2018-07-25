// TODO 整体删除
// 全部由manager提供接口
package main

import (
	"encoding/json"
	"fmt"
	"jiaim/pkg/base"
	"jiaim/pkg/proto"
	"log"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"strings"
	"time"
)

const (
	HeaderContentType              = "Content-Type"
	CharsetUTF8                    = "text/plain; charset=utf-8"
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + CharsetUTF8
)

func RenderJSON(w http.ResponseWriter, v interface{}) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	bts, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.Write(bts)
	return nil
}

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
			b = Hub.Bucket(resp.Sid)
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
		hub:  Hub,
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

	Hub.history.Push(&message)
	if b := Hub.Bucket(sid); b != nil {
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
	start := time.Now()
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
	Hub.history.Push(&message)

	for _, v := range Hub.Buckets {
		v.Broadcast(&message)
	}

	sub := time.Now().Sub(start).Seconds()
	w.Write([]byte(fmt.Sprint(sub)))
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
	Hub.history.Push(&message)

	for _, v := range Hub.Buckets {
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
		Hub.history.Push(&msg)

		if b = Hub.Bucket(v); b != nil {
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

func Stat(w http.ResponseWriter, r *http.Request) {
	data := base.Stat.Collect()
	RenderJSON(w, data)
}

func SystemInfo(w http.ResponseWriter, r *http.Request) {
	info := base.Stat.SystemInfo()
	RenderJSON(w, info)
}

func Index(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}

func FreeMem(w http.ResponseWriter, r *http.Request) {
	debug.FreeOSMemory()
}

func RegisterHandle(mux *http.ServeMux) {
	// 广播
	mux.HandleFunc("/push/all", wrapHandler(PushAll))
	// 组播
	mux.HandleFunc("/push/group", wrapHandler(PushGroup))
	// 群发
	mux.HandleFunc("/push/many", wrapHandler(PushMany))
	// 单发
	mux.HandleFunc("/push", wrapHandler(PushMsg))

	mux.HandleFunc("/im", wrapHandler(serverWs))

	mux.HandleFunc("/debug/stat", wrapHandler(Stat))
	mux.HandleFunc("/debug/pprof/", wrapHandler(Index))

	mux.HandleFunc("/debug/freemem", wrapHandler(FreeMem))
	mux.HandleFunc("/debug/systeminfo", wrapHandler(SystemInfo))
}
