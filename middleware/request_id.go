package middleware

import (
	uuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestID attaches a unique ID to each request via the X-Request-Id header.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uuid := uuid.New()
			c.Response().Header().Set("X-Request-Id", uuid.String())
			return next(c)
		}
	}
}
