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
	// P(i int) (string, bool)
	Param(name string) (string, bool)
	Params() Params
	Get(key string) (value interface{}, exists bool)
	Set(key string, value interface{})
	Next()
	Reset(w http.ResponseWriter, r *http.Request)
	UnderlyingContext() *ctx
}

type ctx struct {
	context.Context
	request  *http.Request
	response *Response
	params   Params
	handlers HandlersChain
	store    store
	index    int
}

var _ context.Context = &ctx{}
var _ Context = &ctx{}

// NewContext returns a new default lars Context object.
// Particularily useful when creating a custom Context
// but still wanting the default Context behavior
func NewContext(l *LARS) Context {
	return &ctx{
		params:   make(Params, l.mostParams),
		response: &Response{},
	}
}

// Request returns context assotiated *http.Request.
func (c *ctx) UnderlyingContext() *ctx {
	return c
}

// Request returns context assotiated *http.Request.
func (c *ctx) Request() *http.Request {
	return c.request
}

// Response returns http.ResponseWriter.
func (c *ctx) Response() *Response {
	return c.response
}

// params are in reverse order, could add this back but have to do some jiggery pokey
// // P returns path parameter by index.
// func (c *ctx) P(i int) (string, bool) {

// 	l := len(c.params)

// 	if i < l {
// 		return c.params[i].Value, true
// 	}

// 	return blank, false
// }

// Param returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned and false is returned.
func (c *ctx) Param(name string) (string, bool) {

	for _, entry := range c.params {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return blank, false
}

// Params returns the array of parameters within the context
func (c *ctx) Params() Params {
	return c.params
}

func (c *ctx) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	c.params = c.params[0:0]
	c.store = nil
	c.index = -1
	c.handlers = nil
}

// Set is used to store a new key/value pair exclusivelly for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *ctx) Set(key string, value interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *ctx) Get(key string) (value interface{}, exists bool) {
	if c.store != nil {
		value, exists = c.store[key]
	}
	return
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
func (c *ctx) Next() {

	c.index++

	if c.index < len(c.handlers) {
		c.handlers[c.index](c)
	}
}
