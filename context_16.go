// +build !go1.7

package lars

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
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
	netContext          context.Context
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
	c.netContext = context.Background() // in go 1.7 will call r.Context(), netContext will go away and be replaced with the Request objects Context
	c.index = -1
	c.handlers = nil
	c.formParsed = false
	c.multipartFormParsed = false
}

// Set is used to store a new key/value pair using the
// golang.org/x/net/context contained on this Context.
// It is a shortcut for context.WithValue(..., ...)
func (c *Ctx) Set(key interface{}, value interface{}) {
	c.netContext = context.WithValue(c.netContext, key, value)
}

// Get returns the value for the given key and is a shortcut
// for the golang.org/x/net/context context.Value(...) ... but it
// also returns if the value was found or not.
func (c *Ctx) Get(key interface{}) (value interface{}, exists bool) {
	value = c.netContext.Value(key)
	exists = value != nil
	return
}

// golang.org/x/net/context functions to comply with context.Context interface and keep context update on lars.Context object

// Context returns the request's context. To change the context, use
// WithContext.
//
// The returned context is always non-nil.
func (c *Ctx) Context() context.Context {
	return c.netContext // TODO: in go 1.7 return c.request.Context()
}

// WithContext updates the underlying request's context with to ctx
// The provided ctx must be non-nil.
func (c *Ctx) WithContext(ctx context.Context) {
	c.netContext = ctx // TODO: in go 1.7 must update Request object after calling c.request.WithContext(...)
}

// Deadline calls the underlying golang.org/x/net/context Deadline()
func (c *Ctx) Deadline() (deadline time.Time, ok bool) {
	return c.netContext.Deadline()
}

// Done calls the underlying golang.org/x/net/context Done()
func (c *Ctx) Done() <-chan struct{} {
	return c.netContext.Done()
}

// Err calls the underlying golang.org/x/net/context Err()
func (c *Ctx) Err() error {
	return c.netContext.Err()
}

// Value calls the underlying golang.org/x/net/context Value()
func (c *Ctx) Value(key interface{}) interface{} {
	return c.netContext.Value(key)
}

// WithCancel calls golang.org/x/net/context WithCancel and automatically
// updates context on the containing las.Context object.
func (c *Ctx) WithCancel() (cf context.CancelFunc) {
	c.netContext, cf = context.WithCancel(c.netContext)
	return
}

// WithDeadline calls golang.org/x/net/context WithDeadline and automatically
// updates context on the containing las.Context object.
func (c *Ctx) WithDeadline(deadline time.Time) (cf context.CancelFunc) {
	c.netContext, cf = context.WithDeadline(c.netContext, deadline)
	return
}

// WithTimeout calls golang.org/x/net/context WithTimeout and automatically
// updates context on the containing las.Context object.
func (c *Ctx) WithTimeout(timeout time.Duration) (cf context.CancelFunc) {
	c.netContext, cf = context.WithTimeout(c.netContext, timeout)
	return
}

// WithValue calls golang.org/x/net/context WithValue and automatically
// updates context on the containing las.Context object.
// Can also use Set() function on Context object (Recommended)
func (c *Ctx) WithValue(key interface{}, val interface{}) {
	c.netContext = context.WithValue(c.netContext, key, val)
}
