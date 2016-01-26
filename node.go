package lars

type nodes map[string]*node

type chainMethods map[string]HandlersChain

// node
type node struct {

	// Static Children
	static nodes

	// Params Children
	params *node

	// Wildcard Children
	wild *node

	chains chainMethods

	// set only on params node
	param string
}

func (n *node) addChain(method string, chain HandlersChain) {

	if n.chains == nil {
		n.chains = map[string]HandlersChain{}
	}

	if n.chains[method] != nil {
		panic("Duplicate Handler for method " + method)
	}

	n.chains[method] = chain
}
