package lars

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

// find with goto's & recursion

// func (r *router) findRoute(context *ctx, method string, path string) {

// 	cn := r.tree

// 	var (
// 		i    int // parameter counter ( saves checking current length on params array )
// 		nk   kind
// 		nn   *node
// 		next string
// 		l    int
// 	)

// 	for {
// 		if path == blank {
// 			goto END
// 		}

// 		for ; l < len(path) && path[l] != slashByte; l++ {
// 		}
// 		// i++ // to get past the lash

// 		if node, ok = n.static[path[0:l+1]]; ok {

// 			// fmt.Println("STATIC 1")
// 			path = path[end+1:]
// 			n = node

// 			goto START
// 		}

// 		goto END
// 	}

// END:
// 	context.handlers = cn.chains[method]
// 	// return
// }

// find with goto's

func (r *router) findRoute(context *ctx, method string, path string) {

	cn := r.tree

	var end int
	var c byte
	var node *node
	var ok bool
	var chunk string
	var i int
	var search string

START:

	// start parsing URL
	for end = 0; end < len(path); end++ {

		c = path[end]

		if c != slashByte {
			continue
		}

		// found chunk ending in slash
		chunk = path[0 : end+1]

		// fmt.Println("CHUNK:", chunk)

		if node, ok = cn.static[chunk]; ok {

			search = path[end+1:]

			if search == blank {
				if context.handlers, ok = node.chains[method]; !ok {
					goto PARAMS
				}

				return
			}

			// fmt.Println("STATIC 1")
			path = search
			cn = node

			goto START
		}

	PARAMS:
		// no matching static chunk look at params if available
		if cn.params != nil {

			// fmt.Println("PARAMS 1")

			search = path[end+1:]

			if search == blank {
				if context.handlers, ok = cn.params.chains[method]; !ok {
					goto WILD
				}

				i = len(context.params)
				context.params = context.params[:i+1]
				context.params[i].Key = cn.params.param
				context.params[i].Value = path[0:end]

				return
			}

			// extract param, then continue recursing over nodes.
			i = len(context.params)
			context.params = context.params[:i+1]
			context.params[i].Key = cn.params.param
			context.params[i].Value = path[0:end]

			path = search
			cn = cn.params

			goto START
		}

	WILD:
		// no matching static or param chunk look at wild if available
		if cn.wild != nil {

			// fmt.Println("WILD 1")
			context.handlers = cn.wild.chains[method]
			return
		}

		return
	}

	// fmt.Println("PATH:", path)

	// no slash encountered, end of path...
	if node, ok = cn.static[path]; ok {
		// fmt.Println("STATIC 2")
		context.handlers = node.chains[method]
		return
	}

	if cn.params != nil {
		// fmt.Println("PARAMS 2")

		context.handlers = cn.params.chains[method]
		i = len(context.params)
		context.params = context.params[:i+1]
		context.params[i].Key = cn.params.param
		context.params[i].Value = path
		return
	}

	// no matching chunk nor param check if wild
	if cn.wild != nil {

		// fmt.Println("WILD 2")
		context.handlers = cn.wild.chains[method]
		return
	}

	if path == blank {

		// fmt.Println("BLANK")
		context.handlers = cn.chains[method]
		return
	}
}
