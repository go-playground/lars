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
	*lars.Ctx  // a little embedding
	AppContext *ApplicationGlobals
}

func newContext(l *lars.LARS) lars.Context {
	return &MyContext{
		Ctx:        lars.NewContext(l),
		AppContext: newGlobals(),
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
func Home(c *MyContext) {

	var username string

	// username = c.AppContext.DB.find(user by .....)

	c.AppContext.Log.Println("Found User")

	c.Response().Write([]byte("Welcome Home " + username))
}

// Users ...
func Users(c *MyContext) {

	c.AppContext.Log.Println("In Users Function")

	c.Response().Write([]byte("Users"))
}

// User ...
func User(c *MyContext) {

	id := c.Param("id")

	var username string

	// username = c.AppContext.DB.find(user by id.....)

	c.AppContext.Log.Println("Found User")

	c.Response().Write([]byte("Welcome " + username + " with id " + id))
}

// UserProfile ...
func UserProfile(c *MyContext) {

	id := c.Param("id")

	var profile string

	// profile = c.AppContext.DB.find(user profile by .....)

	c.AppContext.Log.Println("Found User Profile")

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
