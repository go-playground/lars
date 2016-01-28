package lcars

import "net/http"

// wrapHandler wraps Handler type
func wrapHandler(h Handler) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(*Context):
		return h
	case http.Handler, http.HandlerFunc:
		return func(c *Context) {
			res := c.Response()

			if h.(http.Handler).ServeHTTP(res, c.Request()); res.status != http.StatusOK || res.committed {
				return
			}

			if c.index+1 < len(c.handlers) {
				c.Next()
			}
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c *Context) {
			res := c.Response()

			if h(res, c.Request()); res.status != http.StatusOK || res.committed {
				return
			}

			if c.index+1 < len(c.handlers) {
				c.Next()
			}
		}
	default:
		panic("unknown handler")
	}
}
