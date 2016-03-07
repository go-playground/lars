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

	http.ListenAndServe(":3007", l.Serve())
}

// HelloWorld ...
func HelloWorld(w http.ResponseWriter, r *http.Request) {

	// lar's context! get it and ROCK ON!
	ctx := lars.GetContext(w)

	ctx.Response().Write([]byte("Hello World"))
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
