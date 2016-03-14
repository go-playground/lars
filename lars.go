package lars

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"sync"
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

	Gzip = "gzip"

	WildcardParam = "*wildcard"

	basePath = "/"
	blank    = ""

	slashByte = '/'
	colonByte = ':'
	startByte = '*'
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
	router *Router

	// mostParams used to keep track of the most amount of
	// params in any URL and this will set the default capacity
	// of eachContext Params
	mostParams uint8

	// function that gets called to create the context object... is total overridable using RegisterContext
	contextFunc ContextFunc

	pool sync.Pool

	http404 HandlersChain // 404 Not Found
	http405 HandlersChain // 405 Method Not Allowed

	notFound HandlersChain

	customHandlersFuncs customHandlers

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
}

var (
	default404Handler = func(c Context) {
		http.Error(c.Response(), http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c Context) {

		m, _ := c.Get("methods")
		methods := m.(chainMethods)

		res := c.Response()

		for _, k := range methods {
			res.Header().Add("Allow", k.method)
		}

		res.WriteHeader(http.StatusMethodNotAllowed)
	}
)

// New Creates and returns a new lars instance
func New() *LARS {

	l := &LARS{
		routeGroup: routeGroup{
			middleware: make(HandlersChain, 0),
		},
		contextFunc: func(l *LARS) Context {
			return NewContext(l)
		},
		mostParams:             0,
		http404:                []HandlerFunc{default404Handler},
		http405:                []HandlerFunc{methodNotAllowedHandler},
		redirectTrailingSlash:  true,
		handleMethodNotAllowed: false,
	}

	l.routeGroup.lars = l
	l.router = newRouter(l)
	l.pool.New = func() interface{} {

		c := l.contextFunc(l)
		b := c.BaseContext()
		b.parent = c

		return b
	}

	return l
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

	return http.HandlerFunc(l.serveHTTP)
}

// Conforms to the http.Handler interface.
func (l *LARS) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c := l.pool.Get().(*Ctx)

	c.parent.Reset(w, r)
	l.router.find(c, true)
	c.parent.Next()

	c.parent.RequestComplete()
	l.pool.Put(c)
}

// RouteMap contains a single routes full path
// and other information
type RouteMap struct {
	Depth   int    `json:"depth"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
}

// GetRouteMap returns an array of all registered routes
func (l *LARS) GetRouteMap() []*RouteMap {

	cn := l.router.tree
	var routes []*RouteMap

	results := getNodeRoutes(cn, "/", 0)
	if results != nil && len(results) > 0 {
		routes = append(routes, results...)
	}

	if cn.params != nil {

		pn := cn.params
		pPrefix := "/" + ":" + pn.param

		pResults := getNodeRoutes(pn, pPrefix, 1)
		if pResults != nil && len(pResults) > 0 {

			routes = append(routes, pResults...)
		}

		if pn.wild != nil {

			wResults := getNodeRoutes(pn.wild, pPrefix+"/*", 2)
			if wResults != nil && len(wResults) > 0 {

				routes = append(routes, wResults...)
			}
		}

		pResults = parseTree(pn, pPrefix+"/", 2)
		if pResults != nil && len(pResults) > 0 {
			routes = append(routes, pResults...)
		}

	}

	if cn.wild != nil {
		wPrefix := "/" + "*"

		wResults := getNodeRoutes(cn.wild, wPrefix, 1)
		if wResults != nil && len(wResults) > 0 {
			routes = append(routes, wResults...)
		}
	}

	children := parseTree(cn, "/", 1)
	if children != nil && len(children) > 0 {
		routes = append(routes, children...)
	}

	return routes
}

func parseTree(n *node, prefix string, depth int) []*RouteMap {

	var routes []*RouteMap
	i := 0
	ordered := make([]string, len(n.static))

	for k := range n.static {
		ordered[i] = k
		i++
	}

	sort.Strings(ordered)

	var key string
	var nn *node
	var newPrefix string

	for i = 0; i < len(ordered); i++ {
		key = ordered[i]
		nn = n.static[ordered[i]]
		newPrefix = prefix + key

		// static
		results := getNodeRoutes(nn, newPrefix, depth)
		if results != nil && len(results) > 0 {
			routes = append(routes, results...)
		}

		//params + params wild
		if nn.params != nil {

			pn := nn.params
			pPrefix := newPrefix + ":" + pn.param

			pResults := getNodeRoutes(pn, pPrefix, depth+1)
			if pResults != nil && len(pResults) > 0 {
				routes = append(routes, pResults...)
			}

			if pn.wild != nil {

				wResults := getNodeRoutes(pn.wild, pPrefix+"/*", depth+2)
				if wResults != nil && len(wResults) > 0 {
					routes = append(routes, wResults...)
				}
			}

			pResults = parseTree(pn, pPrefix+"/", depth+2)
			if pResults != nil && len(pResults) > 0 {
				routes = append(routes, pResults...)
			}

		}

		// wild
		if nn.wild != nil {
			wPrefix := newPrefix + "*"

			wResults := getNodeRoutes(nn.wild, wPrefix, depth+1)
			if wResults != nil && len(wResults) > 0 {
				routes = append(routes, wResults...)
			}
		}

		results = parseTree(nn, newPrefix, depth+1)
		if results != nil && len(results) > 0 {
			routes = append(routes, results...)
		}
	}

	return routes
}

func getNodeRoutes(n *node, path string, depth int) []*RouteMap {

	var routes []*RouteMap
	var name string

	for _, r := range n.chains {

		_, name = n.chains.find(r.method)

		routes = append(routes, &RouteMap{
			Depth:   depth,
			Path:    path,
			Method:  r.method,
			Handler: name,
		})
	}

	for _, r := range n.parmsSlashChains {

		_, name = n.parmsSlashChains.find(r.method)

		routes = append(routes, &RouteMap{
			Depth:   depth,
			Path:    path + "/",
			Method:  r.method,
			Handler: name,
		})
	}

	return routes
}
