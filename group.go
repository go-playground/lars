package lars

import (
	"net/http"
	"net/url"
	"strings"
)

// IRouteGroup interface for router group
type IRouteGroup interface {
	IRoutes
	Group(prefix string, middleware ...Handler) IRouteGroup
}

// IRoutes interface for routes
type IRoutes interface {
	Use(...Handler)
	Any(string, ...Handler)
	Get(string, ...Handler)
	Post(string, ...Handler)
	Delete(string, ...Handler)
	Patch(string, ...Handler)
	Put(string, ...Handler)
	Options(string, ...Handler)
	Head(string, ...Handler)
	Connect(string, ...Handler)
	Trace(string, ...Handler)
	Add(method string, path string, rg *routeGroup, h HandlersChain)
	Find(ctx *DefaultContext, processEnd bool)
	Redirect(ctx *DefaultContext)
}

// routeGroup struct containing all fields and methods for use.
type routeGroup struct {
	prefix     string
	middleware HandlersChain
	lars       *LARS
}

var _ IRouteGroup = &routeGroup{}

func (g *routeGroup) handle(method string, path string, handlers []Handler) {

	chain := make(HandlersChain, len(handlers))

	for i, h := range handlers {
		chain[i] = wrapHandler(h)
	}

	g.Add(method, g.prefix+path, g, chain)
}

// Use adds a middleware handler to the group middleware chain.
func (g *routeGroup) Use(m ...Handler) {
	for _, h := range m {
		g.middleware = append(g.middleware, wrapHandler(h))
	}
}

// Connect adds a CONNECT route & handler to the router.
func (g *routeGroup) Connect(path string, h ...Handler) {
	g.handle(CONNECT, path, h)
}

// Delete adds a DELETE route & handler to the router.
func (g *routeGroup) Delete(path string, h ...Handler) {
	g.handle(DELETE, path, h)
}

// Get adds a GET route & handler to the router.
func (g *routeGroup) Get(path string, h ...Handler) {
	g.handle(GET, path, h)
}

// Head adds a HEAD route & handler to the router.
func (g *routeGroup) Head(path string, h ...Handler) {
	g.handle(HEAD, path, h)
}

// Options adds an OPTIONS route & handler to the router.
func (g *routeGroup) Options(path string, h ...Handler) {
	g.handle(OPTIONS, path, h)
}

// Patch adds a PATCH route & handler to the router.
func (g *routeGroup) Patch(path string, h ...Handler) {
	g.handle(PATCH, path, h)
}

// Post adds a POST route & handler to the router.
func (g *routeGroup) Post(path string, h ...Handler) {
	g.handle(POST, path, h)
}

// Put adds a PUT route & handler to the router.
func (g *routeGroup) Put(path string, h ...Handler) {
	g.handle(PUT, path, h)
}

// Trace adds a TRACE route & handler to the router.
func (g *routeGroup) Trace(path string, h ...Handler) {
	g.handle(TRACE, path, h)
}

// Any adds a route & handler to the router for all HTTP methods.
func (g *routeGroup) Any(path string, h ...Handler) {
	g.Connect(path, h...)
	g.Delete(path, h...)
	g.Get(path, h...)
	g.Head(path, h...)
	g.Options(path, h...)
	g.Patch(path, h...)
	g.Post(path, h...)
	g.Put(path, h...)
	g.Trace(path, h...)
}

// Match adds a route & handler to the router for multiple HTTP methods provided.
func (g *routeGroup) Match(methods []string, path string, h ...Handler) {
	for _, m := range methods {
		g.handle(m, path, h)
	}
}

func (g *routeGroup) Add(method string, path string, rg *routeGroup, h HandlersChain) {

	origPath := path
	cn := g.lars.head

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
					if _, ok := cn.wild.chains[method]; ok {
						panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
					}
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
				if _, ok := cn.wild.chains[method]; ok {
					panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
				}
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
				if _, ok := cn.params.chains[method]; ok {
					panic("Cannot add wildcard for path '" + origPath + "', a conflicting param path exists with param '" + cn.params.param + "'")
				}
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

	if pCount > g.lars.mostParams {
		g.lars.mostParams = pCount
	}

	cn.addChain(origPath, method, append(rg.middleware, h...))
}

func (g *routeGroup) Find(ctx *DefaultContext, processEnd bool) {

	cn := g.lars.head
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
				ctx.params[i].Value = path[start:end]
				cn = cn.params

				goto END
			}

			// extract param, then continue recursing over nodes.
			i = len(ctx.params)
			ctx.params = ctx.params[:i+1]
			ctx.params[i].Key = cn.params.param
			ctx.params[i].Value = path[start:end]
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
		if ctx.handlers, ok = nn.chains[ctx.request.Method]; !ok {
			goto PARAMSNOSLASH
		}
		// ctx.handlers = nn.chains[ctx.request.Method]
		cn = nn

		goto END
	}

PARAMSNOSLASH:
	if cn.params != nil {
		if ctx.handlers, ok = cn.params.chains[ctx.request.Method]; !ok {
			goto WILDNOSLASH
		}

		i = len(ctx.params)
		ctx.params = ctx.params[:i+1]
		ctx.params[i].Key = cn.params.param
		ctx.params[i].Value = path[start:]
		cn = cn.params

		goto END
	}

WILDNOSLASH:
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
	// fmt.Println("END:", ctx.handlers)
	if ctx.handlers == nil && processEnd {
		ctx.params = ctx.params[0:0]

		if g.lars.handleMethodNotAllowed && cn != nil && len(cn.chains) > 0 {
			ctx.Set("methods", cn.chains)
			ctx.handlers = g.lars.http405
			return
		}

		if g.lars.redirectTrailingSlash {

			// find again all lowercase
			lc := strings.ToLower(ctx.request.URL.Path)

			if lc != ctx.request.URL.Path {

				ctx.request.URL.Path = lc
				g.Find(ctx, false)

				if ctx.handlers != nil {
					g.Redirect(ctx)
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
			g.Find(ctx, false)
			if ctx.handlers != nil {
				g.Redirect(ctx)
				return
			}
		}

		ctx.params = ctx.params[0:0]
		ctx.handlers = append(g.middleware, g.lars.http404...)
	}
}

func (g *routeGroup) Redirect(ctx *DefaultContext) {

	code := http.StatusMovedPermanently

	if ctx.request.Method != GET {
		code = http.StatusTemporaryRedirect
	}

	fn := func(c Context) {
		req := c.Request()
		http.Redirect(c.Response(), req, req.URL.Path, code)
	}

	ctx.handlers = append(g.middleware, fn)
}

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (g *routeGroup) Group(prefix string, middleware ...Handler) IRouteGroup {

	rg := &routeGroup{
		prefix: g.prefix + prefix,
		lars:   g.lars,
	}

	if len(middleware) == 0 {
		rg.middleware = make(HandlersChain, len(g.middleware)+len(middleware))
		copy(rg.middleware, g.middleware)

		return rg
	}

	if middleware[0] == nil {
		rg.middleware = make(HandlersChain, 0)
		return rg
	}

	rg.middleware = make(HandlersChain, len(middleware))

	for i, m := range middleware {
		rg.middleware[i] = wrapHandler(m)
	}

	return rg
}
