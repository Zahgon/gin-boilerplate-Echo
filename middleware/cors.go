package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CORS returns a Cross-Origin Resource Sharing middleware.
func CORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Response().Header()
			header.Set("Access-Control-Allow-Origin", "http://localhost")
			header.Set("Access-Control-Max-Age", "86400")
			header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			header.Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
			header.Set("Access-Control-Expose-Headers", "Content-Length")
			header.Set("Access-Control-Allow-Credentials", "true")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	}
}
