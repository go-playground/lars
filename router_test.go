package lars

import (
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

func TestDuplicateParams(t *testing.T) {

	l := New()
	l.Get("/store/:id", basicHandler)
	PanicMatches(t, func() { l.Get("/store/:id/employee/:id", basicHandler) }, "Duplicate param name 'id' detected for route '/store/:id/employee/:id'")

	l.Get("/company/:id/", basicHandler)
	PanicMatches(t, func() { l.Get("/company/:id/employee/:id/", basicHandler) }, "Duplicate param name 'id' detected for route '/company/:id/employee/:id/'")
}

func TestWildcardParam(t *testing.T) {
	l := New()
	l.Get("/users/*", func(c Context) {
		c.Response().Write([]byte(c.Param(WildcardParam)))
	})

	code, body := request(GET, "/users/testwild", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "testwild")

	code, body = request(GET, "/users/testwildslash/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "testwildslash/")
}

func TestBadRoutes(t *testing.T) {
	l := New()

	PanicMatches(t, func() { l.Get("/refewrfewf/fefef") }, "No handler mapped to path:/refewrfewf/fefef")
	PanicMatches(t, func() { l.Get("/users//:id", basicHandler) }, "Bad path '/users//:id' contains duplicate // at index:6")
}
