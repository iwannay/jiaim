package handle

import (
	"fmt"
	"jiaim/pkg/proto"
	"jiaim/pkg/rpc"

	"github.com/kataras/iris"
)

func Index(ctx iris.Context) {
	var groups proto.GroupsReply
	err := rpc.Call("localhost:9997", "Broker.Groups", &proto.EmptyArgs{}, &groups)
	if err != nil {
		ctx.ViewData("error", fmt.Sprint("Broker.Groups:", err))
		ctx.View("public/error.html")
		return
	}
	ctx.ViewData("groups", groups)
	ctx.View("index.html")
}

func GroupHistoryMsg(ctx iris.Context) {
	var msgList []proto.Msg
	gid := ctx.FormValue("gid")

	err := rpc.Call("localhost:9997", "Broker.GroupHistoryMsg", gid, &msgList)

	if err != nil {
		ctx.JSON(map[string]interface{}{
			"code": -1,
			"msg":  fmt.Sprintf("Broker.GroupHistoryMsg:%s", err),
		})
		return
	}

	ctx.JSON(map[string]interface{}{
		"code": 0,
		"data": msgList,
	})
}
