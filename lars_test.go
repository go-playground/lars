package lars

import (
	"net/http"
	"net/http/httptest"
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

	Equal(t, code, http.StatusOK)
	Equal(t, body, "808w70")
}

func TestLARSTwoParam(t *testing.T) {
	var p Params

	l := New()
	path := "/github.com/user/:id/:age/"
	l.Get(path, func(c Context) {
		p = c.Params()
	})

	code, _ := request(GET, "/github.com/user/808w70/67/", l)

	Equal(t, code, http.StatusOK)
	Equal(t, p[0].Value, "808w70")
	Equal(t, p[1].Value, "67")
}

func TestRouterMatchAny(t *testing.T) {

	l := New()
	path1 := "/github/"
	path2 := "/github/*"
	path3 := "/users/*"

	l.Get(path1, func(c Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
	})

	l.Get(path2, func(c Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
	})

	l.Get(path3, func(c Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
	})

	code, body := request(GET, "/github/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, path1)

	code, body = request(GET, "/github/department", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/github/department")

	code, body = request(GET, "/users/joe", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/users/joe")

}

func TestRouterMicroParam(t *testing.T) {
	var p Params

	l := New()
	l.Get("/:a/:b/:c", func(c Context) {
		p = c.Params()
	})
	code, _ := request(GET, "/1/2/3", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "1", p[0].Value)
	Equal(t, "2", p[1].Value)
	Equal(t, "3", p[2].Value)

}

func TestRouterMixParamMatchAny(t *testing.T) {
	var p Params

	l := New()

	//Route
	l.Get("/users/:id/*", func(c Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
		p = c.Params()
	})
	code, body := request(GET, "/users/joe/comments", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "joe", p[0].Value)
	Equal(t, "/users/joe/comments", body)
}

func request(method, path string, l *LARS) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	l.serveHTTP(w, r)
	return w.Code, w.Body.String()
}
