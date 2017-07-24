package main

import "log"
import "encoding/json"

type hub struct {
	sessions map[*session]bool
	receiver chan *msg
	register chan *session

	unregister chan *session
}

func newhub() *hub {
	h := &hub{
		receiver:   make(chan *msg, 100),
		sessions:   make(map[*session]bool),
		register:   make(chan *session),
		unregister: make(chan *session),
	}
	go h.run()
	return h
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.sessions[client] = true
			// 调用route 更新client 状态
		case client := <-h.unregister:
			if _, ok := h.sessions[client]; ok {
				delete(h.sessions, client)
				// 调用route 更新client 状态
			}

		case msg := <-h.receiver:
			byts, _ := json.Marshal(*msg)
			log.Println("receiver", string(byts))
			// TODO 调用chat server 发布信息

		}
	}
}
