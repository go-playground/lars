package lars

import "log"

// node
type node struct {
	path string

	// Priority is the top number of slashes "/"
	priority int

	// Static Children
	static []*node

	// Params Children
	params *node

	// Wildcard Children
	wild *node

	// handler func of the last node
	chain HandlersChain
}

type router struct {
	lars *LARS
	tree *node
}

func (r *router) addPath(method string, path string, rg *RouteGroup, h HandlersChain) {

	if path[0:1] != "/" {
		panic("Path does not start with SLASH")
	}

	if path == "/" {
		r.tree.chain = append(rg.middleware, h...)
		return
	}

	n := r.add(path[1:], r.tree)
	if n != nil {

	}

	// Set Node from here
}

func (r *router) add(path string, n *node) *node {
	if path == "" {
		return n
	}

	var end int
	var c int32
	for end, c = range path {

		if c == 47 {
			// Static Path here
			// Extract the string
			chunk := path[0 : end+1]
			log.Println(chunk)

			for _, charNode := range n.static {
				if chunk == charNode.path {
					return r.add(path[end+1:], charNode)
				}
			}

			nn := &node{
				path: chunk,
			}

			if n.static == nil {
				n.static = []*node{}
			}

			n.static = append(n.static, nn)
			return r.add(path[end+1:], nn)
		}

		// Check for Wildcard

		// Check for Parameters
	}
	log.Println(path)
	for _, charNode := range n.static {
		if path == charNode.path {
			return n
		}
	}

	nn := &node{
		path: path,
	}

	if n.static == nil {
		n.static = []*node{}
	}

	n.static = append(n.static, nn)

	return n
}

func (r *router) get() {}

func (r *router) sortNode() {}
