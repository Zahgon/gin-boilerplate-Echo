package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
)

// Recover reproduces the recovery behaviour and the reported text of
// github.com/gin-gonic/gin v1.11.0 (MIT License).
//
// Recover returns middleware that turns a panic in a handler into an empty
// Internal Server Error response and reports it on standard error.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				fmt.Fprintf(os.Stderr, "\n\n[Recovery] %s panic recovered:\n%v\n%s\n",
					time.Now().Format("2006/01/02 - 15:04:05"), recovered, debug.Stack())
				err = c.NoContent(http.StatusInternalServerError)
			}()
			return next(c)
		}
	}
}
