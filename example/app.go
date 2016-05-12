package main

import (
	"fmt"
	"github.com/restgo/restgo"
	"github.com/restgo/session"
	"github.com/restgo/session-mongo"
	"github.com/valyala/fasthttp"
	"time"
)

func main() {

	router := restgo.NewRouter()

	sessionOpts := `{
		"Secret"     :"secret",
		"Secure"     :false,
		"Path"       :"/",
		"HttpOnly"   :true,
		"CookieName" :"cookie-session",
		"MaxAge"     : 86400,
		"EncyptCookie": false
	}`

	mongoOpts := `{
		"Hosts"     :"localhost",
		"Database"  :"test",
		"Collection":"sessions",
		"Username"  :"",
		"Password"  :""
	}`

	router.Use("/", session.NewSessionManager(session_mongo.NewMongoSessionStore(mongoOpts), sessionOpts))

	router.GET("/about", func(ctx *fasthttp.RequestCtx, next restgo.Next) {
		s := ctx.UserValue("session")
		session, _ := s.(*session.Session)
		if _, ok := session.Values["time"]; ok {
			fmt.Println(session.Values["time"])
		} else {
			session.Values["time"] = time.Now().Format("2006-01-02 15:04:05")
		}

		restgo.ServeTEXT(ctx, "About", 200)
	})

	fasthttp.ListenAndServe(":8080", router.FastHttpHandler)
}
