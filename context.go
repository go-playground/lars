package lars

import (
	"net/http"

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

// Context is the context interface
type Context interface {
	context.Context
	Request() *http.Request
	Response() *Response
	P(i int) (string, bool)
	Param(name string) (string, bool)
	Params() Params
	Get(key string) (value interface{}, exists bool)
	Set(key string, value interface{})
	Next()
	Reset(w http.ResponseWriter, r *http.Request)
	UnderlyingContext() *DefaultContext
}

// DefaultContext is the default underlying context
type DefaultContext struct {
	context.Context
	request  *http.Request
	response *Response
	params   Params
	handlers HandlersChain
	store    store
	index    int
}

var _ context.Context = &DefaultContext{}
var _ Context = &DefaultContext{}

// NewContext returns a new default lars Context object.
// Particularily useful when creating a custom Context
// but still wanting the default Context behavior
func NewContext(l *LARS) *DefaultContext {

	return &DefaultContext{
		params:   make(Params, l.mostParams),
		response: &Response{},
	}
}

// UnderlyingContext returns the underlying default context
func (c *DefaultContext) UnderlyingContext() *DefaultContext {
	return c
}

// Request returns context assotiated *http.Request.
func (c *DefaultContext) Request() *http.Request {
	return c.request
}

// Response returns http.ResponseWriter.
func (c *DefaultContext) Response() *Response {
	return c.response
}

// P returns path parameter by index.
func (c *DefaultContext) P(i int) (string, bool) {

	l := len(c.params)

	if i < l {
		return c.params[i].Value, true
	}

	return blank, false
}

// Param returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned and false is returned.
func (c *DefaultContext) Param(name string) (string, bool) {

	for _, entry := range c.params {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return blank, false
}

// Params returns the array of parameters within the context
func (c *DefaultContext) Params() Params {
	return c.params
}

// Reset resets the DefaultContext to it's default request state
func (c *DefaultContext) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	c.params = c.params[0:0]
	c.store = nil
	c.index = -1
	c.handlers = nil
}

// Set is used to store a new key/value pair exclusivelly for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *DefaultContext) Set(key string, value interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *DefaultContext) Get(key string) (value interface{}, exists bool) {
	if c.store != nil {
		value, exists = c.store[key]
	}
	return
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
func (c *DefaultContext) Next() {

	c.index++
	c.handlers[c.index](c)
}
