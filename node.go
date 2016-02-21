package lars

// type nodes []*node
type nodes []*node

type methodChain struct {
	method string
	chain  HandlersChain
}

// type chainMethods map[string]HandlersChain
type chainMethods []methodChain

// node
type node struct {

	// path to match
	path string

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

	// for i := 0; i < len(n.static); i++ {

	// 	// fmt.Println(n.static[i].path, "=", path)
	// 	found = n.static[i]
	// 	if found.path == path {
	// 		return
	// 	}
	// }

	// found = nil

	for _, sn := range n.static {
		if sn.path == path {
			return sn
		}
	}

	return nil
}

func (m chainMethods) find(method string) HandlersChain {
	for _, mc := range m {
		if mc.method == method {
			return mc.chain
		}
	}

	return nil
}

func (n *node) addChain(origPath string, method string, chain HandlersChain) {

	if n.chains == nil {
		n.chains = make(chainMethods, 0)
		// n.chains = map[string]HandlersChain{}
	}

	if n.chains.find(method) != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.chains = append(n.chains, methodChain{method: method, chain: chain})
	// n.chains[method] = chain
}

func (n *node) addSlashChain(origPath, method string, chain HandlersChain) {

	if n.parmsSlashChains == nil {
		n.parmsSlashChains = make(chainMethods, 0)
		// n.parmsSlashChains = map[string]HandlersChain{}
	}

	if n.parmsSlashChains.find(method) != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.parmsSlashChains = append(n.parmsSlashChains, methodChain{method: method, chain: chain})
}
