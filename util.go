package lars

import (
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
)

// NativeChainHandler is used in native handler chain middleware
// example using nosurf crsf middleware nosurf.NewPure(lars.NativeChainHandler)
var NativeChainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	c := GetContext(w)
	b := c.BaseContext()

	*b.request = *r

	if b.index+1 < len(b.handlers) {
		c.Next()
	}
})

// GetContext is a helper method for retrieving the Context object from
// the ResponseWriter when using native go hanlders.
// NOTE: this will panic if fed an http.ResponseWriter not provided by lars's
// chaining.
func GetContext(w http.ResponseWriter) Context {
	return w.(*Response).context
}

func detectContentType(filename string) (t string) {
	if t = mime.TypeByExtension(filepath.Ext(filename)); t == "" {
		t = OctetStream
	}
	return
}

// wrapHandler wraps Handler type
func (l *LARS) wrapHandler(h Handler) HandlerFunc {

	switch h := h.(type) {
	case HandlerFunc:
		return h

	case func(Context):
		return h

	case http.Handler, http.HandlerFunc:
		return func(c Context) {

			ctx := c.BaseContext()

			if h.(http.Handler).ServeHTTP(ctx.response, ctx.request); ctx.response.status != http.StatusOK || ctx.response.committed {
				return
			}

			if ctx.index+1 < len(ctx.handlers) {
				c.Next()
			}
		}

	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {

			ctx := c.BaseContext()

			if h(ctx.response, ctx.request); ctx.response.status != http.StatusOK || ctx.response.committed {
				return
			}

			if ctx.index+1 < len(ctx.handlers) {
				c.Next()
			}
		}

	case func(http.ResponseWriter, *http.Request, http.Handler):

		return func(c Context) {
			ctx := c.BaseContext()

			h(ctx.response, ctx.request, NativeChainHandler)
		}

	case func(http.Handler) http.Handler:

		hf := h(NativeChainHandler)

		return func(c Context) {
			ctx := c.BaseContext()

			hf.ServeHTTP(ctx.response, ctx.request)
		}

	default:
		if fn, ok := l.customHandlersFuncs[reflect.TypeOf(h)]; ok {
			return func(c Context) {
				fn(c, h)
			}
		}

		panic("unknown handler")
	}
}

// wrapHandlerWithName wraps Handler type and returns it including it's name before wrapping
func (l *LARS) wrapHandlerWithName(h Handler) (chain HandlerFunc, handlerName string) {

	chain = l.wrapHandler(h)
	handlerName = runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	return
}

func (l *LARS) redirect(method string, to string) (handlers HandlersChain) {

	code := http.StatusMovedPermanently

	if method != GET {
		code = http.StatusPermanentRedirect
	}

	fn := func(c Context) {
		inCtx := c.BaseContext()
		http.Redirect(inCtx.response, inCtx.request, to, code)
	}

	hndlrs := make(HandlersChain, len(l.routeGroup.middleware)+1)
	copy(hndlrs, l.routeGroup.middleware)
	hndlrs[len(l.routeGroup.middleware)] = fn

	handlers = hndlrs
	return
}

func min(a, b int) int {

	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {

	var n uint // add one just as a buffer

	for i := 0; i < len(path); i++ {
		if path[i] == paramByte || path[i] == wildByte {
			n++
		}
	}

	if n >= 255 {
		panic("too many parameters defined in path, max is 255")
	}

	return uint8(n)
}
