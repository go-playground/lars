package lars

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/websocket"
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
	l := New()
	l.WebSocket("/ws", func(c Context) {

		recv := make([]byte, 1000)

		i, err := c.WebSocket().Read(recv)
		if err == nil {
			_, err := c.WebSocket().Write(recv[:i])
			if err != nil {
				panic(err)
			}
		}
	})

	server := httptest.NewServer(l.Serve())
	defer server.Close()

	addr := server.Listener.Addr().String()
	origin := "http://localhost"

	url := fmt.Sprintf("ws://%s/ws", addr)
	ws, err := websocket.Dial(url, "", origin)
	Equal(t, err, nil)

	defer ws.Close()

	_, err = ws.Write([]byte("websockets in action!"))
	Equal(t, err, nil)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(ws)
	Equal(t, err, nil)
	Equal(t, "websockets in action!", buf.String())
}
