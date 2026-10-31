package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

const gzipEncoding = "gzip"

// incompressibleExtensions is the set github.com/gin-contrib/gzip v1.2.5 excludes by default
// (MIT License), the middleware the service compressed with before the migration. The whole file
// reproduces that middleware's behaviour, not its code.
var incompressibleExtensions = map[string]bool{
	".png": true, ".gif": true, ".jpeg": true, ".jpg": true,
}

// Gzip returns middleware that compresses response bodies for clients that
// accept it. An error response is sent uncompressed and without the encoding
// headers, and bodies of already compressed media types are left alone.
func Gzip() echo.MiddlewareFunc {
	pool := sync.Pool{New: func() interface{} {
		compressor, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return compressor
	}}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !compressible(c.Request()) {
				return next(c)
			}

			response := c.Response()
			header := response.Header()
			plain := response.Writer

			compressor := pool.Get().(*gzip.Writer)
			compressor.Reset(plain)

			header.Set(echo.HeaderContentEncoding, gzipEncoding)
			header.Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			if tag := header.Get("ETag"); tag != "" && !strings.HasPrefix(tag, "W/") {
				header.Set("ETag", "W/"+tag)
			}

			writer := &gzipWriter{ResponseWriter: plain, compressor: compressor}
			response.Writer = writer

			defer func() {
				if writer.status >= http.StatusBadRequest {
					dropEncodingHeaders(header)
					compressor.Reset(io.Discard)
				} else if !writer.compressed {
					compressor.Reset(io.Discard)
				}
				compressor.Close()
				pool.Put(compressor)
				response.Writer = plain
			}()

			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}

func compressible(request *http.Request) bool {
	if !strings.Contains(request.Header.Get(echo.HeaderAcceptEncoding), gzipEncoding) ||
		strings.Contains(request.Header.Get(echo.HeaderConnection), "Upgrade") {
		return false
	}
	return !incompressibleExtensions[filepath.Ext(request.URL.Path)]
}

func dropEncodingHeaders(header http.Header) {
	header.Del(echo.HeaderContentEncoding)
	header.Del(echo.HeaderVary)
	header.Del("ETag")
}

type gzipWriter struct {
	http.ResponseWriter
	compressor *gzip.Writer
	status     int
	compressed bool
}

func (w *gzipWriter) WriteHeader(code int) {
	w.status = code
	w.Header().Del(echo.HeaderContentLength)
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipWriter) Write(data []byte) (int, error) {
	w.Header().Del(echo.HeaderContentLength)
	if w.status >= http.StatusBadRequest {
		dropEncodingHeaders(w.Header())
		return w.ResponseWriter.Write(data)
	}
	w.compressed = true
	return w.compressor.Write(data)
}

func (w *gzipWriter) Flush() {
	w.compressor.Flush()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
