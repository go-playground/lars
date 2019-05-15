package middleware

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/go-playground/ansi"
	"github.com/go-playground/lars"
)

const (
	status500 = ansi.Underline + ansi.Blink + ansi.Red
	status400 = ansi.Red
	status300 = ansi.Yellow
	status    = ansi.Green
)

// LoggingAndRecovery handle HTTP request logging + recovery
func LoggingAndRecovery(c lars.Context) {

	t1 := time.Now()

	defer func() {
		if err := recover(); err != nil {
			trace := make([]byte, 1<<16)
			n := runtime.Stack(trace, true)
			log.Printf(" %srecovering from panic: %+v\nStack Trace:\n %s%s", ansi.Red, err, trace[:n], ansi.Reset)
			HandlePanic(c, trace[:n])
			return
		}
		color := status
	        res := c.Response()
	        req := c.Request()
	        code := res.Status()

	        switch {
	            case code >= http.StatusInternalServerError:
		        color = status500
	            case code >= http.StatusBadRequest:
		        color = status400
	            case code >= http.StatusMultipleChoices:
		        color = status300
	        }

	        t2 := time.Now()

	        log.Printf("%s %d %s[%s%s%s] %q %v %d\n", color, code, ansi.Reset, color, req.Method, ansi.Reset, req.URL, t2.Sub(t1), res.Size())
	}()

	c.Next()	
}

// HandlePanic handles graceful panic by redirecting to friendly error page or rendering a friendly error page.
// trace passed just in case you want rendered to developer when not running in production
func HandlePanic(c lars.Context, trace []byte) {

	// redirect to or directly render friendly error page
}
