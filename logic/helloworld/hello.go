package helloworld

import (
	"jiaim/logic/router"
	"jiaim/pkg/proto"
)

func Hello(msg *proto.Msg) *proto.Msg {
	return &proto.Msg{
		Body: []byte(`{"hello":"world"}`),
	}
}

func init() {
	router.Register("hello world", Hello)
}
