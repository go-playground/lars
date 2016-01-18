package lars

import (
	"net/http"
	"sync"
)

// TODO: tree instance needs to be added to the LARS struct and New() function

// HTTP Constant Terms and Variables
const (
	// CONNECT HTTP method
	CONNECT = "CONNECT"
	// DELETE HTTP method
	DELETE = "DELETE"
	// GET HTTP method
	GET = "GET"
	// HEAD HTTP method
	HEAD = "HEAD"
	// OPTIONS HTTP method
	OPTIONS = "OPTIONS"
	// PATCH HTTP method
	PATCH = "PATCH"
	// POST HTTP method
	POST = "POST"
	// PUT HTTP method
	PUT = "PUT"
	// TRACE HTTP method
	TRACE = "TRACE"

	//-------------
	// Media types
	//-------------

	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + CharsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + CharsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + CharsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + CharsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + CharsetUTF8
	MultipartForm                    = "multipart/form-data"

	//---------
	// Charset
	//---------

	CharsetUTF8 = "charset=utf-8"

	//---------
	// Headers
	//---------

	AcceptEncoding     = "Accept-Encoding"
	Authorization      = "Authorization"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	Location           = "Location"
	Upgrade            = "Upgrade"
	Vary               = "Vary"
	WWWAuthenticate    = "WWW-Authenticate"
	XForwardedFor      = "X-Forwarded-For"
	XRealIP            = "X-Real-IP"

	default404Body = "404 page not found"
	default405Body = "405 method not allowed"

	basePath = "/"
	blank    = ""
)

// Handler is the type used in registering handlers.
// NOTE: these handlers may get wrapped by the HandlerFunc
// type internally.
type Handler interface{}

// HandlerFunc is the internal handler type used for middleware and handlers
type HandlerFunc func(Context)

// HandlersChain is an array of HanderFunc handlers to run
type HandlersChain []HandlerFunc

// ContextFunc is the function that returns a new Context instance.
type ContextFunc func() Context

// LARS is the main routing instance
type LARS struct {
	RouteGroup

	router

	// mostParams used to keep track of the most amount of
	// params in any URL and this will set the default capacity
	// of each context Params
	mostParams uint8

	// router     *router
	//
	pool sync.Pool

	newContext ContextFunc

	http404        HandlerFunc
	httpNotAllowed HandlerFunc

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool
}

var (
	default404Handler = func(c Context) {
		http.Error(c.Response(), default404Body, http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c Context) {
		http.Error(c.Response(), default405Body, http.StatusMethodNotAllowed)
	}
)

// New Creates and returns a new LARS instance
func New() *LARS {

	l := &LARS{
		RouteGroup: RouteGroup{
			middleware: make(HandlersChain, 0),
		},
		router:                 router{},
		mostParams:             0,
		http404:                default404Handler,
		httpNotAllowed:         methodNotAllowedHandler,
		RedirectTrailingSlash:  true,
		HandleMethodNotAllowed: false,
	}

	l.RouteGroup.lars = l
	l.router.lars = l
	l.newContext = func() Context {
		return NewContext(l)
	}
	l.pool.New = func() interface{} {
		return l.newContext()
	}

	return l
}

// RegisterContext sets a custom Context pool initializer
// for use with your own custom context.
func (l *LARS) RegisterContext(fn ContextFunc) {
	l.newContext = fn
}

// Register404 alows for overriding of the not found handler function.
// NOTE: this is run after not finding a route even after redirecting with the trailing slash
func (l *LARS) Register404(notFound Handler) {
	l.http404 = wrapHandler(notFound)
}

// Conforms to the http.Handler interface.
func (l *LARS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := l.pool.Get().(Context)
	c.Reset(w, r)

	// handle requests here passing in c.UnderlyingContext() aka *ctx
	// and everything can be set on the object without a return value and

	c.Next()

	l.pool.Put(c)
}
