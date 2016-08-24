// +build go1.7

package lars

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

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
	c := NewContext(l)

	var varParams []Param

	// Parameter
	param1 := Param{
		Key:   "userID",
		Value: "507f191e810c19729de860ea",
	}

	varParams = append(varParams, param1)
	c.params = varParams
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

func TestNativeContext(t *testing.T) {

	var val1, val2 string

	l := New()

	l.Use(func(w http.ResponseWriter, r *http.Request) {
		*r = *r.WithContext(context.WithValue(r.Context(), 0, "testval1"))
	})
	l.Use(func(w http.ResponseWriter, r *http.Request) {
		*r = *r.WithContext(context.WithValue(r.Context(), 1, "testval2"))
	})
	l.Get("/users/:id", func(c Context) {
		val1 = c.Request().Context().Value(0).(string)
		val2 = c.Request().Context().Value(1).(string)
	})

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")

	Equal(t, val1, "testval1")
	Equal(t, val2, "testval2")
}

func TestNativeContextUsingCtx(t *testing.T) {

	var val1, val2 string

	l := New()

	l.Use(func(w http.ResponseWriter, r *http.Request) {
		c := GetContext(w)
		c.WithContext(context.WithValue(r.Context(), 0, "testval1"))
	})
	l.Use(func(w http.ResponseWriter, r *http.Request) {
		c := GetContext(w)
		*r = *c.Request().WithContext(context.WithValue(r.Context(), 1, "testval2"))
	})
	l.Get("/users/:id", func(c Context) {
		val1 = c.Request().Context().Value(0).(string)
		val2 = c.Request().Context().Value(1).(string)
	})

	code, body := request(GET, "/users/13", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "")

	Equal(t, val1, "testval1")
	Equal(t, val2, "testval2")
}
