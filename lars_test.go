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

var basicHandler = func(*Context) {}

func TestFindOneOffs(t *testing.T) {
	fn := func(c *Context) {
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

	l.Get("/", func(c *Context) {
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
	l.Get(path, func(c *Context) {
		p, _ := c.Param("id")
		c.Response().Write([]byte(p))
	})
	code, body := request(GET, "/github.com/go-playground/808w70/", l)

	Equal(t, code, http.StatusOK)
	Equal(t, body, "808w70")
}

func TestlarsTwoParam(t *testing.T) {
	var p Params

	l := New()
	path := "/github.com/user/:id/:age/"
	l.Get(path, func(c *Context) {
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

	l.Get(path1, func(c *Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
	})

	l.Get(path2, func(c *Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
	})

	l.Get(path3, func(c *Context) {
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
	var context *Context

	l := New()
	l.Get("/:a/:b/:c", func(c *Context) {
		context = c
	})

	code, _ := request(GET, "/1/2/3", l)
	Equal(t, code, http.StatusOK)

	value, exists := context.P(0)

	Equal(t, "1", value)
	Equal(t, true, exists)

	value, exists = context.P(1)
	Equal(t, "2", value)
	Equal(t, true, exists)

	value, exists = context.P(2)
	Equal(t, "3", value)
	Equal(t, true, exists)

	value, exists = context.P(4)
	Equal(t, exists, false)
	Equal(t, value, "")

	value, exists = context.Param("a")

	Equal(t, "1", value)
	Equal(t, true, exists)

	value, exists = context.Param("b")
	Equal(t, "2", value)
	Equal(t, true, exists)

	value, exists = context.Param("c")
	Equal(t, "3", value)
	Equal(t, true, exists)

	value, exists = context.Param("key")
	Equal(t, false, exists)
	Equal(t, "", value)

}

func TestRouterMixParamMatchAny(t *testing.T) {
	var p Params

	l := New()

	//Route
	l.Get("/users/:id/*", func(c *Context) {
		c.Response().Write([]byte(c.Request().URL.Path))
		p = c.Params()
	})
	code, body := request(GET, "/users/joe/comments", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "joe", p[0].Value)
	Equal(t, "/users/joe/comments", body)
}

func TestRouterMultiRoute(t *testing.T) {
	var p string
	var parameters Params

	l := New()
	//Route
	l.Get("/users", func(c *Context) {
		c.Set("path", "/users")
		value, ok := c.Get("path")
		if ok {
			p = value.(string)
		}
	})

	l.Get("/users/:id", func(c *Context) {
		parameters = c.Params()
	})
	// Route > /users
	code, _ := request(GET, "/users", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "/users", p)
	// Route > /users/:id
	code, _ = request(GET, "/users/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "1", parameters[0].Value)

	// Route > /user/1
	code, _ = request(GET, "/user/1", l)
	Equal(t, http.StatusNotFound, code)
}

func TestRouterParamNames(t *testing.T) {
	var getP string
	var p Params

	l := New()
	//Routes
	l.Get("/users", func(c *Context) {
		c.Set("path", "/users")
		value, ok := c.Get("path")
		if ok {
			getP = value.(string)
		}
	})

	l.Get("/users/:id", func(c *Context) {
		p = c.Params()
	})

	l.Get("/users/:id/files/:fid", func(c *Context) {
		p = c.Params()
	})

	// Route > users
	code, _ := request(GET, "/users", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "/users", getP)

	// Route >/users/:id
	code, _ = request(GET, "/users/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "id", p[0].Key)
	Equal(t, "1", p[0].Value)

	// Route > /users/:uid/files/:fid
	code, _ = request(GET, "/users/1/files/1", l)
	Equal(t, code, http.StatusOK)
	Equal(t, "id", p[0].Key)
	Equal(t, "1", p[0].Value)
	Equal(t, "fid", p[1].Key)
	Equal(t, "1", p[1].Value)
}

func TestRouterAPI(t *testing.T) {
	l := New()

	for _, route := range githubAPI {
		l.handle(route.method, route.path, []Handler{func(c *Context) {
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
	fn := func(c *Context) {
		c.Response().Write([]byte(c.Request().Method))
	}

	var log string

	logger := func(c *Context) {
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

	logger2 := func(c *Context) {
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
	fn := func(c *Context) {
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
	fn := func(c *Context) {
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
	fn := func(c *Context) {
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
	fn := func(c *Context) {
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

	fn := func(c *Context) { c.Response().Write([]byte(c.Request().URL.Path)) }

	var hf HandlerFunc

	hf = func(c *Context) { c.Response().Write([]byte(c.Request().URL.Path)) }

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

type myGlobals struct {
	text string
}

func (g *myGlobals) Reset(c *Context) {
	g.text = "URL: " + c.Request().URL.Path
}

func (g *myGlobals) Done() {
	g.text = ""
}

var _ IGlobals = &myGlobals{}

func TestCustomGlobals(t *testing.T) {

	var l *LARS

	globals := &myGlobals{}

	fn := func() IGlobals {
		return globals
	}

	l = New()
	l.RegisterGlobals(fn)

	l.Get("/home/", func(c *Context) {
		c.Response().Write([]byte(c.Globals.(*myGlobals).text))
	})

	code, body := request(GET, "/home/", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "URL: /home/")
	Equal(t, globals.text, "")
}

func TestCustom404(t *testing.T) {

	fn := func(c *Context) {
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

func request(method, path string, l *LARS) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	hf := l.Serve()
	hf.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

type route struct {
	method string
	path   string
}

var githubAPI = []route{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	//{"PUT", "/authorizations/clients/:client_id"},
	//{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	// Activity
	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	//{"PATCH", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	// Gists
	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	//{"GET", "/gists/public"},
	//{"GET", "/gists/starred"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	//{"PATCH", "/gists/:id"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	// Git Data
	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	// Issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	//{"PATCH", "/repos/:owner/:repo/issues/:number"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	//{"GET", "/repos/:owner/:repo/issues/events"},
	//{"GET", "/repos/:owner/:repo/issues/events/:id"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	//{"PATCH", "/repos/:owner/:repo/labels/:name"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	// Organizations
	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	//{"PATCH", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	//{"PATCH", "/teams/:id"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	// Pull Requests
	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	// Repositories
	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	//{"PATCH", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	//{"PATCH", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	//{"GET", "/repos/:owner/:repo/contents/*path"},
	//{"PUT", "/repos/:owner/:repo/contents/*path"},
	//{"DELETE", "/repos/:owner/:repo/contents/*path"},
	//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	//{"PATCH", "/repos/:owner/:repo/keys/:id"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	//{"PATCH", "/repos/:owner/:repo/releases/:id"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	// Users
	{"GET", "/users/:user"},
	{"GET", "/user"},
	//{"PATCH", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	//{"PATCH", "/user/keys/:id"},
	{"DELETE", "/user/keys/:id"},
}
