package lars

import (
	"log"
	"reflect"
)

// node
type node struct {
	path string

	// Priority is the top number of slashes "/"

	// Static Children
	staticIndices   []byte
	staticChildNode []*node

	// Params Children
	paramsChild *node

	// Wildcard Children
	wildcardChild *node

	// handler func of the last node
	chain []HandlersChain

	params []string
}

type router struct {
	lars *LARS
}

func (r *router) add(method string, path string, rg *RouteGroup, h ...Handler) {

	log.Println("path: ", path)
	// for i := 0; i < len(path); i++ {
	// 	log.Println(string(path[i]))
	// }

	for _, value := range path {
		log.Println(value, string(value), reflect.TypeOf(value))
	}
}

func (r *router) get() {}

func (r *router) sortNode() {}
