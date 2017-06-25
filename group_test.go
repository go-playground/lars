package lars

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
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

func TestWebsockets(t *testing.T) {

	origin := "http://localhost"

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			o := r.Header.Get(Origin)
			return o == origin
		},
	}

	l := New()
	l.WebSocket(upgrader, "/ws", func(c Context) {

		messageType, b, err := c.WebSocket().ReadMessage()
		if err != nil {
			return
		}

		if err == nil {
			err := c.WebSocket().WriteMessage(messageType, b)
			if err != nil {
				panic(err)
			}
		}
	})

	server := httptest.NewServer(l.Serve())
	defer server.Close()

	addr := server.Listener.Addr().String()

	header := make(http.Header, 0)
	header.Set(Origin, origin)

	url := fmt.Sprintf("ws://%s/ws", addr)
	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	Equal(t, err, nil)

	defer ws.Close()

	err = ws.WriteMessage(websocket.TextMessage, []byte("websockets in action!"))
	Equal(t, err, nil)

	typ, b, err := ws.ReadMessage()
	Equal(t, err, nil)
	Equal(t, typ, websocket.TextMessage)
	Equal(t, "websockets in action!", string(b))

	wsBad, res, err := websocket.DefaultDialer.Dial(url, nil)
	NotEqual(t, err, nil)
	Equal(t, wsBad, nil)
	Equal(t, res.StatusCode, http.StatusForbidden)
}

func TestGrouplogic(t *testing.T) {

	var aa, bb, cc, tl int

	aM := func(c Context) {
		aa++
		c.Next()
	}

	bM := func(c Context) {
		bb++
		c.Next()
	}

	cM := func(c Context) {
		cc++
		c.Next()
	}

	l := New()
	l.Use(func(c Context) {
		tl++
		c.Next()
	})

	a := l.Group("/a", aM)
	a.Get("/test", func(c Context) {
		c.JSON(http.StatusOK, "a-ok")
	})

	b := a.Group("/b", bM)
	b.Get("/test", func(c Context) {
		c.JSON(http.StatusOK, "b-ok")
	})

	c := b.Group("/c", cM)
	c.Get("/test", func(c Context) {
		c.JSON(http.StatusOK, "c-ok")
	})

	code, body := request(GET, "/a/test", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "\"a-ok\"")
	Equal(t, tl, 1)
	Equal(t, aa, 1)

	code, body = request(GET, "/a/b/test", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "\"b-ok\"")
	Equal(t, tl, 2)
	Equal(t, aa, 2)
	Equal(t, bb, 1)

	code, body = request(GET, "/a/b/c/test", l)
	Equal(t, code, http.StatusOK)
	Equal(t, body, "\"c-ok\"")
	Equal(t, tl, 3)
	Equal(t, aa, 3)
	Equal(t, bb, 2)
	Equal(t, cc, 1)
}
