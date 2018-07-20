package main

import (
	"encoding/json"
	"jiaim/pkg/http"
	"log"

	"github.com/gorilla/websocket"

	"fmt"
	"jiaim/pkg/proto"
	"time"
)

func authWebsocket(msg *proto.Msg, conn *websocket.Conn) (resp *proto.Auth, err error) {
	var (
		bts []byte
		sid string
	)
	err = conn.ReadJSON(msg)

	if err != nil {
		return
	}

	if Debug {
		log.Printf("auth %+v\n", msg)
		log.Println(string(msg.Body))
	}

	if Debug {
		if msg.Sid == "" {
			if Debug {
				log.Printf("empty sessionId:%s|%+v\n", sid, msg)
			}
			sid = fmt.Sprintf("%d", time.Now().Unix())
		} else {
			sid = msg.Sid
		}
		resp = &proto.Auth{
			Sid:  sid,
			Gids: []string{msg.Gid, "verystar"},
		}
	} else {
		bts, err = DefaultHttpClient.Post("xxx", http.ContentTypeJSON, false, map[string]interface{}{
			"session_id": msg.Sid,
		})
		if err != nil {
			json.Unmarshal(bts, msg)
		}
	}

	return

}
