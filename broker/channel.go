package main

import (
	"app/pkg/proto"
)

// TODO 这里后面和session做结合
type Channel struct {
	Group      *Group
	msgChan    chan *proto.Msg
	connectNum int
	Cache      Ring
	Next       *Channel
	Prev       *Channel
}

func NewChannel(cacheSize, msgChanSize int) *Channel {
	c := &Channel{}
	c.Cache.init(uint64(cacheSize))
	c.msgChan = make(chan *proto.Msg, msgChanSize)
	return c
}

func (c *Channel) Push(msg *proto.Msg) (err error) {
	// 这里用select主要是为了防卡死
	select {
	case c.msgChan <- msg:
	default:
	}
	return
}

func (c *Channel) Ready() *proto.Msg {
	return <-c.msgChan
}

func (c *Channel) Signal() {
	c.msgChan <- proto.MsgReady
}

func (c *Channel) Close() {
	c.msgChan <- proto.MsgFinish
}
