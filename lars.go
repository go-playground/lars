package lars

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/form"
)

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
	OctetStream                      = "application/octet-stream"

	//---------
	// Charset
	//---------

	CharsetUTF8 = "charset=utf-8"

	//---------
	// Headers
	//---------

	AcceptedLanguage   = "Accept-Language"
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
	XRealIP            = "X-Real-Ip"
	Allow              = "Allow"
	Origin             = "Origin"

	Gzip = "gzip"

	WildcardParam = "*wildcard"

	basePath = "/"
	blank    = ""

	slashByte = '/'
	paramByte = ':'
	wildByte  = '*'
)

// Handler is the type used in registering handlers.
// NOTE: these handlers may get wrapped by the HandlerFunc
// type internally.
type Handler interface{}

// HandlerFunc is the internal handler type used for middleware and handlers
type HandlerFunc func(Context)

// HandlersChain is an array of HanderFunc handlers to run
type HandlersChain []HandlerFunc

// ContextFunc is the function to run when creating a new context
type ContextFunc func(l *LARS) Context

// CustomHandlerFunc wraped by HandlerFunc and called where you can type cast both Context and Handler
// and call Handler
type CustomHandlerFunc func(Context, Handler)

// customHandlers is a map of your registered custom CustomHandlerFunc's
// used in determining how to wrap them.
type customHandlers map[reflect.Type]CustomHandlerFunc

// LARS is the main routing instance
type LARS struct {
	routeGroup
	trees map[string]*node

	// function that gets called to create the context object... is total overridable using RegisterContext
	contextFunc ContextFunc

	pool sync.Pool

	http404 HandlersChain // 404 Not Found
	http405 HandlersChain // 405 Method Not Allowed

	automaticOPTIONS HandlersChain
	notFound         HandlersChain

	customHandlersFuncs customHandlers

	// mostParams used to keep track of the most amount of
	// params in any URL and this will set the default capacity
	// of eachContext Params
	mostParams uint8

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	redirectTrailingSlash bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	handleMethodNotAllowed bool

	// if enabled automatically handles OPTION requests; manually configured OPTION
	// handlers take precidence. default true
	automaticallyHandleOPTIONS bool

	// form decoder + once initialization
	formDecoder     *form.Decoder
	formDecoderInit sync.Once
}

// RouteMap contains a single routes full path
// and other information
type RouteMap struct {
	Depth   int    `json:"depth"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
}

var (
	default404Handler = func(c Context) {
		http.Error(c.Response(), http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c Context) {
		c.Response().WriteHeader(http.StatusMethodNotAllowed)
	}

	automaticOPTIONSHandler = func(c Context) {
		c.Response().WriteHeader(http.StatusOK)
	}
)

// New Creates and returns a new lars instance
func New() *LARS {

	l := &LARS{
		routeGroup: routeGroup{
			middleware: make(HandlersChain, 0),
		},
		trees: make(map[string]*node),
		contextFunc: func(l *LARS) Context {
			return NewContext(l)
		},
		mostParams:                 0,
		http404:                    []HandlerFunc{default404Handler},
		http405:                    []HandlerFunc{methodNotAllowedHandler},
		redirectTrailingSlash:      true,
		handleMethodNotAllowed:     false,
		automaticallyHandleOPTIONS: true,
	}

	l.routeGroup.lars = l
	l.pool.New = func() interface{} {

		c := l.contextFunc(l)
		b := c.BaseContext()
		b.parent = c

		return b
	}

	return l
}

// BuiltInFormDecoder returns the built in form decoder github.com/go-playground/form
// in order for custom type to be registered.
func (l *LARS) BuiltInFormDecoder() *form.Decoder {

	l.initFormDecoder()

	return l.formDecoder
}

func (l *LARS) initFormDecoder() {
	l.formDecoderInit.Do(func() {
		l.formDecoder = form.NewDecoder()
	})
}

// RegisterCustomHandler registers a custom handler that gets wrapped by HandlerFunc
func (l *LARS) RegisterCustomHandler(customType interface{}, fn CustomHandlerFunc) {

	if l.customHandlersFuncs == nil {
		l.customHandlersFuncs = make(customHandlers)
	}

	t := reflect.TypeOf(customType)

	if _, ok := l.customHandlersFuncs[t]; ok {
		panic(fmt.Sprint("Custom Type + CustomHandlerFunc already declared: ", t))
	}

	l.customHandlersFuncs[t] = fn
}

// RegisterContext registers a custom Context function for creation
// and resetting of a global object passed per http request
func (l *LARS) RegisterContext(fn ContextFunc) {
	l.contextFunc = fn
}

// Register404 alows for overriding of the not found handler function.
// NOTE: this is run after not finding a route even after redirecting with the trailing slash
func (l *LARS) Register404(notFound ...Handler) {

	chain := make(HandlersChain, len(notFound))

	for i, h := range notFound {
		chain[i] = l.wrapHandler(h)
	}

	l.http404 = chain
}

// SetAutomaticallyHandleOPTIONS tells lars whether to
// automatically handle OPTION requests; manually configured
// OPTION handlers take precedence. default true
func (l *LARS) SetAutomaticallyHandleOPTIONS(set bool) {
	l.automaticallyHandleOPTIONS = set
}

// SetRedirectTrailingSlash tells lars whether to try
// and fix a URL by trying to find it
// lowercase -> with or without slash -> 404
func (l *LARS) SetRedirectTrailingSlash(set bool) {
	l.redirectTrailingSlash = set
}

// SetHandle405MethodNotAllowed tells lars whether to
// handle the http 405 Method Not Allowed status code
func (l *LARS) SetHandle405MethodNotAllowed(set bool) {
	l.handleMethodNotAllowed = set
}

// Serve returns an http.Handler to be used.
func (l *LARS) Serve() http.Handler {

	// reserved for any logic that needs to happen before serving starts.
	// i.e. although this router does not use priority to determine route order
	// could add sorting of tree nodes here....

	l.notFound = make(HandlersChain, len(l.middleware)+len(l.http404))
	copy(l.notFound, l.middleware)
	copy(l.notFound[len(l.middleware):], l.http404)

	if l.automaticallyHandleOPTIONS {
		l.automaticOPTIONS = make(HandlersChain, len(l.middleware)+1)
		copy(l.automaticOPTIONS, l.middleware)
		copy(l.automaticOPTIONS[len(l.middleware):], []HandlerFunc{automaticOPTIONSHandler})
	}

	return http.HandlerFunc(l.serveHTTP)
}

// Conforms to the http.Handler interface.
func (l *LARS) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c := l.pool.Get().(*Ctx)

	c.parent.RequestStart(w, r)

	if root := l.trees[r.Method]; root != nil {

		if c.handlers, c.params, c.handlerName = root.find(r.URL.Path, c.params); c.handlers == nil {

			c.params = c.params[0:0]

			if l.redirectTrailingSlash && len(r.URL.Path) > 1 {

				// find again all lowercase
				orig := r.URL.Path
				lc := strings.ToLower(orig)

				if lc != r.URL.Path {

					if c.handlers, _, _ = root.find(lc, c.params); c.handlers != nil {
						r.URL.Path = lc
						c.handlers = l.redirect(r.Method, r.URL.String())
						r.URL.Path = orig
						goto END
					}
				}

				if lc[len(lc)-1:] == basePath {
					lc = lc[:len(lc)-1]
				} else {
					lc = lc + basePath
				}

				if c.handlers, _, _ = root.find(lc, c.params); c.handlers != nil {
					r.URL.Path = lc
					c.handlers = l.redirect(r.Method, r.URL.String())
					r.URL.Path = orig
					goto END
				}
			}

		} else {
			goto END
		}
	}

	if l.automaticallyHandleOPTIONS && r.Method == OPTIONS {
		l.getOptions(c)
		goto END
	}

	if l.handleMethodNotAllowed {

		if l.checkMethodNotAllowed(c) {
			goto END
		}
	}

	// not found
	c.handlers = l.notFound

END:

	c.parent.Next()
	c.parent.RequestEnd()

	l.pool.Put(c)
}

func (l *LARS) getOptions(c *Ctx) {

	if c.request.URL.Path == "*" { // check server-wide OPTIONS

		for m := range l.trees {

			if m == OPTIONS {
				continue
			}

			c.response.Header().Add(Allow, m)
		}

	} else {
		for m, tree := range l.trees {

			if m == c.request.Method || m == OPTIONS {
				continue
			}

			if c.handlers, _, _ = tree.find(c.request.URL.Path, c.params); c.handlers != nil {
				c.response.Header().Add(Allow, m)
			}
		}

	}

	c.response.Header().Add(Allow, OPTIONS)
	c.handlers = l.automaticOPTIONS

	return
}

func (l *LARS) checkMethodNotAllowed(c *Ctx) (found bool) {

	for m, tree := range l.trees {

		if m != c.request.Method {
			if c.handlers, _, _ = tree.find(c.request.URL.Path, c.params); c.handlers != nil {
				// add methods
				c.response.Header().Add(Allow, m)
				found = true
			}
		}
	}

	if found {
		c.handlers = l.http405
	}

	return
}
