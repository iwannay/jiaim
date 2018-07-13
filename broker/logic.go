package main

import (
	"app/pkg/http"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"

	"app/pkg/proto"
	"fmt"
	"time"
)

func authWebsocket(msg *proto.Msg, conn *websocket.Conn) (resp *proto.Auth, err error) {
	var bts []byte
	err = conn.ReadJSON(msg)
	if err != nil {
		return
	}

	if Debug {
		log.Printf("auth %+v\n", msg)
		log.Println(string(msg.Body))
	}

	if Debug {
		resp = &proto.Auth{
			SessionId: fmt.Sprintf("%x", time.Now().Unix()),
			UniqueId:  fmt.Sprintf("%x", time.Now().Unix()),
			GroupId:   "one-for-all",
		}
	} else {
		bts, err = DefaultHttpClient.Post("xxx", http.ContentTypeJSON, false, map[string]interface{}{
			"session_id": msg.SessionId,
		})
		if err != nil {
			json.Unmarshal(bts, msg)
		}
	}

	return

}
