package middleware

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/go-playground/lars"
)

// ANSIEscSeq is a predefined ANSI escape sequence
type ANSIEscSeq string

// ANSI escape sequences
// NOTE: in an standard xterm terminal the light colors will appear BOLD instead of the light variant
const (
	Black        ANSIEscSeq = "\x1b[30m"
	DarkGray                = "\x1b[30;1m"
	Blue                    = "\x1b[34m"
	LightBlue               = "\x1b[34;1m"
	Green                   = "\x1b[32m"
	LightGreen              = "\x1b[32;1m"
	Cyan                    = "\x1b[36m"
	LightCyan               = "\x1b[36;1m"
	Red                     = "\x1b[31m"
	LightRed                = "\x1b[31;1m"
	Magenta                 = "\x1b[35m"
	LightMagenta            = "\x1b[35;1m"
	Brown                   = "\x1b[33m"
	Yellow                  = "\x1b[33;1m"
	LightGray               = "\x1b[37m"
	White                   = "\x1b[37;1m"
	Underscore              = "\x1b[4m"
	Blink                   = "\x1b[5m"
	Inverse                 = "\x1b[7m"
	Reset                   = "\x1b[0m"
)

// LoggingAndRecovery handle HTTP request logging + recovery
func LoggingAndRecovery(c lars.Context) {

	t1 := time.Now()

	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1<<16)
			n := runtime.Stack(trace, true)
			log.Printf(" %srecovering from panic: %+v\nStack Trace:\n %s%s", Red, err, trace[:n], Reset)
			HandlePanic(c, trace[:n])
			return
		}
	}()

	c.Next()

	var color string

	res := c.Response()
	req := c.Request()
	code := res.Status()

	switch {
	case code >= http.StatusInternalServerError:
		color = Underscore + Blink + Red
	case code >= http.StatusBadRequest:
		color = Red
	case code >= http.StatusMultipleChoices:
		color = Yellow
	default:
		color = Green
	}

	t2 := time.Now()

	log.Printf("%s %d %s[%s%s%s] %q %v %d\n", color, code, Reset, color, req.Method, Reset, req.URL, t2.Sub(t1), res.Size())
}

// HandlePanic handles graceful panic by redirecting to friendly error page or rendering a friendly error page.
// trace passed just in case you want rendered to developer when not running in production
func HandlePanic(c lars.Context, trace []byte) {

	// redirect to or directly render friendly error page
}
