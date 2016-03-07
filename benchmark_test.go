package lars

import (
	"net/http"
	"runtime"
	"testing"
)

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

func benchRoutes(b *testing.B, router http.Handler, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTP(w, r)
		}
	}
}
func calcMem(name string, load func()) {

	m := new(runtime.MemStats)

	// before
	runtime.GC()
	runtime.ReadMemStats(m)
	before := m.HeapAlloc

	load()

	// after
	runtime.GC()
	runtime.ReadMemStats(m)
	after := m.HeapAlloc
	println("   "+name+":", after-before, "Bytes")
}

func larsHandler(c Context) {}

// func larsHandlerWrite(c Context) {
// 	io.WriteString(c.Response, c.Param("name"))
// }

// func larsHandlerTest(c Context) {
// 	io.WriteString(c.Response, c.Request().RequestURI)
// }

func loadLARS(routes []route) http.Handler {

	var h HandlerFunc = larsHandler

	e := New()

	for _, r := range routes {
		switch r.method {
		case GET:
			e.Get(r.path, h)
		case POST:
			e.Post(r.path, h)
		case PUT:
			e.Put(r.path, h)
		case PATCH:
			e.Patch(r.path, h)
		case DELETE:
			e.Delete(r.path, h)
		default:
			panic("Unknow HTTP method: " + r.method)
		}
	}
	return e.Serve()
}

// func loadLARSSingle(method, path string, h HandlerFunc) http.Handler {
// 	e := New()
// 	switch method {
// 	case GET:
// 		e.Get(path, h)
// 	case POST:
// 		e.Post(path, h)
// 	case PUT:
// 		e.Put(path, h)
// 	case PATCH:
// 		e.Patch(path, h)
// 	case DELETE:
// 		e.Delete(path, h)
// 	default:
// 		panic("Unknow HTTP method: " + method)
// 	}
// 	return e.Serve()
// }

func BenchmarkLARS_Param(b *testing.B) {
	m := New()
	m.Get("/user/:name", larsHandler)
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, m.Serve(), req)
}

func BenchmarkLARS_Param5(b *testing.B) {
	m := New()
	m.Get("/:a/:b/:c/:d/:e", larsHandler)
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, m.Serve(), req)
}

func BenchmarkLARS_Param20(b *testing.B) {
	twentyColon := "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t"
	twentyRoute := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t"
	m := New()
	m.Get(twentyColon, larsHandler)
	req, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequest(b, m.Serve(), req)
}

func BenchmarkLARS_ParamWrite(b *testing.B) {
	m := New()
	m.Get("/user/:name", larsHandler)
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, m.Serve(), req)
}
