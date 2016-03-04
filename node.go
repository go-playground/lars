package lars

type nodes map[string]*node

type methodChain struct {
	method string
	chain  HandlersChain
}

type chainMethods []methodChain

// node
type node struct {

	// path to match
	// path string

	// Static Children
	static nodes

	// Params Children
	params *node

	// Wildcard Children
	wild *node

	chains           chainMethods
	parmsSlashChains chainMethods

	// set only on params node
	param string
}

func (n *node) findStatic(path string) *node {

	return n.static[path]
	// l := len(n.static)
	// for i := 0; i < l; i++ {

	// 	if len(n.static[i].path) != len(path) {
	// 		continue
	// 	}

	// 	if n.static[i].path == path {
	// 		return n.static[i]
	// 	}
	// }

	// return nil
}

func (m chainMethods) find(method string) HandlersChain {

	l := len(m)
	for i := 0; i < l; i++ {
		if m[i].method == method {
			return m[i].chain
		}
	}

	return nil
}

func (n *node) addChain(origPath string, method string, chain HandlersChain) {

	if n.chains == nil {
		n.chains = make(chainMethods, 0)
	}

	if n.chains.find(method) != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.chains = append(n.chains, methodChain{method: method, chain: chain})
}

func (n *node) addSlashChain(origPath, method string, chain HandlersChain) {

	if n.parmsSlashChains == nil {
		n.parmsSlashChains = make(chainMethods, 0)
	}

	if n.parmsSlashChains.find(method) != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.parmsSlashChains = append(n.parmsSlashChains, methodChain{method: method, chain: chain})
}
