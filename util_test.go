package lars

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

// LogResponseWritter wraps the standard http.ResponseWritter allowing for more
// verbose logging
type logResponseWritter struct {
	status int
	size   int
	http.ResponseWriter
}

// Status provides an easy way to retrieve the status code
func (w *logResponseWritter) Status() int {
	return w.status
}

// Size provides an easy way to retrieve the response size in bytes
func (w *logResponseWritter) Size() int {
	return w.size
}

// Header returns & satisfies the http.ResponseWriter interface
func (w *logResponseWritter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write satisfies the http.ResponseWriter interface and
// captures data written, in bytes
func (w *logResponseWritter) Write(data []byte) (int, error) {

	written, err := w.ResponseWriter.Write(data)
	w.size += written

	return written, err
}

// WriteHeader satisfies the http.ResponseWriter interface and
// allows us to cach the status code
func (w *logResponseWritter) WriteHeader(statusCode int) {

	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func loggingRecoveryHandler(w http.ResponseWriter, r *http.Request) {
	res := w.(*Response)
	wr := &logResponseWritter{status: 200, ResponseWriter: res.Writer()}
	res.SetWriter(wr)
}

func TestOverridingResponseWriterNative(t *testing.T) {
	l := New()
	l.Use(loggingRecoveryHandler)
	l.Get("/test", func(c Context) {
		c.Response().Write([]byte(fmt.Sprint(reflect.TypeOf(c.Response().ResponseWriter))))
	})

	code, body := request(GET, "/test", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "*lars.logResponseWritter")
}

func TestTooManyParams(t *testing.T) {
	s := "/"

	for i := 0; i < 256; i++ {
		s += ":id" + strconv.Itoa(i)
	}

	l := New()
	PanicMatches(t, func() { l.Get(s, func(c Context) {}) }, "too many parameters defined in path, max is 255")
}
