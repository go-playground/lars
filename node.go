package lars

// type nodes []*node
type nodes []*node

type chainMethods map[string]HandlersChain

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

func (n *node) findStatic(path string) (found *node) {

	for i := 0; i < len(n.static); i++ {

		// fmt.Println(n.static[i].path, "=", path)
		found = n.static[i]
		if found.path == path {
			return
		}
	}

	found = nil
	return
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
