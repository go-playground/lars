##LARS
<img align="right" src="https://raw.githubusercontent.com/go-playground/lars/master/examples/README/test.gif">
![Project status](https://img.shields.io/badge/version-0.9.8-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/projects/4351aa2d-2f94-40be-a6ef-85c248490378/679708/badge.svg)](https://semaphoreci.com/joeybloggs/lars)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/lars/badge.svg?branch=master)](https://coveralls.io/github/go-playground/lars?branch=master)
[![Go Report Card](http://goreportcard.com/badge/go-playground/lars)](http://goreportcard.com/badge/go-playground/lars)
[![GoDoc](https://godoc.org/github.com/go-playground/lars?status.svg)](https://godoc.org/github.com/go-playground/lars)
![License](https://img.shields.io/dub/l/vibe-d.svg)

LARS is a fast radix-tree based, zero allocation, HTTP router for Go.  [ view examples](https://github.com/go-playground/lars/tree/master/examples)

Why Another HTTP Router?
------------------------
I have noticed that most routers out there, IMHO, are adding too much functionality that doesn't belong in an HTTP router, and they are turning into web frameworks, with all the bloat that entails. LARS aims to remain a simple yet powerful HTTP router that can be plugged into any existing framework; furthermore LARS allowing the passing of global variables + application context that comply with it's IAppContext interface (right on the Context object) makes frameworks redundant as **LARS wraps the framework instead of the framework wrapping LARS** [see example here](https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go)

Unique Features 
--------------
* Context allows the passing of framework/globals/application specific variables via it's AppContext field.
  * The AppContext object is essentially all of the application specific variables and libraries needed by your handlers and functions, keeping a clear separation between your http and application contexts.
* Handles mutiple url patterns not supported by many other routers.
  * the route algorithm was written from scratch and is **NOT** a modification of any other router.
* Contains helpful logic to help prevent adding bad routes, keeping your url's consistent.
  * i.e. /user/:id and /user/:user_id - the second one will fail to add letting you know that :user_id should be :id
* Has an uber simple middleware + handler definitions!!! middleware and handlers actually have the exact same definition!
* Full support for standard/native http Handler + HandlerFunc [see here](https://github.com/go-playground/lars/blob/master/examples/native/main.go)



Installation
-----------

Use go get 

```go
go get github.com/go-playground/lars
``` 

or to update

```go
go get -u github.com/go-playground/lars
``` 

Then import lars package into your code.

```go
import "github.com/go-playground/lars"
``` 

Usage
------
Below is a full example, for a simpler example [see here](https://github.com/go-playground/lars/blob/master/examples/groups/main.go)
```go
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

var _ lars.IAppContext = &ApplicationGlobals{} // ensures ApplicationGlobals complies with lasr.IGlobals at compile time

func main() {

	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// translator := ...
	// db := ... base db connection or info
	// json := ...
	// schema := ...

	globalsFn := func() lars.IAppContext {
		return &ApplicationGlobals{
			Log: logger,
			// Translator: translator,
			// DB: db,
			// JSON: json,
			// schema:schema,
		}
	}

	l := lars.New()
	l.RegisterAppContext(globalsFn)
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
func Home(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	var username string

	// username = app.DB.find(user by .....)

	app.Log.Println("Found User")

	c.Response.Write([]byte("Welcome Home " + username))
}

// Users ...
func Users(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	app.Log.Println("In Users Function")

	c.Response.Write([]byte("Users"))
}

// User ...
func User(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	id := c.Param("id")

	var username string

	// username = app.DB.find(user by id.....)

	app.Log.Println("Found User")

	c.Response.Write([]byte("Welcome " + username + " with id " + id))
}

// UserProfile ...
func UserProfile(c *lars.Context) {

	app := c.AppContext.(*ApplicationGlobals)

	id := c.Param("id")

	var profile string

	// profile = app.DB.find(user profile by .....)

	app.Log.Println("Found User Profile")

	c.Response.Write([]byte("Here's your profile " + profile + " user " + id))
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
```

Native Handler Support
```go
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

	ctx.Response.Write([]byte("Hello World"))
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
```

Middleware
-----------
There are some pre-defined middlewares within the middleware folder; NOTE: that the middleware inside will
comply with the following rule(s):

* Are completely reusable by the community without modification

Other middleware will be listed under the examples/middleware/... folder for a quick copy/paste modify. as an example a logging or
recovery middleware are very application dependent and therefore will be listed under the examples/middleware/...

Benchmarks
-----------
Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go version go1.5.3 darwin/amd64


```go
go test -bench=. -benchmem=true
#GithubAPI Routes: 203
   LARS: 39688 Bytes

#GPlusAPI Routes: 13
   LARS: 3776 Bytes

#ParseAPI Routes: 26
   LARS: 4672 Bytes

#Static Routes: 157
   LARS: 30992 Bytes

PASS
BenchmarkLARS_Param       	20000000	        81.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param5      	10000000	       130 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Param20     	 5000000	       345 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParamWrite  	10000000	       145 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubStatic	20000000	       110 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubParam 	10000000	       129 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GithubAll   	   50000	     37792 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusStatic 	20000000	        67.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusParam  	20000000	        90.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlus2Params	10000000	       127 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_GPlusAll    	 1000000	      1737 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseStatic 	20000000	        75.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseParam  	20000000	        91.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_Parse2Params	20000000	       108 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_ParseAll    	  500000	      3393 ns/op	       0 B/op	       0 allocs/op
BenchmarkLARS_StaticAll   	   50000	     27306 ns/op	       0 B/op	       0 allocs/op

```

This package is inspired by the following 
- [httptreemux](https://github.com/dimfeld/httptreemux)
- [httprouter](https://github.com/julienschmidt/httprouter)
- [echo](https://github.com/labstack/echo)
- [gin](https://github.com/gin-gonic/gin)

License 
--------
This project is licensed unter MIT, for more information look into the LICENSE file.
Copyright (c) 2016 Go Playground


