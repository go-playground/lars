package lars

import (
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

func TestContext(t *testing.T) {

	l := New()
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c := newContext(l)

	var varParams []Param

	// Parameter
	param1 := Param{
		Key:   "userID",
		Value: "507f191e810c19729de860ea",
	}

	varParams = append(varParams, param1)

	//store
	storeMap := store{
		"User":        "Alice",
		"Information": []string{"Alice", "Bob", "40.712784", "-74.005941"},
	}

	c.params = varParams
	c.store = storeMap
	c.Request = r

	//Request
	NotEqual(t, c.Request, nil)

	//Response
	NotEqual(t, c.Response, nil)

	//Paramter by name
	bsonValue := c.Param("userID")
	NotEqual(t, len(bsonValue), 0)
	Equal(t, "507f191e810c19729de860ea", bsonValue)

	//Store
	c.Set("publicKey", "U|ydN3SX)B(hI8SV1R;(")

	value, exists := c.Get("publicKey")

	//Get
	Equal(t, true, exists)
	Equal(t, "U|ydN3SX)B(hI8SV1R;(", value)

	value, exists = c.Get("User")
	Equal(t, true, exists)
	Equal(t, "Alice", value)

	value, exists = c.Get("UserName")
	NotEqual(t, true, exists)
	NotEqual(t, "Alice", value)

	value, exists = c.Get("Information")
	Equal(t, true, exists)
	vString := value.([]string)

	Equal(t, "Alice", vString[0])
	Equal(t, "Bob", vString[1])
	Equal(t, "40.712784", vString[2])
	Equal(t, "-74.005941", vString[3])

	// Reset
	c.reset(w, r)

	//Request
	NotEqual(t, c.Request, nil)

	//Response
	NotEqual(t, c.Response, nil)

	//Set
	Equal(t, c.store, nil)

	// Index
	Equal(t, c.index, -1)

	// Handlers
	Equal(t, c.handlers, nil)

}

func TestQueryParams(t *testing.T) {
	l := New()
	l.Get("/home/:id", func(c *Context) {
		c.Param("nonexistant")
		c.Response.Write([]byte(c.Request.URL.RawQuery))
	})

	code, body := request(GET, "/home/13?test=true&test2=true", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "test=true&test2=true")
}

func TestNativeHandlersAndParseForm(t *testing.T) {

	l := New()
	l.Use(func(c *Context) {
		// to trigger the form parsing
		c.Param("nonexistant")
		c.Next()

	})
	l.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.FormValue("id")))
	})

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")

	l2 := New()
	l2.Use(func(c *Context) {
		// to trigger the form parsing
		c.ParseForm()
		c.Next()

	})
	l2.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.FormValue("id")))
	})

	code, body = request(GET, "/users/14", l2)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "14")

	l3 := New()
	l3.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		c.ParseForm()

		w.Write([]byte(r.FormValue("id")))
	})

	code, body = request(GET, "/users/15", l3)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "15")

	l4 := New()
	l4.Use(func(c *Context) {
		// to trigger the form parsing
		c.ParseForm()
		c.Next()

	})
	l4.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		c.ParseForm()

		w.Write([]byte(r.FormValue("id")))
	})

	code, body = request(GET, "/users/16", l4)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "16")

	l5 := New()
	l5.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseForm(); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(r.FormValue("id")))
	})

	code, body = request(GET, "/users/16?test=%2f%%efg", l5)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "invalid URL escape \"%%e\"")
}

func TestNativeHandlersAndParseMultiPartForm(t *testing.T) {

	l := New()
	l.Use(func(c *Context) {
		// to trigger the form parsing
		c.Param("nonexistant")
		c.Next()

	})
	l.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.FormValue("id")))
	})

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")

	l2 := New()
	l2.Use(func(c *Context) {
		// to trigger the form parsing
		c.ParseMultipartForm(10 << 5) // 5 MB
		c.Next()
	})
	l2.Post("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.FormValue("id")))
	})

	code, body = requestMultiPart(POST, "/users/14", l2)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "14")

	l3 := New()
	l3.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		c.ParseMultipartForm(10 << 5) // 5 MB

		w.Write([]byte(r.FormValue("id")))
	})

	code, body = requestMultiPart(GET, "/users/15", l3)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "15")

	l4 := New()
	l4.Use(func(c *Context) {
		// to trigger the form parsing
		c.ParseMultipartForm(10 << 5) // 5 MB
		c.Next()

	})
	l4.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		c.ParseMultipartForm(10 << 5) // 5 MB

		w.Write([]byte(r.FormValue("id")))
	})

	code, body = requestMultiPart(GET, "/users/16", l4)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "16")

	l5 := New()
	l5.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseMultipartForm(10 << 5); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(r.FormValue("id")))
	})

	code, body = requestMultiPart(GET, "/users/16?test=%2f%%efg", l5)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "invalid URL escape \"%%e\"")
}

func TestClientIP(t *testing.T) {
	l := New()
	c := newContext(l)

	c.Request, _ = http.NewRequest("POST", "/", nil)

	c.Request.Header.Set("X-Real-IP", " 10.10.10.10  ")
	c.Request.Header.Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	c.Request.RemoteAddr = "  40.40.40.40:42123 "

	Equal(t, c.ClientIP(), "10.10.10.10")

	c.Request.Header.Del("X-Real-IP")
	Equal(t, c.ClientIP(), "20.20.20.20")

	c.Request.Header.Set("X-Forwarded-For", "30.30.30.30  ")
	Equal(t, c.ClientIP(), "30.30.30.30")

	c.Request.Header.Del("X-Forwarded-For")
	Equal(t, c.ClientIP(), "40.40.40.40")
}

func TestAcceptedLanguages(t *testing.T) {
	l := New()
	c := newContext(l)

	c.Request, _ = http.NewRequest("POST", "/", nil)
	c.Request.Header.Set(AcceptedLanguage, "da, en-gb;q=0.8, en;q=0.7")

	languages := c.AcceptedLanguages()

	Equal(t, languages[0], "da")
	Equal(t, languages[1], "en-gb")
	Equal(t, languages[2], "en")

	c.Request.Header.Del(AcceptedLanguage)

	languages = c.AcceptedLanguages()

	Equal(t, languages, nil)

	c.Request.Header.Set(AcceptedLanguage, "")
	languages = c.AcceptedLanguages()

	Equal(t, languages, nil)
}
