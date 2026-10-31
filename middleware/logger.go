package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// requestLogLayout is the layout github.com/gin-gonic/gin v1.11.0 writes each request with
// (MIT License). It is reproduced so the lines this service writes stay the lines it wrote.
const requestLogLayout = "[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s"

// RequestLogger returns middleware that writes one line per request to
// standard output, in the layout the application has always logged.
func RequestLogger() echo.MiddlewareFunc {
	return newRequestLogger(os.Stdout)
}

func newRequestLogger(out io.Writer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			path := request.URL.Path
			query := request.URL.RawQuery

			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			end := time.Now()

			latency := end.Sub(start)
			if latency > time.Minute {
				latency = latency.Truncate(time.Second)
			}
			if query != "" {
				path += "?" + query
			}

			fmt.Fprintf(out, requestLogLayout,
				end.Format("2006/01/02 - 15:04:05"),
				"", c.Response().Status, "",
				latency,
				c.RealIP(),
				"", request.Method, "",
				path,
				"",
			)
			return nil
		}
	}
}
