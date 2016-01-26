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
		p, _ := c.Param("id")
		c.Response().Write([]byte(p))
	})

	code, body = request(GET, "/authorizations/user/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "user")

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

func TestHandlerWrapping(t *testing.T) {
	l := New()

	stdlinHandlerFunc := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(r.URL.Path))
		}
	}

	stdLibRawHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.Path))
	}

	fn := func(c Context) { c.Response().Write([]byte(c.Request().URL.Path)) }

	var hf HandlerFunc

	hf = func(c Context) { c.Response().Write([]byte(c.Request().URL.Path)) }

	l.Get("/built-in-context-handler-func/", hf)
	l.Get("/built-in-context-func/", fn)
	l.Get("/stdlib-context-func/", stdLibRawHandlerFunc)
	l.Get("/stdlib-context-handlerfunc/", stdlinHandlerFunc())

	code, body := request(GET, "/built-in-context-handler-func/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/built-in-context-handler-func/")

	code, body = request(GET, "/built-in-context-func/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/built-in-context-func/")

	code, body = request(GET, "/stdlib-context-func/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/stdlib-context-func/")

	code, body = request(GET, "/stdlib-context-handlerfunc/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/stdlib-context-handlerfunc/")

	// test same as above but already commited

	stdlinHandlerFunc2 := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(r.URL.Path))
			w.WriteHeader(http.StatusOK)
		}
	}

	stdLibRawHandlerFunc2 := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.Path))
		w.WriteHeader(http.StatusOK)
	}

	l.Get("/built-in-context-func2/", fn)
	l.Get("/stdlib-context-func2/", stdLibRawHandlerFunc2)
	l.Get("/stdlib-context-handlerfunc2/", stdlinHandlerFunc2())

	code, body = request(GET, "/built-in-context-func2/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/built-in-context-func2/")

	code, body = request(GET, "/stdlib-context-func2/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/stdlib-context-func2/")

	code, body = request(GET, "/stdlib-context-handlerfunc2/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/stdlib-context-handlerfunc2/")

	// test multiple handlers

	stdlinHandlerFunc3 := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// w.Write([]byte(r.URL.Path))
		}
	}

	stdLibRawHandlerFunc3 := func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte(r.URL.Path))
	}

	l.Get("/stdlib-context-func3/", stdLibRawHandlerFunc3, fn)
	l.Get("/stdlib-context-handlerfunc3/", stdlinHandlerFunc3(), fn)

	code, body = request(GET, "/stdlib-context-func3/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/stdlib-context-func3/")

	code, body = request(GET, "/stdlib-context-handlerfunc3/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "/stdlib-context-handlerfunc3/")

	// test bad/unknown handler

	bad := func() string { return "" }

	PanicMatches(t, func() { l.Get("/bad-handler/", bad) }, "unknown handler")
}

type myCustomContext struct {
	*DefaultContext
	text string
}

func (c *myCustomContext) test() string {
	return c.text
}

func (c *myCustomContext) Reset(w http.ResponseWriter, r *http.Request) {
	c.text = "I AM HERE"
	c.DefaultContext.Reset(w, r)
}

func TestCustomContext(t *testing.T) {

	var l *LARS

	fn := func() Context {
		return &myCustomContext{
			DefaultContext: NewContext(l),
		}
	}

	l = New()
	l.RegisterContext(fn)

	l.Get("/home/", func(c Context) {
		ctx := c.(*myCustomContext)
		c.Response().Write([]byte(ctx.text))
	})

	code, body := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "I AM HERE")
}

func TestCustom404(t *testing.T) {

	fn := func(c Context) {
		http.Error(c.Response(), "My Custom 404 Handler", http.StatusNotFound)
	}

	l := New()
	l.Register404(fn)

	code, body := request(GET, "/nonexistantpath", l)
	Equal(t, code, http.StatusNotFound)
	Equal(t, body, "My Custom 404 Handler\n")
}

func TestMethodNotAllowed(t *testing.T) {
	l := New()
	l.SetHandle405MethodNotAllowed(true)

	l.Get("/home/", basicHandler)
	l.Head("/home/", basicHandler)

	code, _ := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(POST, "/home/", nil)
	w := httptest.NewRecorder()
	l.serveHTTP(w, r)

	Equal(t, w.Code, http.StatusMethodNotAllowed)

	allow, ok := w.Header()["Allow"]

	// Sometimes this array is out of order for whatever reason?
	if allow[0] == GET {
		Equal(t, ok, true)
		Equal(t, allow[0], GET)
		Equal(t, allow[1], HEAD)
	} else {
		Equal(t, ok, true)
		Equal(t, allow[1], GET)
		Equal(t, allow[0], HEAD)
	}

	l.SetHandle405MethodNotAllowed(false)

	code, _ = request(POST, "/home/", l)
	Equal(t, code, http.StatusNotFound)
}

func TestRedirect(t *testing.T) {
	l := New()

	l.Get("/home/", basicHandler)
	l.Post("/home/", basicHandler)

	code, _ := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)

	code, _ = request(POST, "/home/", l)
	Equal(t, code, http.StatusOK)

	code, _ = request(GET, "/home", l)
	Equal(t, code, http.StatusMovedPermanently)

	code, _ = request(GET, "/Home/", l)
	Equal(t, code, http.StatusMovedPermanently)

	code, _ = request(POST, "/home", l)
	Equal(t, code, http.StatusTemporaryRedirect)
}

func request(method, path string, l *LARS) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	l.serveHTTP(w, r)
	return w.Code, w.Body.String()
}
