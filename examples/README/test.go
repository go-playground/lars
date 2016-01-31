package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-playground/lars"
)

func main() {

	l := lars.New()
	l.Use(Logger)

	l.Get("/", HelloWorld)
	l.Get("/redirect", Redirect)

	http.ListenAndServe(":3007", l.Serve())
}

// HelloWorld ...
func HelloWorld(c *lars.Context) {
	c.Response.Write([]byte("Hello World"))
}

// Redirect ...
func Redirect(c *lars.Context) {
	c.Response.Write([]byte("Redirect"))
}

// Logger ...
func Logger(c *lars.Context) {

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := c.Request.URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", c.Request.Method, c.Response.Status(), path, stop.Sub(start))
}
