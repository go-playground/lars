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
			res := c.Response()

			if h.(http.Handler).ServeHTTP(res, c.Request()); res.Status() != http.StatusOK || res.Committed() {
				return
			}

			c.Next()
		}
	case func(http.ResponseWriter, *http.Request):
		return func(c Context) {

			res := c.Response()

			if h(res, c.Request()); res.Status() != http.StatusOK || res.Committed() {
				return
			}

			c.Next()
		}
	default:
		panic("unknown handler")
	}
}
