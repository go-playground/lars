package lars

import (
	"net/http"
	"net/url"
	"strings"
)

type router struct {
	lars *LARS
	tree *node
}

func (r *router) add(method string, path string, rg *RouteGroup, h HandlersChain) {

	var err error

	if path, err = url.QueryUnescape(path); err != nil {
		panic("Query Unescape Error:" + err.Error())
	}

	if path == blank {
		path = basePath
	}

	pCount := new(uint8)
	*pCount++

	n := add(path[1:], pCount, r.tree)
	if n == nil {
		panic("node not added!")
	}

	if *pCount > r.lars.mostParams {
		r.lars.mostParams = *pCount
	}

	n.addChain(method, append(rg.middleware, h...))
}

func (r *router) find(context *ctx, method string, path string) {

	findRoute(context, r.tree, method, path[1:])

	if context.handlers == nil {
		context.params = context.params[0:0]

		if r.lars.FixTrailingSlash {

			// find again all lowercase
			lc := strings.ToLower(path)
			if lc != path {
				findRoute(context, r.tree, method, lc[1:])
				if context.handlers != nil {
					r.redirect(context, method, lc)
					return
				}
			}

			context.params = context.params[0:0]

			if lc[len(lc)-1:] == basePath {
				lc = lc[:len(lc)-1]
			} else {
				lc = lc + basePath
			}

			// find with lowercase + or - sash
			findRoute(context, r.tree, method, lc[1:])
			if context.handlers != nil {
				r.redirect(context, method, lc)
				return
			}
		}

		context.params = context.params[0:0]
		context.handlers = append(r.lars.RouteGroup.middleware, r.lars.http404...)
	}
}

func (r *router) redirect(context *ctx, method, url string) {

	code := http.StatusMovedPermanently

	if method != GET {
		code = http.StatusTemporaryRedirect
	}

	fn := func(c Context) {
		http.Redirect(c.Response(), c.Request(), url, code)
	}

	context.handlers = append(r.lars.RouteGroup.middleware, fn)
}
