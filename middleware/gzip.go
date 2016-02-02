package middleware

import (
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/go-playground/lars"
)

type gzipWriter struct {
	io.Writer
	http.ResponseWriter
	sniffComplete bool
}

func (w gzipWriter) Write(b []byte) (int, error) {

	if !w.sniffComplete {
		if w.Header().Get(lars.ContentType) == "" {
			w.Header().Set(lars.ContentType, http.DetectContentType(b))
		}
		w.sniffComplete = true
	}

	return w.Writer.Write(b)
}

func (w gzipWriter) Flush() error {
	return w.Writer.(*gzip.Writer).Flush()
}

func (w gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(ioutil.Discard)
	},
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip(c *lars.Context) {

	c.Response.Header().Add(lars.Vary, lars.AcceptEncoding)

	if strings.Contains(c.Request.Header.Get(lars.AcceptEncoding), lars.Gzip) {

		w := writerPool.Get().(*gzip.Writer)
		w.Reset(c.Response.Writer())

		defer func() {
			w.Close()
			writerPool.Put(w)
		}()

		gw := gzipWriter{Writer: w, ResponseWriter: c.Response.Writer()}
		c.Response.Header().Set(lars.ContentEncoding, lars.Gzip)
		c.Response.SetWriter(gw)
	}

	c.Next()
}
