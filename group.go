package lcars

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
}

// routeGroup struct containing all fields and methods for use.
type routeGroup struct {
	prefix     string
	middleware HandlersChain
	lcars      *LCARS
}

var _ IRouteGroup = &routeGroup{}

func (g *routeGroup) handle(method string, path string, handlers []Handler) {

	chain := make(HandlersChain, len(handlers))

	for i, h := range handlers {
		chain[i] = wrapHandler(h)
	}

	g.lcars.router.add(method, g.prefix+path, g, chain)
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

// Group creates a new sub router with prefix. It inherits all properties from
// the parent. Passing middleware overrides parent middleware.
func (g *routeGroup) Group(prefix string, middleware ...Handler) IRouteGroup {

	rg := &routeGroup{
		prefix: g.prefix + prefix,
		lcars:  g.lcars,
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
