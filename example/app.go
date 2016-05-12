package main

import (
	"fmt"
	"github.com/restgo/restgo"
	"github.com/restgo/session"
	"github.com/restgo/session-mongo"
	"time"
)

func main() {

	app := restgo.App()

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

	app.Use("/", session.NewSessionManager(session_mongo.NewMongoSessionStore(mongoOpts), sessionOpts))

	app.GET("/about", func(ctx *restgo.Context, next restgo.Next) {
		s := ctx.UserValue("session")
		session, _ := s.(*session.Session)
		if _, ok := session.Values["time"]; ok {
			fmt.Println(session.Values["time"])
		} else {
			session.Values["time"] = time.Now().Format("2006-01-02 15:04:05")
		}

		ctx.ServeText(200, "About")
	})

	app.Run(":8080")
}
