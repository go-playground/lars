package lars

import (
	"net/http"
	"net/url"
	"strings"
)

// Router contains the tree information and
// methods to traverse it
type Router struct {
	lars *LARS
	tree *node
}

// newRouter return a router instance for use
func newRouter(l *LARS) *Router {
	return &Router{
		lars: l,
		tree: &node{
			static: nodes{},
		},
	}
}

// Add parses a route and adds it to the tree
func (r *Router) add(method string, path string, rg *routeGroup, h HandlersChain) {

	origPath := path
	cn := r.tree
	existingParams := map[string]struct{}{}

	var (
		start      int
		end        int
		j          int
		c          byte
		en         *node
		ok         bool
		chunk      string
		err        error
		pCount     uint8 = 1
		paramSlash bool
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
			if en = cn.findStatic(chunk); en != nil {
				cn = en
				start = j

				continue
			}

			// no existing node, adding new one
			if cn.static == nil {
				cn.static = nodes{}
			}

			nn := &node{
				path: chunk,
			}

			cn.static = append(cn.static, nn)
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

				if _, ok = existingParams[chunk]; ok {
					panic("Duplicate param name '" + chunk + "' detected for route '" + origPath + "'")
				}

				// existing param node?
				if cn.params != nil {

					// can't have same prefix paths with different param names example:
					// /users/:id/profile
					// /users/:user_id/profile/settings
					// both params above must be either :id or :user_id, no mix & match
					if cn.params.param != chunk {
						panic("Different param names defined for path '" + origPath + "', param '" + chunk + "'' should be '" + cn.params.param + "'")
					}

					existingParams[chunk] = struct{}{}

					pCount++
					cn = cn.params
					start = end + 1

					if path[start:] == blank {
						paramSlash = true
						goto END
					}

					continue MAIN
				}

				// wild already exists! then will conflict
				if cn.wild != nil {
					if cn.wild.chains.find(method) != nil {
						panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
					}
				}

				existingParams[chunk] = struct{}{}

				nn := &node{
					param: chunk,
				}

				cn.params = nn
				pCount++
				cn = nn
				start = end + 1

				if path[start:] == blank {
					paramSlash = true
					goto END
				}

				continue MAIN
			}

			// param name did not end in slash, extract as last element of path

			pCount++
			chunk = path[start:]

			if _, ok = existingParams[chunk]; ok {
				panic("Duplicate param name '" + chunk + "' detected for route '" + origPath + "'")
			}

			if cn.params != nil {
				if cn.params.param != chunk {
					panic("Different param names defined for path '" + origPath + "', param '" + chunk + "'' should be '" + cn.params.param + "'")
				}

				existingParams[chunk] = struct{}{}
				cn = cn.params

				goto END
			}

			// wild already exists! then will conflict
			if cn.wild != nil {
				if cn.wild.chains.find(method) != nil {
					panic("Cannot add url param '" + chunk + "' for path '" + origPath + "', a conflicting wildcard path exists")
				}
			}

			existingParams[chunk] = struct{}{}

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
				if cn.params.chains.find(method) != nil {
					panic("Cannot add wildcard for path '" + origPath + "', a conflicting param path exists with param '" + cn.params.param + "'")
				}
			}

			cn.wild = &node{}
			cn = cn.wild
			pCount++

			goto END
		}
	}

	chunk = path[start:]

	// if blank we're done move on
	if chunk == blank {
		goto END
	}

	if en = cn.findStatic(chunk); en != nil {
		cn = en
		goto END
	}

	if cn.static == nil {
		cn.static = nodes{}
	}

	cn.static = append(cn.static, &node{path: chunk})
	cn = cn.static[len(cn.static)-1]

END:

	if pCount > r.lars.mostParams {
		r.lars.mostParams = pCount
	}

	hndlrs := make(HandlersChain, len(rg.middleware)+len(h))
	copy(hndlrs, rg.middleware)
	copy(hndlrs[len(rg.middleware):], h)

	if paramSlash {
		cn.addSlashChain(origPath, method, hndlrs)
		return
	}

	cn.addChain(origPath, method, hndlrs)
}

// Find attempts to match a given use to a mapped route
// attempting redirect if specified to do so.
func (r *Router) find(ctx *Context, processEnd bool) {

	var (
		start int
		end   int
		nn    *node
		i     int
		j     int
	)

	cn := r.tree
	path := ctx.Request.URL.Path[1:]

	// start parsing URL
	for ; end < len(path); end++ {

		if path[end] != slashByte {
			continue
		}

		j = end + 1

		if nn = cn.findStatic(path[start:j]); nn != nil {

			if path[j:] == blank {
				if ctx.handlers = nn.chains.find(ctx.Request.Method); ctx.handlers == nil {
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
				if ctx.handlers = cn.params.parmsSlashChains.find(ctx.Request.Method); ctx.handlers == nil {
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
			ctx.handlers = cn.wild.chains.find(ctx.Request.Method)
			cn = cn.wild
			i = len(ctx.params)
			ctx.params = ctx.params[:i+1]
			ctx.params[i].Key = WildcardParam
			ctx.params[i].Value = path[start:j]
			goto END
		}

		cn = nn

		goto END
	}

	// no slash encountered, end of path...
	if nn = cn.findStatic(path[start:]); nn != nil {
		if ctx.handlers = nn.chains.find(ctx.Request.Method); ctx.handlers == nil {
			goto PARAMSNOSLASH
		}

		cn = nn

		goto END
	}

PARAMSNOSLASH:
	if cn.params != nil {

		if ctx.handlers = cn.params.chains.find(ctx.Request.Method); ctx.handlers == nil {
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
		ctx.handlers = cn.wild.chains.find(ctx.Request.Method)
		cn = cn.wild
		i = len(ctx.params)
		ctx.params = ctx.params[:i+1]
		ctx.params[i].Key = WildcardParam
		ctx.params[i].Value = path[start:]

		goto END
	}

	if path == blank {
		ctx.handlers = cn.chains.find(ctx.Request.Method)
	}

	cn = nil

END:
	if ctx.handlers == nil && processEnd {
		ctx.params = ctx.params[0:0]

		if r.lars.handleMethodNotAllowed && cn != nil && len(cn.chains) > 0 {
			ctx.Set("methods", cn.chains)
			ctx.handlers = r.lars.http405
			return
		}

		if r.lars.redirectTrailingSlash {

			// find again all lowercase
			lc := strings.ToLower(ctx.Request.URL.Path)

			if lc != ctx.Request.URL.Path {

				ctx.Request.URL.Path = lc
				r.find(ctx, false)

				if ctx.handlers != nil {
					r.redirect(ctx)
					return
				}
			}

			ctx.params = ctx.params[0:0]

			if ctx.Request.URL.Path[len(ctx.Request.URL.Path)-1:] == basePath {
				ctx.Request.URL.Path = ctx.Request.URL.Path[:len(ctx.Request.URL.Path)-1]
			} else {
				ctx.Request.URL.Path = ctx.Request.URL.Path + basePath
			}

			// find with lowercase + or - sash
			r.find(ctx, false)
			if ctx.handlers != nil {
				r.redirect(ctx)
				return
			}
		}

		ctx.handlers = r.lars.notFound
	}
}

// Redirect redirects the current request
func (r *Router) redirect(ctx *Context) {

	code := http.StatusMovedPermanently

	if ctx.Request.Method != GET {
		code = http.StatusTemporaryRedirect
	}

	fn := func(c *Context) {
		http.Redirect(c.Response, c.Request, c.Request.URL.String(), code)
	}

	hndlrs := make(HandlersChain, len(r.lars.routeGroup.middleware)+1)
	copy(hndlrs, r.lars.routeGroup.middleware)
	hndlrs[len(r.lars.routeGroup.middleware)] = fn

	ctx.handlers = hndlrs
}
