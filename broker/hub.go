package main

import (
	"app/logic"
	"app/pkg/hash/cityhash"
	"app/pkg/proto"
	"encoding/json"
	"log"
)

type hub struct {
	Buckets    []*Bucket
	bucketIdx  uint32
	receiver   chan *proto.Msg
	register   chan *session
	unregister chan *session
}

func newhub(buckets []*Bucket) *hub {
	h := &hub{
		receiver:   make(chan *proto.Msg, 100),
		register:   make(chan *session),
		Buckets:    buckets,
		bucketIdx:  uint32(len(buckets)),
		unregister: make(chan *session),
	}

	go h.run()
	return h
}

func (h *hub) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % h.bucketIdx
	if Debug {
		log.Printf("\"%s\" hit channel bucket index: %d use cityhash \n", subKey, idx)
	}
	return h.Buckets[idx]
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			if Debug {
				log.Printf("login %+v\n", client)
			}
			// TODO 用户登录 ，更新用户信息
		case client := <-h.unregister:
			if Debug {
				log.Printf("logout %+v\n", client)
			}
			// TODO 用户退出，更新用户信息

		case msg := <-h.receiver:
			// TDOO 这里的逻辑暂时应该不用
			byts, _ := json.Marshal(*msg)
			logic.Handle(msg)
			if Debug {
				log.Println("receiver", string(byts))
			}
			// TODO 调用chat server 发布信息
		}
	}
}
