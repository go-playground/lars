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

func TestFind(t *testing.T) {
	l := New()

	// fn := []Handler{func(c Context) {
	// 	c.Response().Write([]byte(c.Params()[0].Value))
	// }}

	// for _, r := range githubAPI {
	// 	l.routeGroup.handle(r.method, r.path, fn)
	// }

	// l.Delete("/authorizations/:id", func(c Context) {

	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })
	// l.Get("/test/two/three/", func(c Context) { c.Response().Write([]byte("in three")) })
	// l.Get("/test/two/three", func(Context) {})
	// l.Get("/test/too%2fthree/four", func(Context) {})

	// var body string

	// code, _ := request(GET, "", l)
	// Equal(t, code, http.StatusNotFound)
	//

	// l.Get("/authorizations", func(c Context) {
	// 	// p, _ := c.Param("id")
	// 	// c.Response().Write([]byte(p))
	// })

	// l.Post("/authorizations", func(c Context) {
	// 	// p, _ := c.Param("id")
	// 	// c.Response().Write([]byte(p))
	// })

	// l.Get("/authorizations/:id", func(c Context) {
	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })

	// l.Delete("/authorizations/:id", func(c Context) {
	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })

	// for idx, n := range l.router.tree.static {
	// 	fmt.Println(idx, n.path, n.params == nil, n.chains, n.params.chains)
	// }

	// {"GET", "/authorizations/:id"},
	// {"POST", "/authorizations"},
	// //{"PUT", "/authorizations/clients/:client_id"},
	// //{"PATCH", "/authorizations/:id"},
	// {"DELETE", "/authorizations/:id"},

	// l.Get("/authorizations/:id/test", func(c Context) {
	// 	p, _ := c.Param("id")
	// 	c.Response().Write([]byte(p))
	// })

	// code, body := request(GET, "/authorizations/11/test", l)
	// Equal(t, code, http.StatusOK)
	// Equal(t, body, "11")

	l.Get("/", func(c Context) {
		// p, _ := c.Param("id")
		c.Response().Write([]byte("home"))
	})

	code, body := request(GET, "/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "home")

	l.Get("/authorizations/user/test/", func(c Context) {
		// p, _ := c.Param("id")
		c.Response().Write([]byte("1"))
	})

	l.Get("/authorizations/:id/", func(c Context) {
		// p, _ := c.Param("id")
		c.Response().Write([]byte("2"))
	})

	code, body = request(GET, "/authorizations/user/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "2")

	code, body = request(GET, "/authorizations/user/test/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "1")

	// code, body := request(GET, "/authorizations/11", l)
	// Equal(t, code, http.StatusOK)
	// Equal(t, body, "11")

	// code, _ = request(GET, "/authorizations", l)
	// Equal(t, code, http.StatusOK)

	// code, _ = request(POST, "/authorizations", l)
	// Equal(t, code, http.StatusOK)

	// code, body = request(DELETE, "/authorizations/13", l)
	// Equal(t, code, http.StatusOK)
	// Equal(t, body, "13")

	// r, _ := http.NewRequest("GET", "", nil)
	// w := httptest.NewRecorder()
	// l.serveHTTP(w, r)

	// fmt.Println(l.router.tree.static[0].param.path)
	// fmt.Println(l.router.tree.static[0].params.priority, l.router.tree.static[0].params.static.path)

	// l.router.sort()

	// for idx, n := range l.router.tree.static[0].static {
	// 	fmt.Println(idx, n.priority, n.path)
	// }

	// l.Get("/github.com/go-experimental/lars3/:blob/master历日本語/⌘/à/:alice/*", func(Context) {})
}

func request(method, path string, l *LARS) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	l.serveHTTP(w, r)
	return w.Code, w.Body.String()
}
