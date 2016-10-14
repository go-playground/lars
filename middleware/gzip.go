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

func (w *gzipWriter) Write(b []byte) (int, error) {

	if !w.sniffComplete {
		if w.Header().Get(lars.ContentType) == "" {
			w.Header().Set(lars.ContentType, http.DetectContentType(b))
		}

		w.sniffComplete = true
	}

	return w.Writer.Write(b)
}

func (w *gzipWriter) Flush() error {
	return w.Writer.(*gzip.Writer).Flush()
}

func (w *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *gzipWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

var writerPool = sync.Pool{
	New: func() interface{} {
		return &gzipWriter{Writer: gzip.NewWriter(ioutil.Discard)}
	},
}

// Gzip returns a middleware which compresses HTTP response using gzip compression
// scheme.
func Gzip(c lars.Context) {

	c.Response().Header().Add(lars.Vary, lars.AcceptEncoding)

	if strings.Contains(c.Request().Header.Get(lars.AcceptEncoding), lars.Gzip) {

		gw := writerPool.Get().(*gzipWriter)
		gw.sniffComplete = false
		w := gw.Writer.(*gzip.Writer)
		w.Reset(c.Response().Writer())
		gw.ResponseWriter = c.Response().Writer()

		defer func() {

			if !gw.sniffComplete {
				// We have to reset response to it's pristine state when
				// nothing is written to body.
				c.Response().Header().Del(lars.ContentEncoding)
				w.Reset(ioutil.Discard)
			}

			w.Close()
			writerPool.Put(gw)
		}()

		c.Response().Header().Set(lars.ContentEncoding, lars.Gzip)
		c.Response().SetWriter(gw)
	}

	c.Next()
}

// GzipLevel returns a middleware which compresses HTTP response using gzip compression
// scheme using the level specified
func GzipLevel(level int) lars.HandlerFunc {

	// test gzip level, then don't have to each time one is created
	// in the pool

	if _, err := gzip.NewWriterLevel(ioutil.Discard, level); err != nil {
		panic(err)
	}

	var pool = sync.Pool{
		New: func() interface{} {
			z, _ := gzip.NewWriterLevel(ioutil.Discard, level)

			return &gzipWriter{Writer: z}
		},
	}

	return func(c lars.Context) {
		c.Response().Header().Add(lars.Vary, lars.AcceptEncoding)

		if strings.Contains(c.Request().Header.Get(lars.AcceptEncoding), lars.Gzip) {

			gw := pool.Get().(*gzipWriter)
			gw.sniffComplete = false
			w := gw.Writer.(*gzip.Writer)
			w.Reset(c.Response().Writer())
			gw.ResponseWriter = c.Response().Writer()

			defer func() {

				if !gw.sniffComplete {
					// We have to reset response to it's pristine state when
					// nothing is written to body.
					c.Response().Header().Del(lars.ContentEncoding)
					w.Reset(ioutil.Discard)
				}

				w.Close()
				pool.Put(gw)
			}()

			c.Response().Header().Set(lars.ContentEncoding, lars.Gzip)
			c.Response().SetWriter(gw)
		}

		c.Next()
	}
}
