package lars

import (
	"net/http"
	"net/http/httptest"

	// . "gopkg.in/go-playground/assert.v1"
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

// func TestMain(m *testing.M) {
// 	flag.Parse()
// 	os.Exit(m.Run())
// }

// func TestFind(t *testing.T) {
// 	l := New()
// 	l.Get("/test/:two/three/", func(c Context) { c.Response().Write([]byte("in three")) })
// 	// l.Get("/test/two/three/", func(c Context) { c.Response().Write([]byte("in three")) })
// 	// l.Get("/test/two/three", func(Context) {})
// 	// l.Get("/test/too%2fthree/four", func(Context) {})

// 	var body string

// 	code, _ := request(GET, "", l)
// 	Equal(t, code, http.StatusNotFound)

// 	code, body = request(GET, "/test/two/three/", l)
// 	Equal(t, code, http.StatusOK)
// 	Equal(t, body, "in three")

// 	// r, _ := http.NewRequest("GET", "", nil)
// 	// w := httptest.NewRecorder()
// 	// l.serveHTTP(w, r)

// 	// fmt.Println(l.router.tree.static[0].param.path)
// 	// fmt.Println(l.router.tree.static[0].params.priority, l.router.tree.static[0].params.static.path)

// 	// l.router.sort()

// 	// for idx, n := range l.router.tree.static[0].static {
// 	// 	fmt.Println(idx, n.priority, n.path)
// 	// }

// 	// l.Get("/github.com/go-experimental/lars3/:blob/master历日本語/⌘/à/:alice/*", func(Context) {})
// }

func request(method, path string, l *LARS) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	l.serveHTTP(w, r)
	return w.Code, w.Body.String()
}
