package handle

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
)

var (
	JWTSigningKey   = "DFJLSDKJFIHFLFF234234"
	ManagerUser     = "admin"
	ManagerPasswd   = "123456"
	TokenCookieName = "token"
	TokenExpires    = 86400
)

func Login(ctx iris.Context) {
	var r = ctx.Request()
	if r.Method == http.MethodPost {

		u := r.FormValue("username")
		pwd := r.FormValue("passwd")
		remb := r.FormValue("remember")

		if u == ManagerUser && pwd == ManagerPasswd {

			clientFeature := ctx.RemoteAddr() + "-" + ctx.Request().Header.Get("User-Agent")

			clientSign := fmt.Sprintf("%x", md5.Sum([]byte(clientFeature)))
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"user":       u,
				"clientSign": clientSign,
			}).SignedString([]byte(JWTSigningKey))

			if err != nil {
				ctx.ViewData("error", fmt.Sprint("无法生成访问凭证:", err))
				ctx.View("public/error.html")
				return
			}
			if remb == "on" {
				ctx.SetCookieKV(TokenCookieName, url.QueryEscape(token), iris.CookiePath("/"),
					iris.CookieExpires(time.Duration(TokenExpires)*time.Second),
					iris.CookieHTTPOnly(true))

			} else {
				ctx.SetCookieKV(TokenCookieName, url.QueryEscape(token))
			}

			ctx.Redirect("/", http.StatusFound)
			return
		}

		ctx.ViewData("error", "auth failed")
		ctx.View("public/error.html")
		return
	}

	_, ok := ctx.Values().Get("jwt").(*jwt.Token)
	if ok {
		ctx.Redirect("/", http.StatusFound)
	}
	ctx.View("login.html")

}

func Logout(ctx iris.Context) {
	ctx.RemoveCookie(TokenCookieName)
	ctx.Redirect("/login", http.StatusFound)

}
