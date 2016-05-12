Mongo session store for [session](https://github.com/restgo/session) of [restgo](https://github.com/restgo/restgo)
[![GoDoc](https://godoc.org/github.com/restgo/session-mongo?status.svg)](https://godoc.org/github.com/restgo/session-mongo)

## Install
```
go get github.com/restgo/session-mongo
```

## Usage

```go
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

app.GET("/about", func(ctx *fasthttp.RequestCtx, next restgo.Next) {
    s := ctx.UserValue("session")
    session, _ := s.(*session.Session)
    if _, ok := session.Values["time"]; ok {
        fmt.Println(session.Values["time"])
    } else {
        session.Values["time"] = time.Now().Format("2006-01-02 15:04:05")
    }

    restgo.ServeTEXT(ctx, "About", 200)
})

app.Run(":8080")
```
