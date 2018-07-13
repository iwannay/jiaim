package logic

import (
	_ "app/logic/helloworld"
	"app/logic/router"
	"app/pkg/proto"
)

func Handle(msg *proto.Msg) (resp *proto.Msg, ok bool) {
	resp, ok = router.Handle(msg)
	if !ok {
		// 走http请求第三方服务
		return nil, true
	}

	return resp, ok
}
