package lars

import (
	"io"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
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

// Context is the context interface type
type Context interface {
	context.Context
	Request() *http.Request
	Response() *Response
	WebSocket() *websocket.Conn
	Param(name string) string
	ParseForm() error
	ParseMultipartForm(maxMemory int64) error
	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	Next()
	Reset(w http.ResponseWriter, r *http.Request)
	RequestComplete()
	ClientIP() (clientIP string)
	AcceptedLanguages(lowercase bool) []string
	HandlerName() string
	Stream(step func(w io.Writer) bool)
	Attachment(r io.Reader, filename string) (err error)
	Inline(r io.Reader, filename string) (err error)
	BaseContext() *Ctx
}

// Ctx encapsulates the http request, response context
type Ctx struct {
	context.Context
	request             *http.Request
	response            *Response
	websocket           *websocket.Conn
	params              Params
	handlers            HandlersChain
	handlerName         string
	store               store
	index               int
	formParsed          bool
	multipartFormParsed bool
	parent              Context
}

var _ context.Context = &Ctx{}

// NewContext returns a new default lars Context object.
func NewContext(l *LARS) *Ctx {

	c := &Ctx{
		params: make(Params, l.mostParams),
	}

	c.response = newResponse(nil, c)

	return c
}

// BaseContext returns the underlying context object LARS uses internally.
// used when overriding the context object
func (c *Ctx) BaseContext() *Ctx {
	return c
}

// Request returns context assotiated *http.Request.
func (c *Ctx) Request() *http.Request {
	return c.request
}

// Response returns http.ResponseWriter.
func (c *Ctx) Response() *Response {
	return c.response
}

// WebSocket returns context's assotiated *websocket.Conn.
func (c *Ctx) WebSocket() *websocket.Conn {
	return c.websocket
}

// RequestComplete fires after request completes and just before
// the *Ctx object gets put back into the pool.
// Used to close DB connections and such on a custom context
func (c *Ctx) RequestComplete() {
	// nothing will ever be put here so feel free to override and not call
}

// Reset resets the Context to it's default request state
func (c *Ctx) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	c.params = c.params[0:0]
	c.store = nil
	c.index = -1
	c.handlers = nil
	c.formParsed = false
	c.multipartFormParsed = false
}

// Param returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (c *Ctx) Param(name string) string {

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
func (c *Ctx) ParseForm() error {

	if c.formParsed {
		return nil
	}

	if err := c.request.ParseForm(); err != nil {
		return err
	}

	for _, entry := range c.params {
		c.request.Form[entry.Key] = []string{entry.Value}
	}

	c.formParsed = true

	return nil
}

// ParseMultipartForm calls the underlying http.Request ParseMultipartForm
// but also adds the URL params to the request Form as if they were defined
// as query params i.e. ?id=13&ok=true but does not add the params to the
// http.Request.URL.RawQuery for SEO purposes
func (c *Ctx) ParseMultipartForm(maxMemory int64) error {

	if c.multipartFormParsed {
		return nil
	}

	if err := c.request.ParseMultipartForm(maxMemory); err != nil {
		return err
	}

	for _, entry := range c.params {
		c.request.Form[entry.Key] = []string{entry.Value}
	}

	c.multipartFormParsed = true

	return nil
}

// Set is used to store a new key/value pair exclusivelly for thisContext.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Ctx) Set(key string, value interface{}) {
	if c.store == nil {
		c.store = make(store)
	}
	c.store[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Ctx) Get(key string) (value interface{}, exists bool) {
	if c.store != nil {
		value, exists = c.store[key]
	}
	return
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
func (c *Ctx) Next() {
	c.index++
	c.handlers[c.index](c.parent)
}

// http request helpers

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func (c *Ctx) ClientIP() (clientIP string) {

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

// AcceptedLanguages returns an array of accepted languages denoted by
// the Accept-Language header sent by the browser
// NOTE: some stupid browsers send in locales lowercase when all the rest send it properly
func (c *Ctx) AcceptedLanguages(lowercase bool) []string {

	var accepted string

	if accepted = c.request.Header.Get(AcceptedLanguage); accepted == blank {
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

// HandlerName returns the current Contexts final handler's name
func (c *Ctx) HandlerName() string {
	return c.handlerName
}

// Stream provides HTTP Streaming
func (c *Ctx) Stream(step func(w io.Writer) bool) {
	w := c.response
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

// Attachment is a helper method for returning an attachement file
// to be downloaded, if you with to open inline see function
func (c *Ctx) Attachment(r io.Reader, filename string) (err error) {

	c.response.Header().Set(ContentDisposition, "attachment;filename="+filename)
	c.response.Header().Set(ContentType, detectContentType(filename))
	c.response.WriteHeader(http.StatusOK)

	_, err = io.Copy(c.response, r)

	return
}

// Inline is a helper method for returning a file inline to
// be rendered/opened by the browser
func (c *Ctx) Inline(r io.Reader, filename string) (err error) {

	c.response.Header().Set(ContentDisposition, "inline;filename="+filename)
	c.response.Header().Set(ContentType, detectContentType(filename))
	c.response.WriteHeader(http.StatusOK)

	_, err = io.Copy(c.response, r)

	return
}
