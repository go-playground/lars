/*
Package lars - Library Access/Retrieval System, is a fast radix-tree based, zero allocation, HTTP router for Go.

Usage

Below is a simple example, for a full example see here https://github.com/go-playground/lars/blob/master/examples/all-in-one/main.go

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

		// this will also work, Response() complies with http.ResponseWriter interface
		fmt.Fprint(c.Response(), "Hello World")
	}

URL Params

	l := l.New()
	l.Get("/user/:id", UserHandler)
	l.Get("/static/*", http.FileServer(http.Dir("static/"))) // serve css, js etc.. c.Param(lars.WildcardParam) will return the
								 // remaining path if you need to use it in a custom handler...

	NOTE: Since this router has only explicit matches, you can not register static routes and parameters for the same path segment.
	For example you can not register the patterns /user/new and /user/:user for the same request method at the same time.
	The routing of different request methods is independent from each other. I was initially against this, and this router allowed
	it in a previous version, however it nearly cost me in a big app where the dynamic param value say :type actually could have matched
	another static route and that's just too dangerous, so it is no longer allowed.

Groups

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

	// creates a group for others + inherits all middleware registered using l.Use() + adds OtherHandler to middleware
	others := l.Group("/others", OtherHandler)

	// creates a group for admin WITH NO MIDDLEWARE... more can be added using admin.Use()
	admin := l.Group("/admin",nil)
	admin.Use(SomeAdminSecurityMiddleware)
	...

Custom Context + Avoid Type Casting / Custom Handlers

	...
	// MyContext is a custom context
	type MyContext struct {
		*lars.Ctx  // a little dash of Duck Typing....
	}

	// RequestStart overriding
	func (mc *MyContext) RequestStart(w http.ResponseWriter, r *http.Request) {
		mc.Ctx.RequestStart(w, r) // MUST be called!

		// do whatever you need to on request start, db connections, variable init...
	}

	// RequestEnd overriding
	func (mc *MyContext) RequestEnd() {

		// do whatever you need on request finish, reset variables, db connections...

		mc.Ctx.RequestEnd() // MUST be called!
	}

	// CustomContextFunction is a function that is specific to your applications needs that you added
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

	// casts custom context and calls you custom handler so you don;t have to type cast lars.Context everywhere
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

	// Home ...notice the receiver is *MyContext, castCustomContext handled the type casting for us
	// quite the time saver if you ask me.
	func Home(c *MyContext) {

		c.CustomContextFunction()
		...
	}

Misc

	...
	// can register multiple handlers, the last is considered the last in the chain and others
	// considered middleware, but just for this route and not added to middleware like l.Use() does.
	l.Get(/"home", AdditionalHandler, HomeHandler)

	// set custom 404 ( not Found ) handler
	l.Register404(404Handler)

	// Redirect to or from ending slash if route not found, default is true
	l.SetRedirectTrailingSlash(true)

	// Handle 405 ( Method Not allowed ), default is false
	l.SetHandle405MethodNotAllowed(false)

	// register custom context
	l.RegisterContext(ContextFunc)

	// Register custom handler type, see util.go https://github.com/go-playground/lars/blob/master/util.go#L62 for example handler creation
	l.RegisterCustomHandler(interface{}, CustomHandlerFunc)

	// Context has 2 methods of which you should be aware of ParseForm and ParseMulipartForm, they just call the default http functions but
	// provide one more additional feature, they copy the URL params to the request Forms variables, just like Query parameters would have been.
	// The functions are for convenience and are totally optional.
*/
package lars
