package lars

import (
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

type store map[string]interface{}

// IGlobals is an interface for a globals http request object that can be passed
// around and allocated efficiently; and most importantly is not tied to the
// context object and can be passed around separately if desired instead of Context
// being the interface, which does not have a clear separation of http Context vs Globals
type IGlobals interface {
	Reset(*Context)
}

// Context encapsulates the http request, response context
type Context struct {
	context.Context
	request  *http.Request
	response *Response
	params   Params
	handlers HandlersChain
	store    store
	index    int
	Globals  IGlobals
}

var _ context.Context = &Context{}

// NewContext returns a new default lars Context object.
func NewContext(l *LARS) *Context {

	return &Context{
		params:   make(Params, l.mostParams),
		response: &Response{},
		Globals:  l.newGlobals(),
	}
}

// Request returns *http.Request of the given context
func (c *Context) Request() *http.Request {
	return c.request
}

// Response returns http.ResponseWriter of the given context
func (c *Context) Response() *Response {
	return c.response
}

// P returns path parameter by index.
func (c *Context) P(i int) (string, bool) {

	l := len(c.params)

	if i < l {
		return c.params[i].Value, true
	}

	return blank, false
}

// Param returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned and false is returned.
func (c *Context) Param(name string) (string, bool) {

	for _, entry := range c.params {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return blank, false
}

// Params returns the array of parameters within the*Context
func (c *Context) Params() Params {
	return c.params
}

// Reset resets the*Context to it's default request state
func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	c.params = c.params[0:0]
	c.store = nil
	c.index = -1
	c.handlers = nil

	if c.Globals != nil {
		c.Globals.Reset(c)
	}
}

// Set is used to store a new key/value pair exclusivelly for this*Context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	if c.store != nil {
		value, exists = c.store[key]
	}
	return
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
func (c *Context) Next() {
	c.index++
	c.handlers[c.index](c)
}

// http request helpers

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func (c *Context) ClientIP() (clientIP string) {

	var values []string

	if values, _ = c.request.Header[XRealIP]; len(values) > 0 {

		clientIP = strings.TrimSpace(values[0])
		if clientIP != blank {
			return
		}
	}

	if values, _ = c.request.Header[XForwardedFor]; len(values) > 0 {
		clientIP = values[0]

		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}

		clientIP = strings.TrimSpace(clientIP)
		if clientIP != blank {
			return
		}
	}

	clientIP, _, _ = net.SplitHostPort(strings.TrimSpace(c.request.RemoteAddr))

	return
}
