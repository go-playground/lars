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

	chains           chainMethods
	parmsSlashChains chainMethods
	// set only on params node
	param string
}

func (n *node) addChain(origPath string, method string, chain HandlersChain) {

	if n.chains == nil {
		n.chains = map[string]HandlersChain{}
	}

	if n.chains[method] != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.chains[method] = chain
}

func (n *node) addSlashChain(origPath, method string, chain HandlersChain) {

	if n.parmsSlashChains == nil {
		n.parmsSlashChains = map[string]HandlersChain{}
	}

	if n.parmsSlashChains[method] != nil {
		panic("Duplicate Handler for method '" + method + "' with path '" + origPath + "'")
	}

	n.parmsSlashChains[method] = chain
}
