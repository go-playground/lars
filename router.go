package lars

import "net/url"

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
		context.handlers = append(r.lars.RouteGroup.middleware, r.lars.http404...)
	}
}
