package lars

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

// Context is the context interface type
type Context interface {
	context.Context
	Request() *http.Request
	Response() *Response
	WebSocket() *websocket.Conn
	Param(name string) string
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
	BaseContext() *Ctx
}

// Ctx encapsulates the http request, response context
type Ctx struct {
	netContext          context.Context
	request             *http.Request
	response            *Response
	websocket           *websocket.Conn
	params              Params
	handlers            HandlersChain
	parent              Context
	handlerName         string
	index               int
	formParsed          bool
	multipartFormParsed bool
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

// RequestEnd fires after request completes and just before
// the *Ctx object gets put back into the pool.
// Used to close DB connections and such on a custom context
func (c *Ctx) RequestEnd() {
}

// RequestStart resets the Context to it's default request state
func (c *Ctx) RequestStart(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response.reset(w)
	c.params = c.params[0:0]
	c.netContext = context.Background() // in go 1.7 will call r.Context(), netContext will go away and be replaced with the Request objects Context
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
		c.request.Form.Add(entry.Key, entry.Value)
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
		c.request.Form.Add(entry.Key, entry.Value)
	}

	c.multipartFormParsed = true

	return nil
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

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in github.
func (c *Ctx) Next() {
	c.index++
	c.handlers[c.index](c.parent)
}

// http response helpers

// JSON marshals provided interface + returns JSON + status code
func (c *Ctx) JSON(code int, i interface{}) (err error) {

	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	return c.JSONBytes(code, b)
}

// JSONBytes returns provided JSON response with status code
func (c *Ctx) JSONBytes(code int, b []byte) (err error) {

	c.response.Header().Set(ContentType, ApplicationJSONCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

// JSONP sends a JSONP response with status code and uses `callback` to construct
// the JSONP payload.
func (c *Ctx) JSONP(code int, i interface{}, callback string) (err error) {

	b, e := json.Marshal(i)
	if e != nil {
		err = e
		return
	}

	c.response.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)

	if _, err = c.response.Write([]byte(callback + "(")); err == nil {

		if _, err = c.response.Write(b); err == nil {
			_, err = c.response.Write([]byte(");"))
		}
	}

	return
}

// XML marshals provided interface + returns XML + status code
func (c *Ctx) XML(code int, i interface{}) error {

	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}

	return c.XMLBytes(code, b)
}

// XMLBytes returns provided XML response with status code
func (c *Ctx) XMLBytes(code int, b []byte) (err error) {

	c.response.Header().Set(ContentType, ApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)

	if _, err = c.response.Write([]byte(xml.Header)); err == nil {
		_, err = c.response.Write(b)
	}

	return
}

// Text returns the provided string with status code
func (c *Ctx) Text(code int, s string) error {
	return c.TextBytes(code, []byte(s))
}

// TextBytes returns the provided response with status code
func (c *Ctx) TextBytes(code int, b []byte) (err error) {

	c.response.Header().Set(ContentType, TextPlainCharsetUTF8)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
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
