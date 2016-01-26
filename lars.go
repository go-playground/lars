package lars

import (
	"net/http"
	"net/url"
	"strings"
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

// ContextFunc is the function that returns a new Context instance.
type ContextFunc func() Context

// LARS is the main routing instance
type LARS struct {
	routeGroup

	head *node

	// mostParams used to keep track of the most amount of
	// params in any URL and this will set the default capacity
	// of each context Params
	mostParams uint8

	pool sync.Pool

	newContext ContextFunc

	http404 HandlersChain // 404 Not Found
	http405 HandlersChain // 405 Method Not Allowed

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
		http.Error(c.Response(), default404Body, http.StatusNotFound)
	}

	methodNotAllowedHandler = func(c Context) {

		m, ok := c.Get("methods")
		if !ok {
			return
		}

		methods := m.(chainMethods)

		res := c.Response()

		for k := range methods {
			res.Header().Add("Allow", k)
		}

		res.WriteHeader(http.StatusMethodNotAllowed)
	}
)

// New Creates and returns a new LARS instance
func New() *LARS {

	l := &LARS{
		routeGroup: routeGroup{
			middleware: make(HandlersChain, 0),
		},
		head: &node{
			static: nodes{},
		},
		mostParams:             0,
		http404:                []HandlerFunc{default404Handler},
		http405:                []HandlerFunc{methodNotAllowedHandler},
		redirectTrailingSlash:  true,
		handleMethodNotAllowed: false,
	}

	l.routeGroup.lars = l
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
func (l *LARS) Register404(notFound ...Handler) {

	chain := make(HandlersChain, len(notFound))

	for i, h := range notFound {
		chain[i] = wrapHandler(h)
	}

	l.http404 = chain
}

// RedirectTrailingSlash tells LARS whether to try
// and fix a URL by trying to find it
// lowercase -> with or without slash -> 404
func (l *LARS) SetRedirectTrailingSlash(set bool) {
	l.redirectTrailingSlash = set
}

// SetHandle405MethodNotAllowed tells LARS whether to
// handle the http 405 Method Not Allowed status code
func (l *LARS) SetHandle405MethodNotAllowed(set bool) {
	l.handleMethodNotAllowed = set
}

// Serve returns an http.Handler to be used.
func (l *LARS) Serve() http.Handler {

	// reserved for any logic that needs to happen before serving starts.
	// i.e. although this router does not use priority to determine route order
	// could add sorting of tree nodes here....

	return http.HandlerFunc(l.serveHTTP)
}

// Conforms to the http.Handler interface.
func (l *LARS) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c := l.pool.Get().(Context)
	c.Reset(w, r)

	// USE PATH as elements are query escaped
	// l.find(c.UnderlyingContext(), r.Method, r.URL.Path[1:])
	l.find(c.UnderlyingContext(), true)

	c.Next()

	l.pool.Put(c)
}

func (l *LARS) add(method string, path string, rg *routeGroup, h HandlersChain) {

	origPath := path
	cn := l.head

	var (
		start  int
		end    int
		j      int
		c      byte
		en     *node
		ok     bool
		chunk  string
		err    error
		pCount uint8 = 1
	)

	if path, err = url.QueryUnescape(path); err != nil {
		panic("Query Unescape Error on path '" + origPath + "': " + err.Error())
	}

	if path == blank {
		path = basePath
	}

	path = path[1:]

MAIN:
	for ; end < len(path); end++ {

		c = path[end]

		if c == slashByte {

			j = end + 1
			chunk = path[start:j]

			// check for existing node
			if en, ok = cn.static[chunk]; ok {
				cn = en
				start = j

				continue
			}

			// no existing node, adding new one
			if cn.static == nil {
				cn.static = nodes{}
			}

			nn := &node{}
			cn.static[chunk] = nn
			cn = nn
			start = j

			continue
		}

		if c == colonByte {
			end++
			start = end

			// extract param name
			for ; end < len(path); end++ {
				if path[end] != slashByte {
					continue
				}

				chunk = path[start:end]

				// existing param node?
				if cn.params != nil {

					// can't have same prefix paths with different param names example:
					// /users/:id/profile
					// /users/:user_id/profile/settings
					// both params above must be either :id or :user_id, no mix & match
					if cn.params.param != chunk {
						panic("Different param names defined for path '" + origPath + "', param '" + chunk + "'' should be '" + cn.params.param + "'")
					}

					pCount++
					cn = cn.params
					start = end + 1 // may be wrong here might be + 2 or plus nothing

					continue MAIN
				}

				// wild already exists! then will conflict
				if cn.wild != nil {
					panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
				}

				nn := &node{
					param: chunk,
				}

				cn.params = nn
				pCount++
				cn = nn
				start = end + 1 // may be wrong here might be + 2 or plus nothing

				continue MAIN
			}

			// param name did not end in slash, extract as last element of path

			pCount++
			chunk = path[start:]

			if cn.params != nil {
				if cn.params.param != chunk {
					panic("Different param names defined for path '" + origPath + "', param '" + chunk + "'' should be '" + cn.params.param + "'")
				}

				cn = cn.params

				goto END
			}

			// wild already exists! then will conflict
			if cn.wild != nil {
				panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
			}

			cn.params = &node{
				param: chunk,
			}
			cn = cn.params

			goto END
		}

		if c == startByte {

			if path[end+1:] != blank {
				panic("Character after the * symbol is not permitted, path '" + origPath + "'")
			}

			//Check the node for existing star then throw a panic information
			//if any
			if cn.wild != nil {
				panic("Wildcard already set by another path, current path '" + origPath + "' conflicts")
			}

			// param already exists! then will conflict
			if cn.params != nil {
				panic("Cannot add wildcard for path '" + origPath + "', a conflicting param path exists with param '" + cn.params.param + "'")
			}

			cn.wild = &node{}
			cn = cn.wild

			goto END
		}
	}

	chunk = path[start:]

	// if blank we're done move on
	if chunk == blank {
		goto END
	}

	if en, ok = cn.static[chunk]; ok {
		cn = en
		goto END
	}

	if cn.static == nil {
		cn.static = nodes{}
	}

	cn.static[chunk] = &node{}
	cn = cn.static[chunk]

END:

	if pCount > l.mostParams {
		l.mostParams = pCount
	}

	cn.addChain(origPath, method, append(rg.middleware, h...))
}

func (l *LARS) find(ctx *DefaultContext, processEnd bool) {

	cn := l.head
	path := ctx.request.URL.Path[1:]

	var (
		start int
		end   int
		nn    *node
		ok    bool
		i     int
		j     int
	)

	// start parsing URL
	for ; end < len(path); end++ {

		if path[end] != slashByte {
			continue
		}

		j = end + 1

		if nn, ok = cn.static[path[start:j]]; ok {

			if path[j:] == blank {
				if ctx.handlers, ok = nn.chains[ctx.request.Method]; !ok {
					goto PARAMS
				}

				cn = nn

				goto END
			}

			cn = nn
			start = j

			continue
		}

	PARAMS:
		// no matching static chunk look at params if available
		if cn.params != nil {

			if path[j:] == blank {
				if ctx.handlers, ok = cn.params.chains[ctx.request.Method]; !ok {
					goto WILD
				}

				i = len(ctx.params)
				ctx.params = ctx.params[:i+1]
				ctx.params[i].Key = cn.params.param
				ctx.params[i].Value = path[0:end]
				cn = cn.params

				goto END
			}

			// extract param, then continue recursing over nodes.
			i = len(ctx.params)
			ctx.params = ctx.params[:i+1]
			ctx.params[i].Key = cn.params.param
			ctx.params[i].Value = path[0:end]
			cn = cn.params
			start = j

			continue
		}

	WILD:
		// no matching static or param chunk look at wild if available
		if cn.wild != nil {
			ctx.handlers = cn.wild.chains[ctx.request.Method]
			cn = cn.wild
			goto END
		}

		cn = nn

		goto END
	}

	// no slash encountered, end of path...
	if nn, ok = cn.static[path[start:]]; ok {
		ctx.handlers = nn.chains[ctx.request.Method]
		cn = nn

		goto END
	}

	if cn.params != nil {
		ctx.handlers = cn.params.chains[ctx.request.Method]
		i = len(ctx.params)
		ctx.params = ctx.params[:i+1]
		ctx.params[i].Key = cn.params.param
		ctx.params[i].Value = path[start:]
		cn = cn.params

		goto END
	}

	// no matching chunk nor param check if wild
	if cn.wild != nil {
		ctx.handlers = cn.wild.chains[ctx.request.Method]
		cn = cn.wild

		goto END
	}

	if path == blank {
		ctx.handlers = cn.chains[ctx.request.Method]
	}

	cn = nil

END:
	if ctx.handlers == nil && processEnd {
		ctx.params = ctx.params[0:0]

		if l.handleMethodNotAllowed && cn != nil && len(cn.chains) > 0 {
			ctx.Set("methods", cn.chains)
			ctx.handlers = l.http405
			return
		}

		if l.redirectTrailingSlash {

			// find again all lowercase
			lc := strings.ToLower(ctx.request.URL.Path)

			if lc != ctx.request.URL.Path {

				ctx.request.URL.Path = lc
				l.find(ctx, false)

				if ctx.handlers != nil {
					l.redirect(ctx)
					return
				}
			}

			ctx.params = ctx.params[0:0]

			if ctx.request.URL.Path[len(ctx.request.URL.Path)-1:] == basePath {
				ctx.request.URL.Path = ctx.request.URL.Path[:len(ctx.request.URL.Path)-1]
			} else {
				ctx.request.URL.Path = ctx.request.URL.Path + basePath
			}

			// find with lowercase + or - sash
			l.find(ctx, false)
			if ctx.handlers != nil {
				l.redirect(ctx)
				return
			}
		}

		ctx.params = ctx.params[0:0]
		ctx.handlers = append(l.routeGroup.middleware, l.http404...)
	}
}

func (l *LARS) redirect(ctx *DefaultContext) {

	code := http.StatusMovedPermanently

	if ctx.request.Method != GET {
		code = http.StatusTemporaryRedirect
	}

	fn := func(c Context) {
		req := c.Request()
		http.Redirect(c.Response(), req, req.URL.Path, code)
	}

	ctx.handlers = append(l.routeGroup.middleware, fn)
}
