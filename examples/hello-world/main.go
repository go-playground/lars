package main

import (
	"net/http"

	"github.com/go-playground/lars"
	mw "github.com/go-playground/lars/examples/middleware/logging-recovery"
)

func main() {

	l := lars.New()
	l.Use(mw.LoggingAndRecovery) // LoggingAndRecovery is just an example copy paste and modify to your needs

	l.Get("/", HelloWorld)

	http.ListenAndServe(":3007", l.Serve())
}

// HelloWorld ...
func HelloWorld(c lars.Context) {
	c.Response().Write([]byte("Hello World"))
}
