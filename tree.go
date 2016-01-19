package lars

import (
	"fmt"
	"log"
)

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

	// set only on params node
	param string
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

		if c == slash {
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

		// Check for Parameters
		if c == colon {

			start := end + 1
			// fmt.Println("LEFT:", path[start:])
			for end, c = range path[start:] {
				if c == slash {

					param := path[start : end+1]

					fmt.Println("Param:", param)

					if n.params != nil {
						if n.params.param != param {
							panic("Different Param names defined")
						}

						r.add(path[end+2:], n.params)
					}

					nn := &node{
						path:  ":",
						param: param,
					}

					n.params = nn

					fmt.Println("PATHH:", path[end+2:])
					return r.add(path[end+2:], nn)
				}
			}

			param := path[start:]

			if n.params != nil {
				if n.params.param != param {
					panic("Different Param names defined")
				}

				return n
			}

			nn := &node{
				path:  ":",
				param: param,
			}

			return nn

		}

		// Check for Wildcard
		if c == star {
			if path[end+1:] != "" {
				panic("Charaecter after the * symbol is not acceptable")
			}

			//Check the node for existing star then throw a panic information
			//if any
			if n.wild != nil {
				panic("Wildcard character already exists")
			}

			nn := &node{
				path: "*",
			}

			return nn

		}
	}

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
