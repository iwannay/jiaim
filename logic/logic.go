package logic

import (
	_ "jiaim/logic/helloworld"
	"jiaim/logic/router"
	"jiaim/pkg/proto"
)

func Handle(msg *proto.Msg) (resp *proto.Msg, ok bool) {
	resp, ok = router.Handle(msg)
	if !ok {
		// 走http请求第三方服务
		return nil, true
	}

	return resp, ok
}
