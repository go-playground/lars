package lars

import (
	"sort"
	"strings"
)

type nodes []*node

func (n nodes) Len() int {
	return len(n)
}

func (n nodes) Less(i, j int) bool {
	return n[i].priority > n[j].priority
}

func (n nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// node
type node struct {
	path string

	// Priority is the top number of slashes "/"
	priority int

	// Static Children
	static nodes

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

	j := strings.Count(path, "/")
	if j > r.tree.priority {
		r.tree.priority = j
	}

	n := r.add(path[1:], r.tree)
	if n == nil {
		panic("node not added!")
	}

	n.chain = append(rg.middleware, h...)
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
			// log.Println("chunk:", chunk)

			// check for existing node
			for _, charNode := range n.static {
				if chunk == charNode.path {
					charNode.priority = n.priority
					return r.add(path[end+1:], charNode)
				}
			}

			nn := &node{
				path:     chunk,
				priority: n.priority,
			}

			if n.static == nil {
				n.static = nodes{}
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

					// fmt.Println("Param:", param)

					if n.params != nil {
						if n.params.param != param {
							panic("Different Param names defined")
						}

						n.params.priority = n.priority
						r.add(path[end+2:], n.params)
					}

					nn := &node{
						path:     ":",
						param:    param,
						priority: n.priority,
					}

					n.params = nn

					// fmt.Println("PATHH:", path[end+2:])
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
				path:     ":",
				param:    param,
				priority: n.priority,
			}

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
				path:     "*",
				priority: n.priority,
			}

			return nn

		}
	}

	// fmt.Println("end chunk:", path)

	for _, charNode := range n.static {
		if path == charNode.path {
			return n
		}
	}

	nn := &node{
		path:     path,
		priority: n.priority,
	}

	if n.static == nil {
		n.static = nodes{}
	}

	n.static = append(n.static, nn)

	return n
}

func (r *router) sort() {
	r.sortNodes(r.tree)
}

func (r *router) sortNodes(n *node) {

	sort.Sort(n.static)

	for _, node := range n.static {
		r.sortNodes(node)
	}

	if n.params != nil {
		r.sortNodes(n.params)
	}

	if n.wild != nil {
		r.sortNodes(n.wild)
	}
}

func (r *router) find() {

}
