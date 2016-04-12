package lars

import (
	"net/http"
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

func TestAddChain(t *testing.T) {
	l := New()

	l.Get("/home", func(Context) {})

	PanicMatches(t, func() { l.Get("/home", func(Context) {}) }, "handlers are already registered for path '/home'")
}

func TestBadWildcard(t *testing.T) {

	l := New()
	PanicMatches(t, func() { l.Get("/test/:test*test", basicHandler) }, "only one wildcard per path segment is allowed, has: ':test*test' in path '/test/:test*test'")

	l.Get("/users/:id/contact-info/:cid", basicHandler)
	PanicMatches(t, func() { l.Get("/users/:id/*", basicHandler) }, "wildcard route '*' conflicts with existing children in path '/users/:id/*'")
	PanicMatches(t, func() { l.Get("/admin/:/", basicHandler) }, "wildcards must be named with a non-empty name in path '/admin/:/'")
	PanicMatches(t, func() { l.Get("/admin/events*", basicHandler) }, "no / before catch-all in path '/admin/events*'")

	l2 := New()
	l2.Get("/", basicHandler)
	PanicMatches(t, func() { l2.Get("/*", basicHandler) }, "catch-all conflicts with existing handle for the path segment root in path '/*'")

	code, _ := request(GET, "/home", l2)
	Equal(t, code, http.StatusNotFound)

	l3 := New()
	l3.Get("/testers/:id", basicHandler)

	code, _ = request(GET, "/testers/13/test", l3)
	Equal(t, code, http.StatusNotFound)
}

func TestDuplicateParams(t *testing.T) {

	l := New()
	l.Get("/store/:id", basicHandler)
	PanicMatches(t, func() { l.Get("/store/:id/employee/:id", basicHandler) }, "Duplicate param name ':id' detected for route '/store/:id/employee/:id'")

	l.Get("/company/:id/", basicHandler)
	PanicMatches(t, func() { l.Get("/company/:id/employee/:id/", basicHandler) }, "Duplicate param name ':id' detected for route '/company/:id/employee/:id/'")
}

func TestWildcardParam(t *testing.T) {
	l := New()
	l.Get("/users/*", func(c Context) {
		if _, err := c.Response().Write([]byte(c.Param(WildcardParam))); err != nil {
			panic(err)
		}
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
