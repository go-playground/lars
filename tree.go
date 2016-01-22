package lars

import "net/url"

type nodes map[string]*node

type chainMethods map[string]HandlersChain

// node
type node struct {

	// Static Children
	static nodes

	// Params Children
	params *node

	// Wildcard Children
	wild *node

	chains chainMethods

	// set only on params node
	param string
}

type router struct {
	lars *LARS
	tree *node
}

func (n *node) addChain(method string, chain HandlersChain) {

	if n.chains == nil {
		n.chains = map[string]HandlersChain{}
	}

	n.chains[method] = chain
}

func (r *router) addPath(method string, path string, rg *RouteGroup, h HandlersChain) {

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

// TODO: Add Warning when a wild is add to the same node as a param or vise-versa
func add(path string, pCount *uint8, n *node) *node {

	// if blank we're done move on
	if path == blank {
		return n
	}

	var end int
	var c int32

	for end, c = range path {

		// found chunk ending in slash?
		if c == slash {

			chunk := path[0 : end+1]

			// check for existing node
			if charNode, ok := n.static[chunk]; ok {
				return add(path[end+1:], pCount, charNode)
			}

			// no existing node, adding new one
			if n.static == nil {
				n.static = nodes{}
			}

			nn := &node{}
			n.static[chunk] = nn

			return add(path[end+1:], pCount, nn)
		}

		// found url parameter?
		if c == colon {

			start := end + 1

			// extract param name
			for end, c = range path[start:] {
				if c != slash {
					continue
				}

				param := path[start : end+1]

				// existing param node?
				if n.params != nil {

					// can't have same prefix paths with different param names example:
					// /users/:id/profile
					// /users/:user_id/profile/settings
					// both params above must be either :id or :user_id, no mix & match
					if n.params.param != param {
						panic("Different Param names defined")
					}

					*pCount++

					return add(path[end+2:], pCount, n.params)
				}

				nn := &node{
					param: param,
				}

				n.params = nn

				*pCount++

				return add(path[end+2:], pCount, nn)
			}

			// param name did not end in slash, extract as last element of path

			*pCount++
			param := path[start:]

			if n.params != nil {
				if n.params.param != param {
					panic("Different Param names defined")
				}

				return n.params
			}

			nn := &node{
				param: param,
			}

			n.params = nn

			return nn
		}

		// Check for Wildcard
		if c == star {

			if path[end+1:] != blank {
				panic("Character after the * symbol is not acceptable")
			}

			//Check the node for existing star then throw a panic information
			//if any
			if n.wild != nil {
				panic("Wildcard character already exists")
			}

			nn := &node{}

			n.wild = nn

			return nn

		}
	}

	// no slash encountered, url musn't end in one so use remaining path

	if charNode, ok := n.static[path]; ok {
		return charNode
	}

	if n.static == nil {
		n.static = nodes{}
	}

	nn := &node{}
	n.static[path] = nn

	return nn
}

func (r *router) find(context *ctx, method string, path string) {

	// homepage, no slash equal to r.tree node
	// if path == basePath {

	// 	chain := r.tree.chains[method]

	// 	if chain == nil {
	// 		context.handlers = append(r.lars.RouteGroup.middleware, r.lars.http404...)
	// 		return
	// 	}

	// 	context.handlers = chain
	// 	return
	// }

	findRoute(context, r.tree, method, path[1:])

	if context.handlers == nil {
		context.params = context.params[0:0]
		context.handlers = append(r.lars.RouteGroup.middleware, r.lars.http404...)
	}
}

func findRoute(context *ctx, n *node, method string, path string) {

	var end int
	var c int32
	var node *node
	var ok bool
	var chunk string

START:

	// start parsing URL
	for end, c = range path {

		if c != slash {
			continue
		}

		// found chunk ending in slash

		chunk = path[0 : end+1]

		if node, ok = n.static[chunk]; ok {

			path = path[end+1:]
			n = node

			goto START
		}

		// no matching static chunk look at params if available
		if n.params != nil {

			// extract param, then continue recursing over nodes.

			i := len(context.params)
			context.params = context.params[:i+1]
			context.params[i].Key = n.params.param
			context.params[i].Value = path[0:end]

			path = path[end+1:]
			n = n.params

			goto START
		}

		// no matching static or param chunk look at wild if available
		if n.wild != nil {
			context.handlers = n.chains[method]
			return
		}
	}

	// no slash encountered, end of path...
	if node, ok = n.static[path]; ok {
		context.handlers = node.chains[method]
		return
	}

	if n.params != nil {
		context.handlers = n.params.chains[method]
		i := len(context.params)
		context.params = context.params[:i+1]
		context.params[i].Key = n.params.param
		context.params[i].Value = path
		return
	}

	// no matching chunk nor param check if wild
	if n.wild != nil {
		context.handlers = n.wild.chains[method]
		return
	}

	if path == blank {
		context.handlers = n.chains[method]
		return
	}
}
