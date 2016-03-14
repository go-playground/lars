package lars

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
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

var basicHandler = func(Context) {}

func TestFindOneOffs(t *testing.T) {
	fn := func(c Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	l := New()
	l.Get("/users/:id", fn)
	l.Post("/users/*", fn)

	code, body := request(GET, "/users/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(POST, "/users/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, POST)

	l.Get("/admins/:id/", fn)
	l.Post("/admins/*", fn)

	code, body = request(GET, "/admins/1/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(POST, "/admins/1/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, POST)

	l.Post("/superheroes/thor", fn)
	l.Get("/superheroes/:name", fn)

	code, body = request(GET, "/superheroes/thor", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	l.Get("/zombies/:id/profile/", fn)
	l.Get("/zombies/:id/", fn)

	code, body = request(GET, "/zombies/10/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(GET, "/zombies/10", l)
	Equal(t, code, http.StatusMovedPermanently)
	Equal(t, body, "<a href=\"/zombies/10/\">Moved Permanently</a>.\n\n")

	PanicMatches(t, func() { l.Get("/zombies/:id/") }, "Duplicate Handler for method 'GET' with path '/zombies/:id/'")
}

func Testlars(t *testing.T) {
	l := New()

	l.Get("/", func(c Context) {
		c.Response().Write([]byte("home"))
	})

	code, body := request(GET, "/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "home")
}

func TestlarsStatic(t *testing.T) {
	l := New()
	path := "/github.com/go-playground/:id"
	l.Get(path, basicHandler)
	code, body := request(GET, "/github.com/go-playground/808w70", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")
}

func TestlarsParam(t *testing.T) {
	l := New()
	path := "/github.com/go-playground/:id/"
	l.Get(path, func(c Context) {
		p := c.Param("id")
		c.Response().Write([]byte(p))
	})
	code, body := request(GET, "/github.com/go-playground/808w70/", l)

	Equal(t, code, http.StatusOK)
	Equal(t, body, "808w70")
}

func TestlarsTwoParam(t *testing.T) {
	var p1 string
	var p2 string

	l := New()
	path := "/github.com/user/:id/:age/"
	l.Get(path, func(c Context) {
		p1 = c.Param("id")
		p2 = c.Param("age")
	})

	code, _ := request(GET, "/github.com/user/808w70/67/", l)

	Equal(t, code, http.StatusOK)
	Equal(t, p1, "808w70")
	Equal(t, p2, "67")
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
	var context Context

	l := New()
	l.Get("/:a/:b/:c", func(c Context) {
		context = c
	})

	code, _ := request(GET, "/1/2/3", l)
	Equal(t, code, http.StatusOK)

	value := context.Param("a")
	NotEqual(t, len(value), 0)
	Equal(t, "1", value)

	value = context.Param("b")
	NotEqual(t, len(value), 0)
	Equal(t, "2", value)

	value = context.Param("c")
	NotEqual(t, len(value), 0)
	Equal(t, "3", value)

	value = context.Param("key")
	Equal(t, len(value), 0)
	Equal(t, "", value)

}

func TestRouterMixParamMatchAny(t *testing.T) {
	var p string

	l := New()

	//Route
	l.Get("/users/:id/*", func(c Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
		p = c.Param("id")
	})
	code, body := request(GET, "/users/joe/comments", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "joe", p)
	Equal(t, "/users/joe/comments", body)
}

func TestRouterMultiRoute(t *testing.T) {
	var p string
	var parameter string

	l := New()
	//Route
	l.Get("/users", func(c Context) {
		c.Set("path", "/users")
		value, ok := c.Get("path")
		if ok {
			p = value.(string)
		}
	})

	l.Get("/users/:id", func(c Context) {
		parameter = c.Param("id")
	})
	// Route > /users
	code, _ := request(GET, "/users", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "/users", p)
	// Route > /users/:id
	code, _ = request(GET, "/users/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "1", parameter)

	// Route > /user/1
	code, _ = request(GET, "/user/1", l)
	Equal(t, http.StatusNotFound, code)
}

func TestRouterParamNames(t *testing.T) {
	var getP string
	var p1 string
	var p2 string

	l := New()
	//Routes
	l.Get("/users", func(c Context) {
		c.Set("path", "/users")
		value, ok := c.Get("path")
		if ok {
			getP = value.(string)
		}
	})

	l.Get("/users/:id", func(c Context) {
		p1 = c.Param("id")
	})

	l.Get("/users/:id/files/:fid", func(c Context) {
		p1 = c.Param("id")
		p2 = c.Param("fid")
	})

	// Route > users
	code, _ := request(GET, "/users", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "/users", getP)

	// Route >/users/:id
	code, _ = request(GET, "/users/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "1", p1)

	// Route > /users/:uid/files/:fid
	code, _ = request(GET, "/users/1/files/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "1", p1)
	Equal(t, "13", p2)
}

func TestRouterAPI(t *testing.T) {
	l := New()

	for _, route := range githubAPI {
		l.handle(route.method, route.path, []Handler{func(c Context) {
			c.Response().Write([]byte(c.Request().URL.Path))
		}})
	}

	for _, route := range githubAPI {
		code, body := request(route.method, route.path, l)
		Equal(t, body, route.path)
		Equal(t, code, http.StatusOK)
	}
}

func TestUseAndGroup(t *testing.T) {
	fn := func(c Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	var log string

	logger := func(c Context) {
		log = c.Request().URL.Path
		c.Next()
	}

	l := New()
	l.Use(logger)
	l.Get("/", fn)

	code, body := request(GET, "/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/")

	g := l.Group("/users")
	g.Get("/", fn)
	g.Get("/list/", fn)

	code, body = request(GET, "/users/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/users/")

	code, body = request(GET, "/users/list/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/users/list/")

	logger2 := func(c Context) {
		log = c.Request().URL.Path + "2"
		c.Next()
	}

	sh := l.Group("/superheros", logger2)
	sh.Get("/", fn)
	sh.Get("/list/", fn)

	code, body = request(GET, "/superheros/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/superheros/2")

	code, body = request(GET, "/superheros/list/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/superheros/list/2")

	sc := sh.Group("/children")
	sc.Get("/", fn)
	sc.Get("/list/", fn)

	code, body = request(GET, "/superheros/children/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/superheros/children/2")

	code, body = request(GET, "/superheros/children/list/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "/superheros/children/list/2")

	log = ""

	g2 := l.Group("/admins", nil)
	g2.Get("/", fn)
	g2.Get("/list/", fn)

	code, body = request(GET, "/admins/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "")

	code, body = request(GET, "/admins/list/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)
	Equal(t, log, "")
}

func TestBadAdd(t *testing.T) {
	fn := func(c Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	l := New()
	PanicMatches(t, func() { l.Get("/%%%2frs#@$/", fn) }, "Query Unescape Error on path '/%%%2frs#@$/': invalid URL escape \"%%%\"")

	// bad existing params

	l.Get("/user/:id", fn)
	PanicMatches(t, func() { l.Get("/user/:user_id/profile", fn) }, "Different param names defined for path '/user/:user_id/profile', param 'user_id'' should be 'id'")
	l.Get("/user/:id/profile", fn)

	l.Get("/admin/:id/profile", fn)
	PanicMatches(t, func() { l.Get("/admin/:admin_id", fn) }, "Different param names defined for path '/admin/:admin_id', param 'admin_id'' should be 'id'")

	PanicMatches(t, func() { l.Get("/assets/*/test", fn) }, "Character after the * symbol is not permitted, path '/assets/*/test'")

	l.Get("/superhero/*", fn)
	PanicMatches(t, func() { l.Get("/superhero/:id", fn) }, "Cannot add url param 'id' for path '/superhero/:id', a conflicting wildcard path exists")
	PanicMatches(t, func() { l.Get("/superhero/*", fn) }, "Wildcard already set by another path, current path '/superhero/*' conflicts")
	PanicMatches(t, func() { l.Get("/superhero/:id/", fn) }, "Cannot add url param 'id' for path '/superhero/:id/', a conflicting wildcard path exists")

	l.Get("/supervillain/:id", fn)
	PanicMatches(t, func() { l.Get("/supervillain/*", fn) }, "Cannot add wildcard for path '/supervillain/*', a conflicting param path exists with param 'id'")
	PanicMatches(t, func() { l.Get("/supervillain/:id", fn) }, "Duplicate Handler for method 'GET' with path '/supervillain/:id'")
}

func TestAddAllMethods(t *testing.T) {
	fn := func(c Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	l := New()

	l.Get("", fn)
	l.Get("/home/", fn)
	l.Post("/home/", fn)
	l.Put("/home/", fn)
	l.Delete("/home/", fn)
	l.Head("/home/", fn)
	l.Trace("/home/", fn)
	l.Patch("/home/", fn)
	l.Options("/home/", fn)
	l.Connect("/home/", fn)

	code, body := request(GET, "/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(POST, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, POST)

	code, body = request(PUT, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, PUT)

	code, body = request(DELETE, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, DELETE)

	code, body = request(HEAD, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, HEAD)

	code, body = request(TRACE, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, TRACE)

	code, body = request(PATCH, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, PATCH)

	code, body = request(OPTIONS, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, OPTIONS)

	code, body = request(CONNECT, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, CONNECT)
}

func TestAddAllMethodsMatch(t *testing.T) {
	fn := func(c Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	l := New()

	l.Match([]string{GET, POST, PUT, DELETE, HEAD, TRACE, PATCH, OPTIONS, CONNECT}, "/home/", fn)

	code, body := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(POST, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, POST)

	code, body = request(PUT, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, PUT)

	code, body = request(DELETE, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, DELETE)

	code, body = request(HEAD, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, HEAD)

	code, body = request(TRACE, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, TRACE)

	code, body = request(PATCH, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, PATCH)

	code, body = request(OPTIONS, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, OPTIONS)

	code, body = request(CONNECT, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, CONNECT)
}

func TestAddAllMethodsAny(t *testing.T) {
	fn := func(c Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	l := New()

	l.Any("/home/", fn)

	code, body := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, GET)

	code, body = request(POST, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, POST)

	code, body = request(PUT, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, PUT)

	code, body = request(DELETE, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, DELETE)

	code, body = request(HEAD, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, HEAD)

	code, body = request(TRACE, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, TRACE)

	code, body = request(PATCH, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, PATCH)

	code, body = request(OPTIONS, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, OPTIONS)

	code, body = request(CONNECT, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, CONNECT)
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

	// test same as above but already committed

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
		return func(w http.ResponseWriter, r *http.Request) {}
	}

	stdLibRawHandlerFunc3 := func(w http.ResponseWriter, r *http.Request) {}

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

type myContext struct {
	*Ctx
	text string
}

func (c *myContext) BaseContext() *Ctx {
	return c.Ctx
}

func (c *myContext) Reset(w http.ResponseWriter, r *http.Request) {
	c.Ctx.Reset(w, r)
	c.text = "test"
}

func (c *myContext) RequestComplete() {
	c.text = ""
}

func newCtx(l *LARS) Context {

	return &myContext{
		Ctx: NewContext(l),
	}
}

func TestCustomContext(t *testing.T) {

	var ctx *myContext

	l := New()
	l.RegisterContext(newCtx)

	l.Get("/home/", func(c Context) {
		ctx = c.(*myContext)
		c.Response().Write([]byte(ctx.text))
	})

	code, body := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "test")
	Equal(t, ctx.text, "")
}

func castContext(c Context, handler Handler) {
	handler.(func(*myContext))(c.(*myContext))
}

func TestCustomContextWrap(t *testing.T) {

	var ctx *myContext

	l := New()
	l.RegisterContext(newCtx)
	l.RegisterCustomHandler(func(*myContext) {}, castContext)

	PanicMatches(t, func() { l.RegisterCustomHandler(func(*myContext) {}, castContext) }, "Custom Type + CustomHandlerFunc already declared: func(*lars.myContext)")

	l.Get("/home/", func(c *myContext) {
		ctx = c
		c.Response().Write([]byte(c.text))
	})

	code, body := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "test")
	Equal(t, ctx.text, "")

	l2 := New()
	l2.Use(func(c Context) {
		c.(*myContext).text = "first handler"
		c.Next()
	})
	l2.RegisterContext(newCtx)
	l2.RegisterCustomHandler(func(*myContext) {}, castContext)

	l2.Get("/home/", func(c *myContext) {
		ctx = c
		c.Response().Write([]byte(c.text))
	})

	code, body = request(GET, "/home/", l2)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "first handler")
	Equal(t, ctx.text, "")

	l3 := New()
	l3.RegisterContext(newCtx)
	l3.RegisterCustomHandler(func(*myContext) {}, castContext)
	l3.Use(func(c Context) {
		c.(*myContext).text = "first handler"
		c.Next()
	})
	l3.Use(func(c *myContext) {
		c.text += " - second handler"
		c.Next()
	})
	l3.Use(func(c Context) {
		c.(*myContext).text += " - third handler"
		c.Next()
	})

	l3.Get("/home/", func(c *myContext) {
		ctx = c
		c.Response().Write([]byte(c.text))
	})

	code, body = request(GET, "/home/", l3)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "first handler - second handler - third handler")
	Equal(t, ctx.text, "")
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

var mRoutes = map[string]int{
	"/":                        0,
	"/:id":                     1,
	"/:id/test":                2,
	"/:id/*":                   2,
	"/user":                    1,
	"/user/:id":                2,
	"/admin/":                  1,
	"/admin/:id/":              2,
	"/admin/:id/*":             3,
	"/assets/*":                2,
	"/home/user":               2,
	"/home/user/:id":           3,
	"/home/user/:id/profile/*": 5,
}

func TestRouteMap(t *testing.T) {
	l := New()

	for k := range mRoutes {
		l.Get(k, basicHandler)
	}

	routes := l.GetRouteMap()
	var ok bool

	for _, r := range routes {

		_, ok = mRoutes[r.Path]

		Equal(t, ok, true)
		Equal(t, r.Depth, mRoutes[r.Path])
		Equal(t, r.Method, GET)
		MatchRegex(t, r.Handler, "^(.*/vendor/)?github.com/go-playground/lars.glob.func4$")
	}

	// next test must be separate, don't know why anyone would do this but it is possible
	l2 := New()
	l2.Get("/*", basicHandler)

	routes = l2.GetRouteMap()
	Equal(t, len(routes), 1)
	Equal(t, ok, true)
	Equal(t, routes[0].Path, "/*")
	Equal(t, routes[0].Depth, 1)
	Equal(t, routes[0].Method, GET)
	MatchRegex(t, routes[0].Handler, "^(.*/vendor/)?github.com/go-playground/lars.glob.func4$")
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

	l.SetRedirectTrailingSlash(false)

	code, _ = request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)

	code, _ = request(POST, "/home/", l)
	Equal(t, code, http.StatusOK)

	code, _ = request(GET, "/home", l)
	Equal(t, code, http.StatusNotFound)

	code, _ = request(GET, "/Home/", l)
	Equal(t, code, http.StatusNotFound)

	code, _ = request(POST, "/home", l)
	Equal(t, code, http.StatusNotFound)

	l.SetRedirectTrailingSlash(true)

	l.Get("/users/:id", basicHandler)
	l.Get("/users/:id/profile", basicHandler)

	code, _ = request(GET, "/users/10", l)
	Equal(t, code, http.StatusOK)

	code, _ = request(GET, "/users/10/", l)
	Equal(t, code, http.StatusMovedPermanently)

	l.SetRedirectTrailingSlash(false)

	code, _ = request(GET, "/users/10", l)
	Equal(t, code, http.StatusOK)

	code, _ = request(GET, "/users/10/", l)
	Equal(t, code, http.StatusNotFound)
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func request(method, path string, l *LARS) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := l.Serve()
	hf.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

func requestMultiPart(method string, url string, l *LARS) (int, string) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		fmt.Println("ERR FILE:", err)
	}

	buff := bytes.NewBufferString("FILE TEST DATA")
	_, err = io.Copy(part, buff)
	if err != nil {
		fmt.Println("ERR COPY:", err)
	}

	writer.WriteField("username", "joeybloggs")

	err = writer.Close()
	if err != nil {
		fmt.Println("ERR:", err)
	}

	r, _ := http.NewRequest(method, url, body)
	r.Header.Set(ContentType, writer.FormDataContentType())
	wr := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := l.Serve()
	hf.ServeHTTP(wr, r)

	return wr.Code, wr.Body.String()
}

type route struct {
	method string
	path   string
}
