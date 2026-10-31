// Package middleware holds the interceptors the service runs around every request.
//
// Several of them reproduce behaviour or output formats of the framework the service was built
// on, github.com/gin-gonic/gin v1.11.0 (MIT License), and of its companion
// github.com/gin-contrib/gzip v1.2.5 (MIT License). No code is taken from either; only the
// behaviour a client can observe, and the text the service writes, are reproduced. Each file that
// does so carries a notice of its own.
package middleware

import (
	"net/http"

	"github.com/Massad/gin-boilerplate/models"
	"github.com/labstack/echo/v4"
)

var authModel = new(models.AuthModel)

// TokenAuth validates the JWT access token and sets userID in the context.
func TokenAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenAuth, err := authModel.ExtractTokenMetadata(c.Request())
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Please login first"})
			}

			userID, err := authModel.FetchAuth(tokenAuth)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Please login first"})
			}

			c.Set("userID", userID)
			return next(c)
		}
	}
}
