package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// DeferredStatus reproduces the response-writer behaviour of github.com/gin-gonic/gin v1.11.0
// (MIT License), which holds a status back until the body is written so a later handler can still
// change it.
//
// DeferredStatus returns middleware that holds the response status back until
// the first body byte is written, and writes it once the whole chain has run.
// Handlers and the middleware around them can therefore still correct the
// status and the headers after first setting them.
func DeferredStatus() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			response := c.Response()
			writer := &deferredWriter{ResponseWriter: response.Writer, status: http.StatusOK}
			response.Writer = writer

			err := next(c)

			writer.commit()
			return err
		}
	}
}

type deferredWriter struct {
	http.ResponseWriter
	status    int
	committed bool
}

func (w *deferredWriter) WriteHeader(code int) {
	if !w.committed {
		w.status = code
	}
}

func (w *deferredWriter) Write(data []byte) (int, error) {
	w.commit()
	return w.ResponseWriter.Write(data)
}

func (w *deferredWriter) Flush() {
	w.commit()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *deferredWriter) commit() {
	if w.committed {
		return
	}
	w.committed = true
	w.ResponseWriter.WriteHeader(w.status)
}
