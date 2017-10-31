// +build go1.7

package lars

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Context is the context interface type
type Context interface {
	context.Context
	Request() *http.Request
	Response() *Response
	WebSocket() *websocket.Conn
	Param(name string) string
	QueryParams() url.Values
	ParseForm() error
	ParseMultipartForm(maxMemory int64) error
	Set(key interface{}, value interface{})
	Get(key interface{}) (value interface{}, exists bool)
	Context() context.Context
	WithContext(context.Context)
	WithCancel() context.CancelFunc
	WithDeadline(time.Time) context.CancelFunc
	WithTimeout(time.Duration) context.CancelFunc
	WithValue(key interface{}, val interface{})
	Next()
	RequestStart(w http.ResponseWriter, r *http.Request)
	RequestEnd()
	ClientIP() (clientIP string)
	AcceptedLanguages(lowercase bool) []string
	HandlerName() string
	Stream(step func(w io.Writer) bool)
	JSON(int, interface{}) error
	JSONBytes(int, []byte) error
	JSONP(int, interface{}, string) error
	XML(int, interface{}) error
	XMLBytes(int, []byte) error
	Text(int, string) error
	TextBytes(int, []byte) error
	Attachment(r io.Reader, filename string) (err error)
	Inline(r io.Reader, filename string) (err error)
	Decode(includeFormQueryParams bool, maxMemory int64, v interface{}) (err error)
	BaseContext() *Ctx
}

var _ context.Context = &Ctx{}

// Ctx encapsulates the http request, response context
type Ctx struct {
	request             *http.Request
	response            *Response
	websocket           *websocket.Conn
	params              Params
	queryParams         url.Values
	handlers            HandlersChain
	parent              Context
	handlerName         string
	index               int
	formParsed          bool
	multipartFormParsed bool
}

// RequestStart resets the Context to it's default request state
func (c *Ctx) RequestStart(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	c.params = c.params[0:0]
	c.queryParams = nil
	c.index = -1
	c.handlers = nil
	c.formParsed = false
	c.multipartFormParsed = false
}

// Set is used to store a new key/value pair using the
// context contained on this Ctx.
// It is a shortcut for context.WithValue(..., ...)
func (c *Ctx) Set(key interface{}, value interface{}) {
	c.request = c.request.WithContext(context.WithValue(c.request.Context(), key, value))
}

// Get returns the value for the given key and is a shortcut
// for context.Value(...) ... but it
// also returns if the value was found or not.
func (c *Ctx) Get(key interface{}) (value interface{}, exists bool) {
	value = c.request.Context().Value(key)
	exists = value != nil
	return
}

// context functions to comply with context.Context interface and keep context update on lars.Context object

// Context returns the request's context. To change the context, use
// WithContext.
//
// The returned context is always non-nil.
func (c *Ctx) Context() context.Context {
	return c.request.Context()
}

// WithContext updates the underlying request's context with to ctx
// The provided ctx must be non-nil.
func (c *Ctx) WithContext(ctx context.Context) {
	c.request = c.request.WithContext(ctx)
}

// Deadline calls the underlying context.Deadline()
func (c *Ctx) Deadline() (deadline time.Time, ok bool) {
	return c.request.Context().Deadline()
}

// Done calls the underlying context.Done()
func (c *Ctx) Done() <-chan struct{} {
	return c.request.Context().Done()
}

// Err calls the underlying context.Err()
func (c *Ctx) Err() error {
	return c.request.Context().Err()
}

// Value calls the underlying context.Value()
func (c *Ctx) Value(key interface{}) interface{} {
	return c.request.Context().Value(key)
}

// WithCancel calls context.WithCancel and automatically
// updates context on the containing lars.Ctx object.
func (c *Ctx) WithCancel() context.CancelFunc {
	ctx, cf := context.WithCancel(c.request.Context())
	c.request = c.request.WithContext(ctx)
	return cf
}

// WithDeadline calls context.WithDeadline and automatically
// updates context on the containing lars.Ctx object.
func (c *Ctx) WithDeadline(deadline time.Time) context.CancelFunc {
	ctx, cf := context.WithDeadline(c.request.Context(), deadline)
	c.request = c.request.WithContext(ctx)
	return cf
}

// WithTimeout calls context.WithTimeout and automatically
// updates context on the containing lars.Ctx object.
func (c *Ctx) WithTimeout(timeout time.Duration) context.CancelFunc {
	ctx, cf := context.WithTimeout(c.request.Context(), timeout)
	c.request = c.request.WithContext(ctx) // temporarily shallow copying to avoid problems with external libraries
	return cf
}

// WithValue calls golang.org/x/net/context WithValue and automatically
// updates context on the containing lars.Ctx object.
// Can also use Set() function on Ctx object (Recommended)
func (c *Ctx) WithValue(key interface{}, val interface{}) {
	c.request = c.request.WithContext(context.WithValue(c.request.Context(), key, val)) // temporarily shallow copying to avoid problems with external libraries
}
