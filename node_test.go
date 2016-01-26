package lars

import (
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

func TestAddChain(t *testing.T) {
	l := New()

	l.Get("/home", func(Context) {})

	PanicMatches(t, func() { l.Get("/home", func(Context) {}) }, "Duplicate Handler for method 'GET' with path '/home'")
}
