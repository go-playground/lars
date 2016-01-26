package lars

import (
	"log"
	"net/http"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
)

// . "gopkg.in/go-playground/assert.v1"

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

var basicHandler = func(Context) {

}

func TestLARS(t *testing.T) {
	l := New()

	l.Get("/", func(c Context) {
		c.Response().Write([]byte("home"))
	})

	code, body := request(GET, "/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "home")
}

func TestLARSStatic(t *testing.T) {
	l := New()
	path := "/github.com/go-experimental/:id"
	l.Get(path, basicHandler)
	code, body := request(GET, "/github.com/go-experimental/808w70", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")
}

func TestLARSParam(t *testing.T) {
	l := New()
	path := "/github.com/go-experimental/:id/"
	l.Get(path, func(c Context) {
		p, _ := c.Param("id")
		c.Response().Write([]byte(p))
	})
	code, body := request(GET, "/github.com/go-experimental/808w70/", l)

	log.Println(code, body)

}
