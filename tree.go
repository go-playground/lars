package lars

import "net/url"

// type nodes []*node
type nodes map[string]*node

// type chainMethods struct {
// 	connect HandlersChain
// 	delete  HandlersChain
// 	get     HandlersChain
// 	head    HandlersChain
// 	options HandlersChain
// 	patch   HandlersChain
// 	post    HandlersChain
// 	put     HandlersChain
// 	trace   HandlersChain
// }

type chainMethods map[string]HandlersChain

// func (n nodes) Len() int {
// 	return len(n)
// }

// func (n nodes) Less(i, j int) bool {
// 	return n[i].priority > n[j].priority
// }

// func (n nodes) Swap(i, j int) {
// 	n[i], n[j] = n[j], n[i]
// }

// node
type node struct {
	// path string

	// Priority is the top number of slashes "/"
	// priority float64

	// Static Children
	static nodes

	// Params Children
	params *node

	// Wildcard Children
	wild *node

	// handler func of the last node
	// chain HandlersChain

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

	// switch method {
	// case GET:
	// 	n.chains.get = chain
	// case POST:
	// 	n.chains.post = chain
	// case PUT:
	// 	n.chains.put = chain
	// case DELETE:
	// 	n.chains.delete = chain
	// case PATCH:
	// 	n.chains.patch = chain
	// case OPTIONS:
	// 	n.chains.options = chain
	// case HEAD:
	// 	n.chains.head = chain
	// case CONNECT:
	// 	n.chains.connect = chain
	// case TRACE:
	// 	n.chains.trace = chain
	// }
}

func (n *node) getChain(method string) HandlersChain {

	return n.chains[method]
	// switch method {
	// case GET:
	// 	return n.chains.get
	// case POST:
	// 	return n.chains.post
	// case PUT:
	// 	return n.chains.put
	// case DELETE:
	// 	return n.chains.delete
	// case PATCH:
	// 	return n.chains.patch
	// case OPTIONS:
	// 	return n.chains.options
	// case HEAD:
	// 	return n.chains.head
	// case CONNECT:
	// 	return n.chains.connect
	// case TRACE:
	// 	return n.chains.trace
	// default:
	// 	return nil
	// }
}

func (r *router) addPath(method string, path string, rg *RouteGroup, h HandlersChain) {

	var err error

	if path, err = url.QueryUnescape(path); err != nil {
		panic("Query Unescape Error:" + err.Error())
	}

	if path == "" {
		path = "/"
	}

	pCount := new(uint8)

	n := add(path[1:], pCount, r.tree)
	if n == nil {
		panic("node not added!")
	}

	if *pCount+1 > r.lars.mostParams {
		r.lars.mostParams = *pCount + 1
	}

	n.addChain(method, append(rg.middleware, h...))
}

// TODO: Add Warning when a wild is add to the same node as a param or vise-versa
func add(path string, pCount *uint8, n *node) *node {

	// if blank we're done move on
	if path == "" {
		return n
	}

	var end int
	var c int32

	for end, c = range path {

		// found chunk ending in slash?
		if c == slash {

			chunk := path[0 : end+1]
			// log.Println("chunk:", chunk)

			// check for existing node
			if charNode, ok := n.static[chunk]; ok {
				return add(path[end+1:], pCount, charNode)
			}
			// for _, charNode := range n.static {
			// 	if chunk == charNode.path {
			// 		return add(path[end+1:], pCount, charNode)
			// 	}
			// }

			// no existing node, adding new one
			nn := &node{
			// path: chunk,

			}

			if n.static == nil {
				n.static = nodes{}
			}

			n.static[chunk] = nn

			// n.static = append(n.static, nn)
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
					// path:  ":",
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
				// path:  ":",
				param: param,
			}

			n.params = nn

			return nn
		}

		// Check for Wildcard
		if c == star {

			if path[end+1:] != "" {
				panic("Character after the * symbol is not acceptable")
			}

			//Check the node for existing star then throw a panic information
			//if any
			if n.wild != nil {
				panic("Wildcard character already exists")
			}

			nn := &node{
			// path: "*",
			}

			n.wild = nn

			return nn

		}
	}

	// no slash encountered, url musn't end in one so use remaining path
	// fmt.Println("end chunk:", path)

	if charNode, ok := n.static[path]; ok {
		return charNode
	}
	// for _, charNode := range n.static {
	// 	if path == charNode.path {
	// 		return charNode
	// 	}
	// }

	nn := &node{
	// path: path,
	}

	if n.static == nil {
		n.static = nodes{}
	}

	n.static[path] = nn
	// n.static = append(n.static, nn)

	return nn
}

// NOTE: may need to sort not just by number or "/" but also by least # of ":" to allow :one/:b and :two/aaa to cooexist

// func (r *router) sort() {
// 	sortNodes(r.tree)
// }

// func sortNodes(n *node) {

// 	sort.Sort(n.static)

// 	for _, node := range n.static {
// 		sortNodes(node)
// 	}

// 	if n.params != nil {
// 		sortNodes(n.params)
// 	}

// 	if n.wild != nil {
// 		sortNodes(n.wild)
// 	}
// }

func (r *router) find(context *ctx, method string, path string) {

	// homepage, no slash equal to r.tree node
	if path == "" || path == "/" {

		chain := r.tree.getChain(method)

		if chain == nil {
			context.handlers = r.lars.http404
			return
		}

		context.handlers = chain
		return
	}

	findRoute(context, r.tree, method, path[1:])

	// fmt.Println("Handlers Nil?", context.handlers == nil)
	if context.handlers == nil {
		context.handlers = r.lars.http404
	}
}

func findRoute(context *ctx, n *node, method string, path string) {

	var end int
	var c int32

	// start parsing URL
	for end, c = range path {

		// found chunk ending in slash
		if c == slash {

			chunk := path[0 : end+1]

			if node, ok := n.static[chunk]; ok {
				newPath := path[end+1:]
				// fmt.Println("NEW PATH:", newPath)

				if newPath == "" {
					context.handlers = n.getChain(method)
					return
				}

				findRoute(context, node, method, newPath)
				if context.handlers != nil {
					return
				}
			}
			// // find matching static node
			// for _, node := range n.static {

			// 	// fmt.Println("NODEPATH:", node.path)
			// 	if chunk == node.path {

			// 		// fmt.Println("MATCHED:", chunk)
			// 		newPath := path[end+1:]
			// 		// fmt.Println("NEW PATH:", newPath)

			// 		if newPath == "" {
			// 			context.handlers = n.getChain(method)
			// 			return
			// 		}

			// 		findRoute(context, node, method, newPath)
			// 		if context.handlers != nil {
			// 			return
			// 		}
			// 	}
			// }

			// no matching static chunk look at params if available
			if n.params != nil {

				// extract param, then continue recursing over nodes.

				newPath := path[end+1:]

				if newPath == "" {
					context.handlers = n.params.getChain(method)
				} else {
					findRoute(context, n.params, method, newPath)
				}

				if context.handlers != nil {
					i := len(context.params)
					context.params = context.params[:i+1]
					context.params[i].Key = n.params.param
					context.params[i].Value = path[0:end]
					return
				}
			}

			// no matching static or param chunk look at wild if available
			if n.wild != nil {
				context.handlers = n.getChain(method)
				return
			}
		}
	}

	// no slash encountered, end of path...
	if node, ok := n.static[path]; ok {
		context.handlers = node.getChain(method)
		return
	}
	// for _, node := range n.static {
	// 	if path == node.path {
	// 		context.handlers = node.getChain(method)
	// 		return
	// 	}
	// }

	if n.params != nil {
		context.handlers = n.params.getChain(method)
		i := len(context.params)
		// fmt.Println(i)
		context.params = context.params[:i+1]
		context.params[i].Key = n.params.param
		context.params[i].Value = path
		return
	}

	// no matching chunk nor param check if wild
	if n.wild != nil {
		context.handlers = n.wild.getChain(method)
		return
	}
}
