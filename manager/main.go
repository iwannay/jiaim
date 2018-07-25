package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
)

var (
	AppVersion = "0.0.1"
	AppName    = "im"
	Debug      bool
	StaticDir  = "/static"
	TplDir     = "template"
	TplExt     = ".html"

	JWTSigningKey   = "DFJLSDKJFIHFLFF234234"
	ManagerUser     = "admin"
	ManagerPasswd   = "123456"
	TokenCookieName = "token"
	TokenExpires    = 86400
)

func main() {
	Debug = true
	app := iris.New()
	if Debug {
		app.Logger().SetLevel("debug")
		app.Use(logger.New())
	}

	app.Use(recover.New())
	html := iris.HTML(TplDir, TplExt)
	html.Layout("layouts/layout.html")
	html.Reload(true)
	app.RegisterView(html)
	router(app)
	app.Run(iris.Addr(":9998"))
}
