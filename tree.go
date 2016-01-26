package lars

import "strings"

type kind uint8

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

func (n *node) addChain(method string, chain HandlersChain) {

	if n.chains == nil {
		n.chains = map[string]HandlersChain{}
	}

	n.chains[method] = chain
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

				// wild already exists! then will conflict
				if n.wild != nil {
					panic("Cannot add url param " + param + ", wildcard already exists on this path")
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

			// wild already exists! then will conflict
			if n.wild != nil {
				panic("Cannot add url param " + param + ", wildcard already exists on this path")
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

			// param already exists! then will conflict
			if n.params != nil {
				panic("Cannot add url wildcard, param already exists on this path")
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

func (r *router) find(context *DefaultContext, method string, path string) {

	cn := r.tree

	var (
		start int
		end   int
		node  *node
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

		if node, ok = cn.static[path[start:j]]; ok {

			if path[j:] == blank {
				if context.handlers, ok = node.chains[method]; !ok {
					goto PARAMS
				}

				goto END
			}

			cn = node
			start = j

			continue
		}

	PARAMS:
		// no matching static chunk look at params if available
		if cn.params != nil {

			if path[j:] == blank {
				if context.handlers, ok = cn.params.chains[method]; !ok {
					goto WILD
				}

				i = len(context.params)
				context.params = context.params[:i+1]
				context.params[i].Key = cn.params.param
				context.params[i].Value = path[0:end]

				goto END
			}

			// extract param, then continue recursing over nodes.
			i = len(context.params)
			context.params = context.params[:i+1]
			context.params[i].Key = cn.params.param
			context.params[i].Value = path[0:end]
			cn = cn.params
			start = j

			continue
		}

	WILD:
		// no matching static or param chunk look at wild if available
		if cn.wild != nil {
			context.handlers = cn.wild.chains[method]
		}

		goto END
	}

	// no slash encountered, end of path...
	if node, ok = cn.static[path[start:]]; ok {
		context.handlers = node.chains[method]

		goto END
	}

	if cn.params != nil {
		context.handlers = cn.params.chains[method]
		i = len(context.params)
		context.params = context.params[:i+1]
		context.params[i].Key = cn.params.param
		context.params[i].Value = path[start:]

		goto END
	}

	// no matching chunk nor param check if wild
	if cn.wild != nil {
		context.handlers = cn.wild.chains[method]

		goto END
	}

	if path == blank {
		context.handlers = cn.chains[method]
	}

END:
	if context.handlers == nil {
		context.params = context.params[0:0]

		if r.lars.FixTrailingSlash {

			// find again all lowercase
			lc := strings.ToLower(path)
			if lc != path {
				r.find(context, method, lc[1:])
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
			r.find(context, method, lc[1:])
			if context.handlers != nil {
				r.redirect(context, method, lc)
				return
			}
		}

		context.params = context.params[0:0]
		context.handlers = append(r.lars.RouteGroup.middleware, r.lars.http404...)
	}
}
