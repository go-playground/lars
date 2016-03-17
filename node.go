package lars

import "net/url"

type nodeType uint8

const (
	isStatic nodeType = iota // default
	isRoot
	hasParams
	matchEverything
)

// type nodes map[string]*node

type methodChain struct {
	// method      string
	handlerName string
	chain       HandlersChain
}

type existingParams map[string]struct{}

type node struct {
	path      string
	wildChild bool
	nType     nodeType
	indices   string
	children  []*node
	handler   *methodChain
	priority  uint32
}

func (e existingParams) Check(param string, path string) {

	if _, ok := e[param]; ok {
		panic("Duplicate param name '" + param + "' detected for route '" + path + "'")
	}

	e[param] = struct{}{}
}

// increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {

	n.children[pos].priority++
	prio := n.children[pos].priority

	// adjust position (move to front)
	newPos := pos
	for newPos > 0 && n.children[newPos-1].priority < prio {
		// swap node positions
		tmpN := n.children[newPos-1]
		n.children[newPos-1] = n.children[newPos]
		n.children[newPos] = tmpN

		newPos--
	}

	// build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // unchanged prefix, might be empty
			n.indices[pos:pos+1] + // the index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given handle to the path.
// here we set a Middleware because we have  to transfer all route's middlewares (it's a chain of functions) (with it's handler) to the node
func (n *node) add(path string, handlerName string, handler HandlersChain) (lp uint8) {

	var err error

	if path == "" {
		path = "/"
	}

	existing := make(existingParams)
	fullPath := path

	if path, err = url.QueryUnescape(path); err != nil {
		panic("Query Unescape Error on path '" + fullPath + "': " + err.Error())
	}

	fullPath = path

	n.priority++
	numParams := countParams(path)
	lp = numParams

	// non-empty tree
	if len(n.path) > 0 || len(n.children) > 0 {
	walk:
		for {
			// Update maxParams of the current node
			// if numParams > n.maxParams {
			// 	n.maxParams = numParams
			// }

			// Find the longest common prefix.
			// This also implies that the common prefix contains no : or *
			// since the existing key can't contain those chars.
			i := 0
			max := min(len(path), len(n.path))
			for i < max && path[i] == n.path[i] {
				i++
			}

			// Split edge
			if i < len(n.path) {
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					handler:   n.handler,
					priority:  n.priority - 1,
				}

				// Update maxParams (max of all children)
				// for i := range child.children {
				// 	if child.children[i].maxParams > child.maxParams {
				// 		child.maxParams = child.children[i].maxParams
				// 	}
				// }

				n.children = []*node{&child}
				// []byte for proper unicode char conversion, see #65
				n.indices = string([]byte{n.path[i]})
				n.path = path[:i]
				n.handler = nil
				n.wildChild = false
			}

			// Make new node a child of this node
			if i < len(path) {
				path = path[i:]

				if n.wildChild {
					n = n.children[0]
					n.priority++

					// // Update maxParams of the child node
					// if numParams > n.maxParams {
					// 	n.maxParams = numParams
					// }
					numParams--

					// fmt.Println("PARAM A:", n.path)

					existing.Check(n.path, fullPath)

					// Check if the wildcard matches
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						// check for longer wildcard, e.g. :name and :names
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue walk
						}
					}

					panic("path segment '" + path +
						"' conflicts with existing wildcard '" + n.path +
						"' in path '" + fullPath + "'")

					// badParam := path
					// bpi := strings.Index(path, "/")

					// if bpi != -1 {
					// 	badParam = path[:bpi]
					// }

					// fmt.Println(path)

					// panic("Different param names defined for path '" + fullPath + "', param '" + badParam + "'' should be '" + n.path + "'")
				}

				c := path[0]

				// slash after param
				if n.nType == hasParams && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.priority++
					continue walk
				}

				// Check if a child with the next path byte exists
				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue walk
					}
				}

				// Otherwise insert it
				if c != colonByte && c != wildByte {
					// []byte for proper unicode char conversion, see #65
					n.indices += string([]byte{c})
					child := &node{
					// maxParams: numParams,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.indices) - 1)
					n = child
				}
				n.insertChild(numParams, existing, path, fullPath, handlerName, handler)
				return

			} else if i == len(path) { // Make node a (in-path) leaf
				if n.handler != nil {
					panic("handlers are already registered for path '" + fullPath + "'")
				}
				n.handler = &methodChain{
					handlerName: handlerName,
					chain:       handler,
				}
			}
			return
		}
	} else { // Empty tree
		n.insertChild(numParams, existing, path, fullPath, handlerName, handler)
		n.nType = isRoot
	}

	return
}

func (n *node) insertChild(numParams uint8, existing existingParams, path string, fullPath string, handlerName string, handler HandlersChain) {

	// fmt.Println(path)

	var offset int // already handled bytes of the path

	// find prefix until first wildcard (beginning with colonByte' or wildByte')
	for i, max := 0, len(path); numParams > 0; i++ {
		// fmt.Println(path, numParams, i)
		c := path[i]
		if c != colonByte && c != wildByte {
			continue
		}

		// find wildcard end (either '/' or path end)
		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			// the wildcard name must not contain ':' and '*'
			case colonByte, wildByte:
				panic("only one wildcard per path segment is allowed, has: '" +
					path[i:] + "' in path '" + fullPath + "'")
			default:
				end++
			}
		}

		// check if this Node existing children which would be
		// unreachable if we insert the wildcard here
		if len(n.children) > 0 {
			panic("wildcard route '" + path[i:end] +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		// // check if the wildcard has a name
		// if end-i < 2 {
		// 	panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		// }

		if c == colonByte { // param

			// check if the wildcard has a name
			if end-i < 2 {
				panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
			}

			// split path at the beginning of the wildcard
			if i > 0 {
				// fmt.Println("ADD PARAM 1:", path[offset:i])
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType: hasParams,
				// maxParams: numParams,
			}
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--

			// if the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if end < max {

				existing.Check(path[offset:end], fullPath)
				// fmt.Println("ADD PARAM 2:", path[offset:end])

				n.path = path[offset:end]
				offset = end

				child := &node{
					// maxParams: numParams,
					priority: 1,
				}
				n.children = []*node{child}
				n = child
			}

		} else { // catchAll
			if end != max || numParams > 1 {
				panic("Character after the * symbol is not permitted, path '" + fullPath + "'")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     matchEverything,
				// maxParams: 1,
			}
			n.children = []*node{child}
			n.indices = string(path[i])
			n = child
			n.priority++

			// second node: node holding the variable
			child = &node{
				path:  path[i:],
				nType: matchEverything,
				// maxParams: 1,
				handler:  &methodChain{handlerName: handlerName, chain: handler},
				priority: 1,
			}
			n.children = []*node{child}

			return
		}
	}

	// fmt.Println("PARAM C:", path[offset:], n.nType == hasParams, n.nType)

	if n.nType == hasParams {
		existing.Check(path[offset:], fullPath)
	}

	// insert remaining path part and handle to the leaf
	n.path = path[offset:]
	n.handler = &methodChain{handlerName: handlerName, chain: handler}
}

// Returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (n *node) find(path string, po Params) (handler HandlersChain, p Params, handlerName string) {

	// origPath := path
	p = po

walk: // Outer loop for walking the tree
	for {
		if len(path) > len(n.path) {

			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]

				// If this node does not have a wildcard (param or catchAll)
				// child,  we can just look up the next child node and continue
				// to walk down the tree
				if !n.wildChild {
					c := path[0]
					for i := 0; i < len(n.indices); i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					// tsr = (path == "/" && n.handler != nil)
					return
				}

				// handle wildcard child
				n = n.children[0]
				switch n.nType {
				case hasParams:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// save param value
					// if cap(p) < int(n.maxParams) {
					// 	p = make(PathParameters, 0, n.maxParams)
					// }
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.path[1:]
					p[i].Value = path[:end]

					// we need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						// tsr = (len(path) == end+1)
						return
					}

					if n.handler != nil {
						handler = n.handler.chain
						handlerName = n.handler.handlerName
					}

					if handler != nil {
						return
					} else if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						// tsr = (n.path == "/" && n.handler != nil)
					}

					return

				case matchEverything:
					// save param value
					// if cap(p) < int(n.maxParams) {
					// 	p = make(PathParameters, 0, n.maxParams)
					// }
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					// p[i].Key = n.path[2:]
					p[i].Key = WildcardParam
					p[i].Value = path[1:]

					handler = n.handler.chain
					handlerName = n.handler.handlerName
					return

					// can't happen, but left here in case I'm wrong
					// default:
					// 	panic("invalid node type")
				}
			}

		} else if path == n.path {

			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.

			if n.handler != nil {
				if handler, handlerName = n.handler.chain, n.handler.handlerName; handler != nil {
					return
				}
			}
		}

		// Nothing found
		return
	}
}

// func (n *node) addChain(origPath string, method string, chain HandlersChain, handlerName string) {

// 	if n.chains == nil {
// 		n.chains = make(chainMethods, 0)
// 	}

// 	if c, _ := n.chains.find(method); c != nil {
// 		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
// 	}

// 	n.chains = append(n.chains, methodChain{method: method, chain: chain, handlerName: handlerName})
// }

// func (n *node) addSlashChain(origPath, method string, chain HandlersChain, handlerName string) {

// 	if n.parmsSlashChains == nil {
// 		n.parmsSlashChains = make(chainMethods, 0)
// 	}

// 	if c, _ := n.parmsSlashChains.find(method); c != nil {
// 		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
// 	}

// 	n.parmsSlashChains = append(n.parmsSlashChains, methodChain{method: method, chain: chain, handlerName: handlerName})
// }
