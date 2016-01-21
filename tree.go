package lars

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
)

const (
	baseStatic = 3
	baseParam  = 2
	baseWild   = 1
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
	priority float64

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

	var err error

	if path, err = url.QueryUnescape(path); err != nil {
		panic("Query Unescape Error:" + err.Error())
	}

	// need to rethink the initial node, could be "/" or ""

	// if r.tree == nil {
	// 	var p string
	// 	if p
	// 	r.tree = &node{
	// 			path:   "/",
	// 			static: []*node{},
	// 		}
	// }

	// if path[0:1] != "" || path[0:1] != "/" {
	// 	panic("Path does not start with SLASH")
	// }

	// if path == "/" || path == "" {
	// 	r.tree.chain = append(rg.middleware, h...)
	// 	return
	// }

	j := float64(strings.Count(path, "/"))
	if j > r.tree.priority {
		r.tree.priority = j
	}

	if path == "" {
		path = "/"
	}

	pCount := new(uint8)
	idx := new(float64)
	priority := new(float64)
	// *idx +=1

	n := r.add(path[1:], idx, priority, pCount, r.tree)
	if n == nil {
		panic("node not added!")
	}

	if *pCount > r.lars.mostParams {
		r.lars.mostParams = *pCount
	}

	n.chain = append(rg.middleware, h...)
}

func (r *router) add(path string, idx *float64, priority *float64, pCount *uint8, n *node) *node {

	if path == "" {
		return n
	}

	*idx++

	var end int
	var c int32

	for end, c = range path {

		if c == slash {

			*priority += baseStatic / *idx

			// Static Path here
			// Extract the string
			chunk := path[0 : end+1]
			log.Println("chunk:", chunk)

			// check for existing node
			for _, charNode := range n.static {
				if chunk == charNode.path {

					// charNode.priority = n.priority
					nd := r.add(path[end+1:], idx, priority, pCount, charNode)

					if *priority > charNode.priority {
						charNode.priority = *priority
					}

					return nd
				}
			}

			nn := &node{
				path: chunk,
				// priority: n.priority,
			}

			if n.static == nil {
				n.static = nodes{}
			}

			n.static = append(n.static, nn)
			nd := r.add(path[end+1:], idx, priority, pCount, nn)

			if *priority > nn.priority {
				nn.priority = *priority
			}

			return nd
		}

		// Check for Parameters
		if c == colon {

			*priority += baseParam / *idx

			start := end + 1
			// fmt.Println("LEFT:", path[start:])
			for end, c = range path[start:] {
				if c != slash {
					continue
				}

				param := path[start : end+1]

				// fmt.Println("Param:", param)

				if n.params != nil {
					if n.params.param != param {
						panic("Different Param names defined")
					}

					// n.params.priority = n.priority
					*pCount++
					*idx *= 2
					nd := r.add(path[end+2:], idx, priority, pCount, n.params)
					if *priority > n.params.priority {
						n.params.priority = *priority
					}

					return nd
				}

				nn := &node{
					path:  ":",
					param: param,
					// priority: n.priority,
				}

				n.params = nn

				*pCount++
				*idx *= 2
				// *idx << 1
				// fmt.Println("PATHH:", path[end+2:])
				nd := r.add(path[end+2:], idx, priority, pCount, nn)

				if *priority > nn.priority {
					nn.priority = *priority
				}

				return nd
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
				priority: *priority,
			}

			return nn

		}

		// Check for Wildcard
		if c == star {

			*priority += baseWild / *idx

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
				priority: *priority,
			}

			return nn

		}
	}

	*priority += baseStatic / *idx
	// fmt.Println("end chunk:", path)

	for _, charNode := range n.static {
		if path == charNode.path {

			if *priority > charNode.priority {
				charNode.priority = *priority
			}

			return charNode
		}
	}

	nn := &node{
		path:     path,
		priority: *priority,
	}

	if n.static == nil {
		n.static = nodes{}
	}

	n.static = append(n.static, nn)

	return nn
}

// NOTE: may need to sort not just by number or "/" but also by least # of ":" to allow :one/:b and :two/aaa to cooexist

func (r *router) sort() {
	sortNodes(r.tree)
}

func sortNodes(n *node) {

	sort.Sort(n.static)

	for _, node := range n.static {
		sortNodes(node)
	}

	if n.params != nil {
		sortNodes(n.params)
	}

	if n.wild != nil {
		sortNodes(n.wild)
	}
}

func (r *router) find(context *ctx, path string) {

	// context.handlers = r.lars.http404

	// homepage, no slash equal to r.tree node
	if path == "" || path == "/" {

		if r.tree.chain == nil {
			context.handlers = r.lars.http404
			return
		}

		context.handlers = r.tree.chain
		return
	}

	findRoute(context, r.tree, path[1:])

	fmt.Println("Handlers Nil?", context.handlers == nil)
	if context.handlers == nil {
		context.handlers = r.lars.http404
	}
}

func findRoute(context *ctx, n *node, path string) {

	fmt.Println("PATH:", path)
	var end, end2 int
	var c, c2 int32

	for end, c = range path {
		if c == slash {

			chunk := path[0 : end+1]

			for _, node := range n.static {
				fmt.Println("NODEPATH:", node.path)
				if chunk == node.path {
					fmt.Println("MATCHED:", chunk)
					newPath := path[end+1:]
					fmt.Println("NEW PATH:", newPath)
					if newPath == "" {
						context.handlers = node.chain
						return
					}

					findRoute(context, node, newPath)
					if context.handlers != nil {
						return
					}
				}
			}

			// no matching chunk look at params then wild
			if n.params != nil {

				// extract param, then continue recursing over nodes.
				start := end + 1
				p := path[start:]
				for end2, c2 = range p {
					if c2 != slash {
						continue
					}

					newPath := path[end2+1:]

					if newPath == "" {
						context.handlers = n.params.chain
					} else {
						findRoute(context, n, path[end2+1:])
					}

					if context.handlers != nil {
						i := len(context.params)
						context.params = context.params[:i+1]
						context.params[i].Key = n.param
						context.params[i].Value = path[start:end2]
						return
					}
				}

				// no slash encountered, param is last value is param
				context.handlers = n.params.chain
				i := len(context.params)
				context.params = context.params[:i+1]
				context.params[i].Key = n.param
				context.params[i].Value = path[start:end2]
				return
				// if n.params.chain != nil {

				// }
				// findRoute(context, n.params, path[end+1:])
				// 	if context.handlers != nil {
				// 		return
				// 	}
			}

			// no matching chunk nor param check if wild
			if n.wild != nil {
				context.handlers = n.chain
				return
			}

			// fmt.Println("Chunk:", chunk)
		}
	}

	// no slash encountered, end of path...
	for _, node := range n.static {
		if path == node.path {

			fmt.Println("MATCHED:", path, len(node.chain))

			context.handlers = node.chain
			return
			// fmt.Println("MATCHED:", chunk)
			// findRoute(context, node, path[end+1:])
			// if context.handlers != nil {
			// 	return
			// }
		}
	}
}
