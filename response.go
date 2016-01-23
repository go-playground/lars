package lars

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
)

// Response struct contains a context *Response
// object if a custom context is not defined.
type Response struct {
	http.ResponseWriter
	status    int
	size      int64
	committed bool
}

// NewResponse creates a new Response for testing purposes
func NewResponse(w http.ResponseWriter, l *LARS) *Response {
	return &Response{ResponseWriter: w}
}

// SetWriter sets the provided writer as the new *Response http.ResponseWriter
func (r *Response) SetWriter(w http.ResponseWriter) {
	r.ResponseWriter = w
}

// Writer return the *Response's http.ResponseWriter object.
// Usually only used when creating middleware.
func (r *Response) Writer() http.ResponseWriter {
	return r.ResponseWriter
}

// Header returns the header map that will be sent by
// WriteHeader. Changing the header after a call to
// WriteHeader (or Write) has no effect unless the modified
// headers were declared as trailers by setting the
// "Trailer" header before the call to WriteHeader (see example).
// To suppress implicit *Response headers, set their value to nil.
func (r *Response) Header() http.Header {
	return r.ResponseWriter.Header()
}

// WriteHeader sends an HTTP *Response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (r *Response) WriteHeader(code int) {
	if r.committed {
		log.Println("response already committed")
		return
	}
	r.status = code
	r.ResponseWriter.WriteHeader(code)
	r.committed = true
}

// Write writes the data to the connection as part of an HTTP reply.
// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
// before writing the data.  If the Header does not contain a
// Content-Type line, Write adds a Content-Type set to the result of passing
// the initial 512 bytes of written data to DetectContentType.
func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.ResponseWriter.Write(b)
	r.size += int64(n)
	return n, err
}

// WriteString write string to ResponseWriter
func (r *Response) WriteString(s string) (n int, err error) {
	n, err = io.WriteString(r.ResponseWriter, s)
	r.size += int64(n)
	return
}

// Flush wraps response writer's Flush function.
func (r *Response) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

// Hijack wraps response writer's Hijack function.
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify wraps response writer's CloseNotify function.
func (r *Response) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Status returns the *Response's current http status code.
func (r *Response) Status() int {
	return r.status
}

// Size returns the number of bytes written in the *Response
func (r *Response) Size() int64 {
	return r.size
}

// Committed returns whether the *Response header has already been written to
// and if has been commited to this return.
func (r *Response) Committed() bool {
	return r.committed
}

func (r *Response) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.size = 0
	r.status = http.StatusOK
	r.committed = false
}
