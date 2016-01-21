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

// func TestRouter(t *testing.T) {
// 	l := New()
// 	// l.Get("/test/two", func(Context) {})
// 	// l.Get("/test/two/three", func(Context) {})
// 	// l.Get("/test/too/three/four", func(Context) {})

// 	// l.Get("/aa/ab/:ac/ad/ae", func(Context) {})
// 	// l.Get("/aa/bb/bc/:bd/:be", func(Context) {}) // bb priority 1

// 	l.Get("/aa/ab/:ac/ad", func(Context) {}) // ab priority 1
// 	l.Get("/aa/bb/:bc/:bd", func(Context) {})

// 	// fmt.Println(l.router.tree.static[0].param.path)
// 	// fmt.Println(l.router.tree.static[0].params.priority, l.router.tree.static[0].params.static.path)

// 	l.router.sort()

// 	// fmt.Println("PTH:", l.router.tree.params.priority, l.router.tree.params.path)

// 	for idx, n := range l.router.tree.static[0].static {
// 		fmt.Println(idx, n.priority, n.path)
// 	}

// 	// l.Get("/github.com/go-experimental/lars3/:blob/master历日本語/⌘/à/:alice/*", func(Context) {})
// }

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
