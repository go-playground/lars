/*
Package lars - Library Access/Retrieval System, is a fast radix-tree based, zero allocation, HTTP router for Go.


Usage

Below is a simple example, for a full example see here https://github.com/go-playground/lars/blob/master/_examples/all-in-one/main.go

	package main

	import (
		"net/http"

		"github.com/go-playground/lars"
		mw "github.com/go-playground/lars/_examples/middleware/logging-recovery"
	)

	func main() {

		l := lars.New()
		l.Use(mw.LoggingAndRecovery) // LoggingAndRecovery is just an example copy
					     // paste and modify to your needs
		l.Get("/", HelloWorld)

		http.ListenAndServe(":3007", l.Serve())
	}

	// HelloWorld ...
	func HelloWorld(c lars.Context) {
		c.Response().Write([]byte("Hello World"))

		// this will also work, Response() complies with http.ResponseWriter interface
		fmt.Fprint(c.Response(), "Hello World")
	}

URL Params

example param usage

	l := l.New()
	l.Get("/user/:id", UserHandler)

	// serve css, js etc.. c.Param(lars.WildcardParam) will return the
	// remaining path if you need to use it in a custom handler...
	l.Get("/static/*", http.FileServer(http.Dir("static/")))

	NOTE: Since this router has only explicit matches, you can not register static routes
	and parameters for the same path segment. For example you can not register the patterns
	/user/new and /user/:user for the same request method at the same time. The routing of
	different request methods is independent from each other. I was initially against this,
	and this router allowed it in a previous version, however it nearly cost me in a big
	app where the dynamic param value say :type actually could have matched another static
	route and that's just too dangerous, so it is no longer allowed.


Groups

example group definitions

	...
	l.Use(LoggingAndRecovery)
	...
	l.Post("/users/add", ...)

	// creates a group for user + inherits all middleware registered using l.Use()
	user := l.Group("/user/:userid")
	user.Get("", ...)
	user.Post("", ...)
	user.Delete("/delete", ...)

	contactInfo := user.Group("/contact-info/:ciid")
	contactinfo.Delete("/delete", ...)

	// creates a group for others + inherits all middleware registered using l.Use() +
	// adds OtherHandler to middleware
	others := l.GroupWithMore("/others", OtherHandler)

	// creates a group for admin WITH NO MIDDLEWARE... more can be added using admin.Use()
	admin := l.GroupWithNone("/admin")
	admin.Use(SomeAdminSecurityMiddleware)
	...


Custom Context - Avoid Type Casting - Custom Handlers


example context + custom handlers

	...
	// MyContext is a custom context
	type MyContext struct {
		*lars.Ctx  // a little dash of Duck Typing....
	}

	// CustomContextFunction is a function that is specific to your applications
	// needs that you added
	func (mc *MyContext) CustomContextFunction() {
		// do something
	}

	// newContext is the function that creates your custom context +
	// contains lars's default context
	func newContext(l *lars.LARS) lars.Context {
		return &MyContext{
			Ctx:        lars.NewContext(l),
		}
	}

	// casts custom context and calls you custom handler so you don't have to
	// type cast lars.Context everywhere
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

	// Home ...notice the receiver is *MyContext, castCustomContext handled the
	// type casting for us; quite the time saver if you ask me.
	func Home(c *MyContext) {

		c.CustomContextFunction()
		...
	}

Decoding Body

For full example see https://github.com/go-playground/lars/blob/master/_examples/decode/main.go
currently JSON, XML, FORM + Multipart Form's are support out of the box.

	// first argument denotes yes or no I would like URL query parameter fields
	// to be included. i.e. 'id' in route '/user/:id' should it be included.
	// run, then change to false and you'll see user.ID is not populated.
	if err := c.Decode(true, maxBytes, &user); err != nil {
		log.Println(err)
	}


Misc

misc examples and noteworthy features

	...
	// can register multiple handlers, the last is considered the last in the chain and
	// others considered middleware, but just for this route and not added to middleware
	// like l.Use() does.
	l.Get(/"home", AdditionalHandler, HomeHandler)

	// set custom 404 ( not Found ) handler
	l.Register404(404Handler)

	// Redirect to or from ending slash if route not found, default is true
	l.SetRedirectTrailingSlash(true)

	// Handle 405 ( Method Not allowed ), default is false
	l.SetHandle405MethodNotAllowed(false)

	// automatically handle OPTION requests; manually configured
	// OPTION handlers take precedence. default true
	l.SetAutomaticallyHandleOPTIONS(set bool)

	// register custom context
	l.RegisterContext(ContextFunc)

	// Register custom handler type, see util.go
	// https://github.com/go-playground/lars/blob/master/util.go#L62 for example handler
	// creation
	l.RegisterCustomHandler(interface{}, CustomHandlerFunc)

	// NativeChainHandler is used as a helper to create your own custom handlers, or use
	// custom handlers that already exist an example usage can be found here
	// https://github.com/go-playground/lars/blob/master/util.go#L86, below is an example
	// using nosurf CSRF middleware
	l.Use(nosurf.NewPure(lars.NativeChainHandler))

	// Context has 2 methods of which you should be aware of ParseForm and
	// ParseMulipartForm, they just call the default http functions but provide one more
	// additional feature, they copy the URL params to the request Forms variables, just
	// like Query parameters would have been. The functions are for convenience and are
	// totally optional.
*/
package lars
