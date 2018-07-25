package handle

import (
	"jiaim/pkg/base"
	"net/http/pprof"
	"runtime/debug"

	"github.com/kataras/iris"
)

func Stat(ctx iris.Context) {
	data := base.Stat.Collect()
	ctx.JSON(data)
}

func IndexDebug(ctx iris.Context) {
	pprof.Index(ctx.ResponseWriter(), ctx.Request())
}

func FreeMem(ctx iris.Context) {
	debug.FreeOSMemory()
}
