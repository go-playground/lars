package lars

import (
	"bytes"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

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

func TestDecode(t *testing.T) {

	type TestStruct struct {
		ID              int `form:"id"`
		Posted          string
		MultiPartPosted string
	}

	test := new(TestStruct)

	l := New()
	l.Post("/decode/:id", func(c Context) {
		err := c.Decode(true, 16<<10, test)
		Equal(t, err, nil)
	})
	l.Post("/decode2/:id", func(c Context) {
		err := c.Decode(false, 16<<10, test)
		Equal(t, err, nil)
	})
	hf := l.Serve()

	form := url.Values{}
	form.Add("Posted", "value")

	r, _ := http.NewRequest(POST, "/decode/13", strings.NewReader(form.Encode()))
	r.Header.Set(ContentType, ApplicationForm)
	w := httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "")

	test = new(TestStruct)
	r, _ = http.NewRequest(POST, "/decode2/13", strings.NewReader(form.Encode()))
	r.Header.Set(ContentType, ApplicationForm)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 0)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("MultiPartPosted", "value")
	Equal(t, err, nil)

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	err = writer.Close()
	Equal(t, err, nil)

	test = new(TestStruct)
	r, _ = http.NewRequest(POST, "/decode/13", body)
	r.Header.Set(ContentType, writer.FormDataContentType())
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)
	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "")
	Equal(t, test.MultiPartPosted, "value")

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)

	err = writer.WriteField("MultiPartPosted", "value")
	Equal(t, err, nil)

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	err = writer.Close()
	Equal(t, err, nil)

	test = new(TestStruct)
	r, _ = http.NewRequest(POST, "/decode2/13", body)
	r.Header.Set(ContentType, writer.FormDataContentType())
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)
	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 0)
	Equal(t, test.Posted, "")
	Equal(t, test.MultiPartPosted, "value")

	jsonBody := `{"ID":13,"Posted":"value","MultiPartPosted":"value"}`
	test = new(TestStruct)
	r, _ = http.NewRequest(POST, "/decode/13", strings.NewReader(jsonBody))
	r.Header.Set(ContentType, ApplicationJSON)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "value")

	xmlBody := `<TestStruct><ID>13</ID><Posted>value</Posted><MultiPartPosted>value</MultiPartPosted></TestStruct>`
	test = new(TestStruct)
	r, _ = http.NewRequest(POST, "/decode/13", strings.NewReader(xmlBody))
	r.Header.Set(ContentType, ApplicationXML)
	w = httptest.NewRecorder()

	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, test.ID, 13)
	Equal(t, test.Posted, "value")
	Equal(t, test.MultiPartPosted, "value")
}

func TestStream(t *testing.T) {
	l := New()

	count := 0

	l.Get("/stream/:id", func(c Context) {
		c.Stream(func(w io.Writer) bool {

			if _, err := w.Write([]byte("a")); err != nil {
				panic(err)
			}

			count++
			return count != 13
		})
	})

	l.Get("/stream2/:id", func(c Context) {
		c.Stream(func(w io.Writer) bool {

			if _, err := w.Write([]byte("a")); err != nil {
				panic(err)
			}

			count++

			if count == 5 {
				c.Response().Writer().(*closeNotifyingRecorder).close()
			}

			if count == 1000 {
				return false
			}

			return true
		})
	})

	code, body := request(GET, "/stream/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "aaaaaaaaaaaaa")

	count = 0

	code, body = request(GET, "/stream2/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "aaaaa")

}

func HandlerForName(c Context) {
	if _, err := c.Response().Write([]byte(c.HandlerName())); err != nil {
		panic(err)
	}
}

func TestHandlerName(t *testing.T) {
	l := New()
	l.Get("/users/:id", HandlerForName)

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	MatchRegex(t, body, "^(.*/vendor/)?github.com/go-playground/lars.HandlerForName$")
}

func TestContext(t *testing.T) {

	l := New()
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(l)

	var varParams []Param

	// Parameter
	param1 := Param{
		Key:   "userID",
		Value: "507f191e810c19729de860ea",
	}

	varParams = append(varParams, param1)
	c.params = varParams
	c.netContext = context.Background()
	c.request = r

	//Request
	NotEqual(t, c.request, nil)

	//Response
	NotEqual(t, c.response, nil)

	//Paramter by name
	bsonValue := c.Param("userID")
	NotEqual(t, len(bsonValue), 0)
	Equal(t, "507f191e810c19729de860ea", bsonValue)

	//Store
	ctx := c.Context()
	ctx = context.WithValue(ctx, "publicKey", "U|ydN3SX)B(hI8SV1R;(")
	c.WithContext(ctx)

	value, exists := c.Get("publicKey")

	//Get
	Equal(t, true, exists)
	Equal(t, "U|ydN3SX)B(hI8SV1R;(", value)

	c.WithValue("User", "Alice")
	value, exists = c.Value("User").(string)
	Equal(t, true, exists)
	Equal(t, "Alice", value)

	value, exists = c.Get("UserName")
	NotEqual(t, true, exists)
	NotEqual(t, "Alice", value)

	c.Set("Information", []string{"Alice", "Bob", "40.712784", "-74.005941"})

	value, exists = c.Get("Information")
	Equal(t, true, exists)
	vString := value.([]string)

	Equal(t, "Alice", vString[0])
	Equal(t, "Bob", vString[1])
	Equal(t, "40.712784", vString[2])
	Equal(t, "-74.005941", vString[3])

	// Reset
	c.RequestStart(w, r)

	//Request
	NotEqual(t, c.request, nil)

	//Response
	NotEqual(t, c.response, nil)

	//Set
	Equal(t, c.Value("test"), nil)

	// Index
	Equal(t, c.index, -1)

	// Handlers
	Equal(t, c.handlers, nil)

	cancelFunc := c.WithCancel()
	Equal(t, reflect.TypeOf(cancelFunc).String(), "context.CancelFunc")

	dt := time.Now().Add(time.Minute)
	cancelFunc = c.WithDeadline(dt)
	Equal(t, reflect.TypeOf(cancelFunc).String(), "context.CancelFunc")

	cancelFunc = c.WithTimeout(time.Minute)
	Equal(t, reflect.TypeOf(cancelFunc).String(), "context.CancelFunc")

	deadline, ok := c.Deadline()
	Equal(t, ok, true)
	Equal(t, deadline, dt)

	dc := c.Done()
	Equal(t, reflect.TypeOf(dc).String(), "<-chan struct {}")

	err := c.Err()
	Equal(t, err, nil)
}

func TestQueryParams(t *testing.T) {
	l := New()
	l.Get("/home/:id", func(c Context) {
		c.Param("nonexistant")
		if _, err := c.Response().Write([]byte(c.Request().URL.RawQuery)); err != nil {
			panic(err)
		}
	})

	code, body := request(GET, "/home/13?test=true&test2=true", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "test=true&test2=true")
}

func TestNativeHandlersAndParseForm(t *testing.T) {

	l := New()
	l.Use(func(c Context) {
		// to trigger the form parsing
		c.Param("nonexistant")
		c.Next()

	})
	l.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")

	l2 := New()
	l2.Use(func(c Context) {
		// to trigger the form parsing
		if err := c.ParseForm(); err != nil {
			panic(err)
		}
		c.Next()

	})
	l2.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = request(GET, "/users/14", l2)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "14")

	l3 := New()
	l3.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseForm(); err != nil {
			panic(err)
		}

		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = request(GET, "/users/15", l3)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "15")

	l4 := New()
	l4.Use(func(c Context) {
		// to trigger the form parsing
		if err := c.ParseForm(); err != nil {
			panic(err)
		}
		c.Next()

	})
	l4.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseForm(); err != nil {
			panic(err)
		}

		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = request(GET, "/users/16", l4)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "16")

	l5 := New()
	l5.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseForm(); err != nil {
			if _, errr := w.Write([]byte(err.Error())); errr != nil {
				panic(err)
			}
			return
		}

		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = request(GET, "/users/16?test=%2f%%efg", l5)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "invalid URL escape \"%%e\"")

	l6 := New()
	l6.Get("/chain-handler", func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("a")); err != nil {
				panic(err)
			}
			handler.ServeHTTP(w, r)
		})
	}, func(c Context) {
		if _, err := c.Response().Write([]byte("ok")); err != nil {
			panic(err)
		}
	})

	code, body = request(GET, "/chain-handler", l6)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "aok")

	l7 := New()
	l7.Get("/chain-handler", func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if _, err := w.Write([]byte("a")); err != nil {
			panic(err)
		}
		next.ServeHTTP(w, r)
	}, func(c Context) {
		if _, err := c.Response().Write([]byte("ok")); err != nil {
			panic(err)
		}
	})

	code, body = request(GET, "/chain-handler", l7)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "aok")
}

func TestNativeHandlersAndParseMultiPartForm(t *testing.T) {

	l := New()
	l.Use(func(c Context) {
		// to trigger the form parsing
		c.Param("nonexistant")
		c.Next()

	})
	l.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")

	l2 := New()
	l2.Use(func(c Context) {
		// to trigger the form parsing
		if err := c.ParseMultipartForm(10 << 5); err != nil {
			panic(err)
		} // 5 MB
		c.Next()
	})
	l2.Post("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = requestMultiPart(POST, "/users/14", l2)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "14")

	l3 := New()
	l3.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseMultipartForm(10 << 5); err != nil {
			panic(err)
		} // 5 MB

		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = requestMultiPart(GET, "/users/15", l3)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "15")

	l4 := New()
	l4.Use(func(c Context) {
		// to trigger the form parsing
		if err := c.ParseMultipartForm(10 << 5); err != nil {
			panic(err)
		} // 5 MB
		c.Next()

	})
	l4.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseMultipartForm(10 << 5); err != nil {
			panic(err)
		} // 5 MB

		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = requestMultiPart(GET, "/users/16", l4)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "16")

	l5 := New()
	l5.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {

		c := GetContext(w)
		if err := c.ParseMultipartForm(10 << 5); err != nil {
			if _, errr := w.Write([]byte(err.Error())); errr != nil {
				panic(err)
			}
			return
		}

		if _, err := w.Write([]byte(r.FormValue("id"))); err != nil {
			panic(err)
		}
	})

	code, body = requestMultiPart(GET, "/users/16?test=%2f%%efg", l5)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "invalid URL escape \"%%e\"")
}

func TestClientIP(t *testing.T) {
	l := New()
	c := NewContext(l)

	c.request, _ = http.NewRequest("POST", "/", nil)

	c.Request().Header.Set("X-Real-IP", " 10.10.10.10  ")
	c.Request().Header.Set("X-Forwarded-For", "  20.20.20.20, 30.30.30.30")
	c.Request().RemoteAddr = "  40.40.40.40:42123 "

	Equal(t, c.ClientIP(), "10.10.10.10")

	c.Request().Header.Del("X-Real-IP")
	Equal(t, c.ClientIP(), "20.20.20.20")

	c.Request().Header.Set("X-Forwarded-For", "30.30.30.30  ")
	Equal(t, c.ClientIP(), "30.30.30.30")

	c.Request().Header.Del("X-Forwarded-For")
	Equal(t, c.ClientIP(), "40.40.40.40")
}

func TestAttachment(t *testing.T) {

	l := New()

	l.Get("/dl", func(c Context) {
		f, _ := os.Open("logo.png")
		if err := c.Attachment(f, "logo.png"); err != nil {
			panic(err)
		}
	})

	l.Get("/dl-unknown-type", func(c Context) {
		f, _ := os.Open("logo.png")
		if err := c.Attachment(f, "logo"); err != nil {
			panic(err)
		}
	})

	r, _ := http.NewRequest(GET, "/dl", nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := l.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "attachment;filename=logo.png")
	Equal(t, w.Header().Get(ContentType), "image/png")
	Equal(t, w.Body.Len(), 3041)

	r, _ = http.NewRequest(GET, "/dl-unknown-type", nil)
	w = &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf = l.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "attachment;filename=logo")
	Equal(t, w.Header().Get(ContentType), "application/octet-stream")
	Equal(t, w.Body.Len(), 3041)
}

func TestInline(t *testing.T) {

	l := New()
	l.Get("/dl-inline", func(c Context) {
		f, _ := os.Open("logo.png")
		if err := c.Inline(f, "logo.png"); err != nil {
			panic(err)
		}
	})

	l.Get("/dl-unknown-type-inline", func(c Context) {
		f, _ := os.Open("logo.png")
		if err := c.Inline(f, "logo"); err != nil {
			panic(err)
		}
	})

	r, _ := http.NewRequest(GET, "/dl-inline", nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf := l.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "inline;filename=logo.png")
	Equal(t, w.Header().Get(ContentType), "image/png")
	Equal(t, w.Body.Len(), 3041)

	r, _ = http.NewRequest(GET, "/dl-unknown-type-inline", nil)
	w = &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
	hf = l.Serve()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentDisposition), "inline;filename=logo")
	Equal(t, w.Header().Get(ContentType), "application/octet-stream")
	Equal(t, w.Body.Len(), 3041)
}

func TestAcceptedLanguages(t *testing.T) {
	l := New()
	c := NewContext(l)

	c.request, _ = http.NewRequest("POST", "/", nil)
	c.Request().Header.Set(AcceptedLanguage, "da, en-GB;q=0.8, en;q=0.7")

	languages := c.AcceptedLanguages(false)

	Equal(t, languages[0], "da")
	Equal(t, languages[1], "en-GB")
	Equal(t, languages[2], "en")

	languages = c.AcceptedLanguages(true)

	Equal(t, languages[0], "da")
	Equal(t, languages[1], "en-gb")
	Equal(t, languages[2], "en")

	c.Request().Header.Del(AcceptedLanguage)

	languages = c.AcceptedLanguages(false)

	Equal(t, languages, []string{})

	c.Request().Header.Set(AcceptedLanguage, "")
	languages = c.AcceptedLanguages(false)

	Equal(t, languages, []string{})
}

type zombie struct {
	ID   int    `json:"id"   xml:"id"`
	Name string `json:"name" xml:"name"`
}

func TestXML(t *testing.T) {
	xmlData := `<zombie><id>1</id><name>Patient Zero</name></zombie>`

	l := New()
	l.Get("/xml", func(c Context) {
		if err := c.XML(http.StatusOK, zombie{1, "Patient Zero"}); err != nil {
			panic(err)
		}
	})
	l.Get("/badxml", func(c Context) {
		if err := c.XML(http.StatusOK, func() {}); err != nil {
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		}
	})

	hf := l.Serve()

	r, _ := http.NewRequest(GET, "/xml", nil)
	w := httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationXMLCharsetUTF8)
	Equal(t, w.Body.String(), xml.Header+xmlData)

	r, _ = http.NewRequest(GET, "/badxml", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusInternalServerError)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), "xml: unsupported type: func()\n")
}

func TestJSON(t *testing.T) {
	jsonData := `{"id":1,"name":"Patient Zero"}`
	callbackFunc := "CallbackFunc"

	l := New()
	l.Get("/json", func(c Context) {
		if err := c.JSON(http.StatusOK, zombie{1, "Patient Zero"}); err != nil {
			panic(err)
		}
	})
	l.Get("/badjson", func(c Context) {
		if err := c.JSON(http.StatusOK, func() {}); err != nil {
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		}
	})
	l.Get("/jsonp", func(c Context) {
		if err := c.JSONP(http.StatusOK, zombie{1, "Patient Zero"}, callbackFunc); err != nil {
			panic(err)
		}
	})
	l.Get("/badjsonp", func(c Context) {
		if err := c.JSONP(http.StatusOK, func() {}, callbackFunc); err != nil {
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
		}
	})

	hf := l.Serve()

	r, _ := http.NewRequest(GET, "/json", nil)
	w := httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationJSONCharsetUTF8)
	Equal(t, w.Body.String(), jsonData)

	r, _ = http.NewRequest(GET, "/badjson", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusInternalServerError)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), "json: unsupported type: func()\n")

	r, _ = http.NewRequest(GET, "/jsonp", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), ApplicationJavaScriptCharsetUTF8)
	Equal(t, w.Body.String(), callbackFunc+"("+jsonData+");")

	r, _ = http.NewRequest(GET, "/badjsonp", nil)
	w = httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusInternalServerError)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), "json: unsupported type: func()\n")
}

func TestText(t *testing.T) {
	txtData := `OMG I'm infected! #zombie`

	l := New()
	l.Get("/text", func(c Context) {
		if err := c.Text(http.StatusOK, txtData); err != nil {
			panic(err)
		}
	})

	hf := l.Serve()

	r, _ := http.NewRequest(GET, "/text", nil)
	w := httptest.NewRecorder()
	hf.ServeHTTP(w, r)

	Equal(t, w.Code, http.StatusOK)
	Equal(t, w.Header().Get(ContentType), TextPlainCharsetUTF8)
	Equal(t, w.Body.String(), txtData)
}
