package helloworld

import (
	"app/logic/router"
	"app/pkg/proto"
)

func Hello(msg *proto.Msg) *proto.Msg {
	return &proto.Msg{
		Body: []byte(`{"hello":"world"}`),
	}
}

func init() {
	router.Register("hello world", Hello)
}
