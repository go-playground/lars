package lars

type nodes map[string]*node

type methodChain struct {
	method      string
	handlerName string
	chain       HandlersChain
}

type chainMethods []methodChain

// node
type node struct {

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

func (m chainMethods) find(method string) (HandlersChain, string) {

	l := len(m)
	for i := 0; i < l; i++ {
		if m[i].method == method {
			return m[i].chain, m[i].handlerName
		}
	}

	return nil, blank
}

func (n *node) addChain(origPath string, method string, chain HandlersChain, handlerName string) {

	if n.chains == nil {
		n.chains = make(chainMethods, 0)
	}

	if c, _ := n.chains.find(method); c != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.chains = append(n.chains, methodChain{method: method, chain: chain, handlerName: handlerName})
}

func (n *node) addSlashChain(origPath, method string, chain HandlersChain, handlerName string) {

	if n.parmsSlashChains == nil {
		n.parmsSlashChains = make(chainMethods, 0)
	}

	if c, _ := n.parmsSlashChains.find(method); c != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.parmsSlashChains = append(n.parmsSlashChains, methodChain{method: method, chain: chain, handlerName: handlerName})
}
