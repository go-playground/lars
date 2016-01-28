package lcars

import (
	"net/http"
	"net/http/httptest"
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

func TestResponse(t *testing.T) {

	l := New()
	w := httptest.NewRecorder()
	r := NewResponse(w, l)

	// SetWriter
	r.SetWriter(w)

	// Assert Write
	Equal(t, w, r.Writer())

	// Assert Header
	NotEqual(t, nil, r.Header())

	// WriteHeader
	r.WriteHeader(http.StatusOK)
	Equal(t, http.StatusOK, r.status)

	Equal(t, true, r.committed)

	// Already committed
	r.WriteHeader(http.StatusTeapot)
	NotEqual(t, http.StatusTeapot, r.Status())

	// Status
	r.status = http.StatusOK
	Equal(t, http.StatusOK, r.Status())

	// Write
	info := "Information"
	_, err := r.Write([]byte(info))
	Equal(t, nil, err)

	// Flush
	r.Flush()

	// Size
	IsEqual(len(info), r.Size())

	// WriteString
	s := "LCARS"
	n, err := r.WriteString(s)
	Equal(t, err, nil)
	Equal(t, n, 5)

	//committed
	Equal(t, true, r.Committed())

	panicStr := "interface conversion: *httptest.ResponseRecorder is not http.Hijacker: missing method Hijack"
	fnPanic := func() {
		r.Hijack()
	}
	PanicMatches(t, fnPanic, panicStr)

	panicStr = "interface conversion: *httptest.ResponseRecorder is not http.CloseNotifier: missing method CloseNotify"
	fnPanic = func() {
		r.CloseNotify()
	}
	PanicMatches(t, fnPanic, panicStr)

	// reset
	r.reset(httptest.NewRecorder())
}
