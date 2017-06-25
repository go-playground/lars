package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/lars"
)

// MyContext is a custom context
type MyContext struct {
	*lars.Ctx // a little dash of Duck Typing....
}

func (c *MyContext) String(code int, s string) {

	res := c.Response()

	res.Header().Set(lars.ContentType, lars.TextPlainCharsetUTF8)
	res.WriteHeader(code)
	res.Write([]byte(s))
}

func newContext(l *lars.LARS) lars.Context {
	return &MyContext{
		Ctx: lars.NewContext(l),
	}
}

func castCustomContext(c lars.Context, handler lars.Handler) {

	// could do it in all one statement, but in long form for readability
	h := handler.(func(*MyContext))
	ctx := c.(*MyContext)

	h(ctx)
}

func main() {

	l := lars.New()
	l.RegisterContext(newContext) // all gets cached in pools for you
	l.RegisterCustomHandler(func(*MyContext) {}, castCustomContext)
	l.Use(Logger)

	l.Get("/", Home)

	http.ListenAndServe(":3007", l.Serve())
}

// Home ...
func Home(c *MyContext) {
	c.String(http.StatusOK, "Welcome Home")
}

// Logger ...
func Logger(c lars.Context) {

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := c.Request().URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", c.Request().Method, c.Response().Status(), path, stop.Sub(start))
}
