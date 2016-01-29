package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-experimental/lars"
)

func main() {

	l := lars.New()
	l.Use(Logger)

	l.Get("/", HelloWorld)
	l.Get("/redirect", Redirect)

	http.ListenAndServe(":3007", l.Serve())
}

func HelloWorld(c *lars.Context) {
	c.Response().Write([]byte("Hello World"))
}

func Redirect(c *lars.Context) {
	c.Response().Write([]byte("Redirect"))
}

// Logger ...
func Logger(c *lars.Context) {

	req := c.Request()

	start := time.Now()

	c.Next()

	stop := time.Now()
	method := req.Method
	path := req.URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", method, c.Response().Status(), path, stop.Sub(start))
}
