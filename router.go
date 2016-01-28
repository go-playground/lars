package lcars

import (
	"net/http"
	"net/url"
	"strings"
)

// Router contains the tree information and
// methods to traverse it
type Router struct {
	lcars *LCARS
	tree  *node
}

// NewRouter return a router instance for use
func NewRouter(l *LCARS) *Router {
	return &Router{
		lcars: l,
		tree: &node{
			static: nodes{},
		},
	}
}

// Add parses a route and adds it to the tree
func (r *Router) add(method string, path string, rg *routeGroup, h HandlersChain) {

	origPath := path
	cn := r.tree

	var (
		start  int
		end    int
		j      int
		c      byte
		en     *node
		ok     bool
		chunk  string
		err    error
		pCount uint8 = 1
	)

	if path, err = url.QueryUnescape(path); err != nil {
		panic("Query Unescape Error on path '" + origPath + "': " + err.Error())
	}

	if path == blank {
		path = basePath
	}

	path = path[1:]

MAIN:
	for ; end < len(path); end++ {

		c = path[end]

		if c == slashByte {

			j = end + 1
			chunk = path[start:j]

			// check for existing node
			if en, ok = cn.static[chunk]; ok {
				cn = en
				start = j

				continue
			}

			// no existing node, adding new one
			if cn.static == nil {
				cn.static = nodes{}
			}

			nn := &node{}
			cn.static[chunk] = nn
			cn = nn
			start = j

			continue
		}

		if c == colonByte {
			end++
			start = end

			// extract param name
			for ; end < len(path); end++ {
				if path[end] != slashByte {
					continue
				}

				chunk = path[start:end]

				// existing param node?
				if cn.params != nil {

					// can't have same prefix paths with different param names example:
					// /users/:id/profile
					// /users/:user_id/profile/settings
					// both params above must be either :id or :user_id, no mix & match
					if cn.params.param != chunk {
						panic("Different param names defined for path '" + origPath + "', param '" + chunk + "'' should be '" + cn.params.param + "'")
					}

					pCount++
					cn = cn.params
					start = end + 1 // may be wrong here might be + 2 or plus nothing

					continue MAIN
				}

				// wild already exists! then will conflict
				if cn.wild != nil {
					if _, ok := cn.wild.chains[method]; ok {
						panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
					}
				}

				nn := &node{
					param: chunk,
				}

				cn.params = nn
				pCount++
				cn = nn
				start = end + 1 // may be wrong here might be + 2 or plus nothing

				continue MAIN
			}

			// param name did not end in slash, extract as last element of path

			pCount++
			chunk = path[start:]

			if cn.params != nil {
				if cn.params.param != chunk {
					panic("Different param names defined for path '" + origPath + "', param '" + chunk + "'' should be '" + cn.params.param + "'")
				}

				cn = cn.params

				goto END
			}

			// wild already exists! then will conflict
			if cn.wild != nil {
				if _, ok := cn.wild.chains[method]; ok {
					panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
				}
			}

			cn.params = &node{
				param: chunk,
			}
			cn = cn.params

			goto END
		}

		if c == startByte {

			if path[end+1:] != blank {
				panic("Character after the * symbol is not permitted, path '" + origPath + "'")
			}

			//Check the node for existing star then throw a panic information
			//if any
			if cn.wild != nil {
				panic("Wildcard already set by another path, current path '" + origPath + "' conflicts")
			}

			// param already exists! then will conflict
			if cn.params != nil {
				if _, ok := cn.params.chains[method]; ok {
					panic("Cannot add wildcard for path '" + origPath + "', a conflicting param path exists with param '" + cn.params.param + "'")
				}
			}

			cn.wild = &node{}
			cn = cn.wild

			goto END
		}
	}

	chunk = path[start:]

	// if blank we're done move on
	if chunk == blank {
		goto END
	}

	if en, ok = cn.static[chunk]; ok {
		cn = en
		goto END
	}

	if cn.static == nil {
		cn.static = nodes{}
	}

	cn.static[chunk] = &node{}
	cn = cn.static[chunk]

END:

	if pCount > r.lcars.mostParams {
		r.lcars.mostParams = pCount
	}

	cn.addChain(origPath, method, append(rg.middleware, h...))
}

// Find attempts to match a given use to a mapped route
// attempting redirect if specified to do so.
func (r *Router) find(ctx *Context, processEnd bool) {

	cn := r.tree
	path := ctx.request.URL.Path[1:]

	var (
		start int
		end   int
		nn    *node
		ok    bool
		i     int
		j     int
	)

	// start parsing URL
	for ; end < len(path); end++ {

		if path[end] != slashByte {
			continue
		}

		j = end + 1

		if nn, ok = cn.static[path[start:j]]; ok {

			if path[j:] == blank {
				if ctx.handlers, ok = nn.chains[ctx.request.Method]; !ok {
					goto PARAMS
				}

				cn = nn

				goto END
			}

			cn = nn
			start = j

			continue
		}

	PARAMS:
		// no matching static chunk look at params if available
		if cn.params != nil {

			if path[j:] == blank {
				if ctx.handlers, ok = cn.params.chains[ctx.request.Method]; !ok {
					goto WILD
				}

				i = len(ctx.params)
				ctx.params = ctx.params[:i+1]
				ctx.params[i].Key = cn.params.param
				ctx.params[i].Value = path[start:end]
				cn = cn.params

				goto END
			}

			// extract param, then continue recursing over nodes.
			i = len(ctx.params)
			ctx.params = ctx.params[:i+1]
			ctx.params[i].Key = cn.params.param
			ctx.params[i].Value = path[start:end]
			cn = cn.params
			start = j

			continue
		}

	WILD:
		// no matching static or param chunk look at wild if available
		if cn.wild != nil {
			ctx.handlers = cn.wild.chains[ctx.request.Method]
			cn = cn.wild
			goto END
		}

		cn = nn

		goto END
	}

	// no slash encountered, end of path...
	if nn, ok = cn.static[path[start:]]; ok {
		if ctx.handlers, ok = nn.chains[ctx.request.Method]; !ok {
			goto PARAMSNOSLASH
		}
		// ctx.handlers = nn.chains[ctx.request.Method]
		cn = nn

		goto END
	}

PARAMSNOSLASH:
	if cn.params != nil {
		if ctx.handlers, ok = cn.params.chains[ctx.request.Method]; !ok {
			goto WILDNOSLASH
		}

		i = len(ctx.params)
		ctx.params = ctx.params[:i+1]
		ctx.params[i].Key = cn.params.param
		ctx.params[i].Value = path[start:]
		cn = cn.params

		goto END
	}

WILDNOSLASH:
	// no matching chunk nor param check if wild
	if cn.wild != nil {
		ctx.handlers = cn.wild.chains[ctx.request.Method]
		cn = cn.wild

		goto END
	}

	if path == blank {
		ctx.handlers = cn.chains[ctx.request.Method]
	}

	cn = nil

END:
	// fmt.Println("END:", ctx.handlers)
	if ctx.handlers == nil && processEnd {
		ctx.params = ctx.params[0:0]

		if r.lcars.handleMethodNotAllowed && cn != nil && len(cn.chains) > 0 {
			ctx.Set("methods", cn.chains)
			ctx.handlers = r.lcars.http405
			return
		}

		if r.lcars.redirectTrailingSlash {

			// find again all lowercase
			lc := strings.ToLower(ctx.request.URL.Path)

			if lc != ctx.request.URL.Path {

				ctx.request.URL.Path = lc
				r.find(ctx, false)

				if ctx.handlers != nil {
					r.redirect(ctx)
					return
				}
			}

			ctx.params = ctx.params[0:0]

			if ctx.request.URL.Path[len(ctx.request.URL.Path)-1:] == basePath {
				ctx.request.URL.Path = ctx.request.URL.Path[:len(ctx.request.URL.Path)-1]
			} else {
				ctx.request.URL.Path = ctx.request.URL.Path + basePath
			}

			// find with lowercase + or - sash
			r.find(ctx, false)
			if ctx.handlers != nil {
				r.redirect(ctx)
				return
			}
		}

		ctx.params = ctx.params[0:0]
		ctx.handlers = append(r.lcars.routeGroup.middleware, r.lcars.http404...)
	}
}

// Redirect redirects the current request
func (r *Router) redirect(ctx *Context) {

	code := http.StatusMovedPermanently

	if ctx.request.Method != GET {
		code = http.StatusTemporaryRedirect
	}

	fn := func(c *Context) {
		req := c.Request()
		http.Redirect(c.Response(), req, req.URL.Path, code)
	}

	ctx.handlers = append(r.lcars.routeGroup.middleware, fn)
}
