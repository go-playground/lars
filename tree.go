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

	// TODO: add check to ensure not overriding a currently set method!
	n.chains[method] = chain
}

func (r *router) addRoute(path string, pCount *uint8) *node {

	cn := r.tree

	var (
		start int
		end   int
		j     int
		c     byte
		en    *node
		ok    bool
		chunk string
	)

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
						panic("Different Param names defined")
					}

					*pCount++
					cn = cn.params
					start = end + 1 // may be wrong here might be + 2 or plus nothing

					continue MAIN
				}

				// wild already exists! then will conflict
				if cn.wild != nil {
					panic("Cannot add url param " + chunk + ", wildcard already exists on this path")
				}

				nn := &node{
					param: chunk,
				}

				cn.params = nn
				*pCount++
				cn = nn
				start = end + 1 // may be wrong here might be + 2 or plus nothing

				continue MAIN
			}

			// param name did not end in slash, extract as last element of path

			*pCount++
			chunk = path[start:]

			if cn.params != nil {
				if cn.params.param != chunk {
					panic("Different Param names defined")
				}

				return cn.params
			}

			// wild already exists! then will conflict
			if cn.wild != nil {
				panic("Cannot add url param " + chunk + ", wildcard already exists on this path")
			}

			nn := &node{
				param: chunk,
			}

			cn.params = nn

			return nn
		}

		if c == startByte {

			if path[end+1:] != blank {
				panic("Character after the * symbol is not acceptable")
			}

			//Check the node for existing star then throw a panic information
			//if any
			if cn.wild != nil {
				panic("Wildcard character already exists")
			}

			// param already exists! then will conflict
			if cn.params != nil {
				panic("Cannot add url wildcard, param already exists on this path")
			}

			nn := &node{}
			cn.wild = nn

			return nn
		}
	}

	chunk = path[start:]

	// if blank we're done move on
	if chunk == blank {
		return cn
	}

	if en, ok = cn.static[chunk]; ok {
		return en
	}

	if cn.static == nil {
		cn.static = nodes{}
	}

	nn := &node{}
	cn.static[chunk] = nn

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
