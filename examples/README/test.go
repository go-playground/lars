package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-experimental/lcars"
)

func main() {

	l := lcars.New()
	l.Use(Logger)

	l.Get("/", HelloWorld)
	l.Get("/redirect", Redirect)

	http.ListenAndServe(":3007", l.Serve())
}

func HelloWorld(c *lcars.Context) {
	c.Response().Write([]byte("Hello World"))
}

func Redirect(c *lcars.Context) {
	c.Response().Write([]byte("Redirect"))
}

// Logger ...
func Logger(c *lcars.Context) {

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
