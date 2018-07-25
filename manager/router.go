package main

import (
	"jiaim/manager/handle"
	"jiaim/pkg/file"
	"path/filepath"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"

	"runtime"

	"net/http"
	"net/url"

	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

func notFound(ctx iris.Context) {
	ctx.View("public/404.html")
}

func catchError(ctx iris.Context) {
	ctx.ViewData("error", "服务不可用")
	ctx.View("public/error.html")
}

func router(app *iris.Application) {
	currDir, _ := file.GetCurrentDirectory()
	app.StaticWeb(StaticDir, filepath.Join(currDir, StaticDir))

	app.OnAnyErrorCode(catchError)
	app.OnErrorCode(iris.StatusNotFound, notFound)

	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSigningKey), nil
		},
		Extractor: func(ctx iris.Context) (string, error) {
			token, err := url.QueryUnescape(ctx.GetCookie(TokenCookieName))
			return token, err
		},

		ErrorHandler: func(ctx iris.Context, data string) {
			app.Logger().Error("jwt 认证失败", data)

			if ctx.RequestPath(true) != "/login" {
				ctx.Redirect("/login", http.StatusFound)
				return
			}

			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	app.Use(func(ctx iris.Context) {
		path := ctx.Request().URL.Path
		ctx.ViewData("action", filepath.Base(path))
		ctx.ViewData("title", AppName)
		ctx.ViewData("goVersion", runtime.Version())
		ctx.ViewData("appVersion", AppVersion)
		ctx.ViewData("appName", AppName)
		ctx.ViewData("requestPath", ctx.Request().URL.Path)
		ctx.ViewData("staticDir", StaticDir)
		ctx.Next()
	})

	app.Use(jwtHandler.Serve, func(ctx iris.Context) {
		token, ok := ctx.Values().Get("jwt").(*jwt.Token)
		if ok {
			ctx.ViewData("user", token.Claims)
		}
		ctx.Next()
	})

	app.Get("/", handle.Index)
	app.Any("/login", handle.Login)
	app.Get("/logout", handle.Logout)
	app.Post("/ajax/group/historyMsg", handle.GroupHistoryMsg)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", handle.Stat)
		debug.Get("/pprof/", handle.IndexDebug)
		debug.Get("/freemem", handle.FreeMem)

	}

}
