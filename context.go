package lars

import (
	"io"
	"net"
	"net/http"
	"reflect"
	"runtime"
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

// IAppContext is an interface for an AppContext http request object that can be passed
// around and allocated efficiently; and most importantly is not tied to the
// context object and can be passed around separately if desired instead of Context
// being the interface, which does not have a clear separation of http Context vs App Context
type IAppContext interface {
	Reset(*Context)
	Done()
}

// Context encapsulates the http request, response context
type Context struct {
	context.Context
	Request             *http.Request
	Response            *Response
	AppContext          IAppContext
	params              Params
	handlers            HandlersChain
	store               store
	index               int
	formParsed          bool
	multipartFormParsed bool
}

var _ context.Context = &Context{}

// newContext returns a new default lars Context object.
func newContext(l *LARS) *Context {

	c := &Context{
		params:     make(Params, l.mostParams),
		AppContext: l.newAppContext(),
	}

	c.Response = newResponse(nil, c)

	return c
}

// reset resets the Context to it's default request state
func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Response.reset(w)
	c.params = c.params[0:0]
	c.store = nil
	c.index = -1
	c.handlers = nil
	c.formParsed = false
	c.multipartFormParsed = false
}

// Param returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (c *Context) Param(name string) string {

	for _, entry := range c.params {
		if entry.Key == name {
			return entry.Value
		}
	}

	return blank
}

// ParseForm calls the underlying http.Request ParseForm
// but also adds the URL params to the request Form as if
// they were defined as query params i.e. ?id=13&ok=true but
// does not add the params to the http.Request.URL.RawQuery
// for SEO purposes
func (c *Context) ParseForm() error {

	if c.formParsed {
		return nil
	}

	if err := c.Request.ParseForm(); err != nil {
		return err
	}

	for _, entry := range c.params {
		c.Request.Form[entry.Key] = []string{entry.Value}
	}

	c.formParsed = true

	return nil
}

// ParseMultipartForm calls the underlying http.Request ParseMultipartForm
// but also adds the URL params to the request Form as if they were defined
// as query params i.e. ?id=13&ok=true but does not add the params to the
// http.Request.URL.RawQuery for SEO purposes
func (c *Context) ParseMultipartForm(maxMemory int64) error {

	if c.multipartFormParsed {
		return nil
	}

	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		return err
	}

	for _, entry := range c.params {
		c.Request.Form[entry.Key] = []string{entry.Value}
	}

	c.multipartFormParsed = true

	return nil
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

	if values, _ = c.Request.Header[XRealIP]; len(values) > 0 {

		clientIP = strings.TrimSpace(values[0])
		if clientIP != blank {
			return
		}
	}

	if values, _ = c.Request.Header[XForwardedFor]; len(values) > 0 {
		clientIP = values[0]

		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}

		clientIP = strings.TrimSpace(clientIP)
		if clientIP != blank {
			return
		}
	}

	clientIP, _, _ = net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))

	return
}

// AcceptedLanguages returns an array of accepted languages denoted by
// the Accept-Language header sent by the browser
// NOTE: some stupid browsers send in locales lowercase when all the rest send it properly
func (c *Context) AcceptedLanguages(lowercase bool) []string {

	var accepted string

	if accepted = c.Request.Header.Get(AcceptedLanguage); accepted == blank {
		return []string{}
	}

	options := strings.Split(accepted, ",")
	l := len(options)

	language := make([]string, l)

	if lowercase {

		for i := 0; i < l; i++ {
			locale := strings.SplitN(options[i], ";", 2)
			language[i] = strings.ToLower(strings.Trim(locale[0], " "))
		}
	} else {

		for i := 0; i < l; i++ {
			locale := strings.SplitN(options[i], ";", 2)
			language[i] = strings.Trim(locale[0], " ")
		}
	}

	return language
}

// HandlerName returns the current Contexts final handler name
// NOTE: this only works for lars HandlerFunc i.e. func(*Context)
// as native middleware functions are wrapped
func (c *Context) HandlerName() string {

	if c.handlers == nil || len(c.handlers) == 0 {
		return blank
	}

	handler := c.handlers[len(c.handlers)-1]

	return runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
}

// Stream provides HTTP Streaming
func (c *Context) Stream(step func(w io.Writer) bool) {
	w := c.Response
	clientGone := w.CloseNotify()

	for {
		select {
		case <-clientGone:
			return
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				return
			}
		}
	}
}
