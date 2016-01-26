package lars

import "net/http"

// wrapHandler wraps handler.
func wrapHandler(h Handler) HandlerFunc {
	switch h := h.(type) {
	case HandlerFunc:
		return h
	case func(Context):
		return h
	case http.Handler, http.HandlerFunc:
		return func(c Context) {
			dc := c.UnderlyingContext()

			if h.(http.Handler).ServeHTTP(dc.response, dc.request); dc.response.status != http.StatusOK || dc.response.committed {
				return
			}

			if dc.index+1 < len(dc.handlers) {
				c.Next(c)
			}
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {
			dc := c.UnderlyingContext()

			if h(dc.response, dc.request); dc.response.status != http.StatusOK || dc.response.committed {
				return
			}

			if dc.index+1 < len(dc.handlers) {
				c.Next(c)
			}
		}
	default:
		panic("unknown handler")
	}
}
