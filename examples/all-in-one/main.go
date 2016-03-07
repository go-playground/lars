package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/lars"
)

// This is a contrived example of how I would use in production
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
func (g *ApplicationGlobals) Reset() {
	// DB = new database connection or reset....
	//
	// We don't touch translator + log as they don't change per request
}

// Done gets called after the HTTP request has completed right before
// Context gets put back into the pool
func (g *ApplicationGlobals) Done() {
	// DB.Close()
}

func newGlobals() *ApplicationGlobals {

	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// translator := ...
	// db := ... base db connection or info
	// json := ...
	// schema := ...

	return &ApplicationGlobals{
		Log: logger,
		// Translator: translator,
		// DB: db,
		// JSON: json,
		// schema:schema,
	}
}

// MyContext is a custom context
type MyContext struct {
	*lars.Ctx  // a little dash of Duck Typing....
	AppContext *ApplicationGlobals
}

// Reset overriding
func (mc *MyContext) Reset(w http.ResponseWriter, r *http.Request) {

	// call lars context reset, must be done
	mc.Ctx.Reset(w, r)
	mc.AppContext.Reset()
}

// RequestComplete overriding
func (mc *MyContext) RequestComplete() {
	mc.AppContext.Done()
}

func newContext(l *lars.LARS) lars.Context {
	return &MyContext{
		Ctx:        lars.NewContext(l),
		AppContext: newGlobals(),
	}
}

func main() {

	l := lars.New()
	l.RegisterContext(newContext) // all gets cached in pools for you
	l.Use(Logger)

	l.Get("/", Home)

	users := l.Group("/users")
	users.Get("", Users)

	// you can break it up however you with, just demonstrating that you can
	// have groups of group
	user := users.Group("/:id")
	user.Get("", User)
	user.Get("/profile", UserProfile)

	http.ListenAndServe(":3007", l.Serve())
}

// Home ...
func Home(c lars.Context) {

	ctx := c.(*MyContext)

	var username string

	// username = ctx.AppContext.DB.find(user by .....)

	ctx.AppContext.Log.Println("Found User")

	c.Response().Write([]byte("Welcome Home " + username))
}

// Users ...
func Users(c lars.Context) {

	ctx := c.(*MyContext)

	ctx.AppContext.Log.Println("In Users Function")

	c.Response().Write([]byte("Users"))
}

// User ...
func User(c lars.Context) {

	ctx := c.(*MyContext)

	id := c.Param("id")

	var username string

	// username = ctx.AppContext.DB.find(user by id.....)

	ctx.AppContext.Log.Println("Found User")

	c.Response().Write([]byte("Welcome " + username + " with id " + id))
}

// UserProfile ...
func UserProfile(c lars.Context) {

	ctx := c.(*MyContext)

	id := c.Param("id")

	var profile string

	// profile = ctx.AppContext.DB.find(user profile by .....)

	ctx.AppContext.Log.Println("Found User Profile")

	c.Response().Write([]byte("Here's your profile " + profile + " user " + id))
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
