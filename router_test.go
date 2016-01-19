package lars

import (
	"flag"
	"os"
	"testing"
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

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestRouter(t *testing.T) {
	l := New()
	l.Get("/github.com/go-experimental/lars3/:blob/master历日本語/⌘/à/", func(Context) {})
}

// func TestParamRouter(t *testing.T) {
// 	// l := New()
// 	// l.router.add("path", n)
// }

func BenchmarkRouter(b *testing.B) {

	for i := 0; i < b.N; i++ {
		for _, v := range "/github.com/go-experimental/lars3/blob/master历日本語/⌘/" {
			if v == 23 {
			}
		}
	}
}
