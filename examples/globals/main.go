package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/lars"
)

// This is a contrived example using globals as I would use it in production
// I would break things into separate files but all here for simplicity

// ApplicationGlobals houses all the application info for use.
type ApplicationGlobals struct {
	// DB - some database connection
	Log *log.Logger
	// Translator - some i18n translator
	// JSON - encoder/decoder
	// Schema - gorilla schema
	// .......
}

// Reset gets called just before a new HTTP request starts calling
// middleware + handlers
func (g *ApplicationGlobals) Reset(c *lars.Context) {
	// DB = new database connection or reset....
	//
	// We don't touch translator + log as they don't change per request
}

// Done gets called after the HTTP request has completed right before
// Context gets put back into the pool
func (g *ApplicationGlobals) Done() {
	// DB.Close()
}

var _ lars.IGlobals = &ApplicationGlobals{} // ensures ApplicationGlobals complies with lasr.IGlobals at compile time

func main() {

	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// translator := ...
	// db := ... base db connection or info
	// json := ...
	// schema := ...

	globalsFn := func() lars.IGlobals {
		return &ApplicationGlobals{
			Log: logger,
			// Translator: translator,
			// DB: db,
			// JSON: json,
			// schema:schema,
		}
	}

	l := lars.New()
	l.RegisterGlobals(globalsFn)
	l.Use(Logger)

	l.Get("/", Home)

	http.ListenAndServe(":3007", l.Serve())
}

// Home ...
func Home(c *lars.Context) {

	app := c.Globals.(*ApplicationGlobals)

	var username string

	// username = app.DB.find(user by .....)

	app.Log.Println("Found User")

	c.Response().Write([]byte("Welcome Home " + username))
}

// Logger ...
func Logger(c *lars.Context) {

	req := c.Request()

	start := time.Now()

	c.Next()

	stop := time.Now()
	path := req.URL.Path

	if path == "" {
		path = "/"
	}

	log.Printf("%s %d %s %s", req.Method, c.Response().Status(), path, stop.Sub(start))
}
