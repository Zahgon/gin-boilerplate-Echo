package middleware

import (
	"net/http"
	"path"
	"regexp"

	"github.com/Massad/gin-boilerplate/router"
	"github.com/labstack/echo/v4"
)

var (
	unsafePrefixCharacters = regexp.MustCompile("[^a-zA-Z0-9/-]+")
	repeatedSlashes        = regexp.MustCompile("/{2,}")
)

// TrailingSlashRedirect reproduces the redirect behaviour of github.com/gin-gonic/gin v1.11.0
// (MIT License), whose own routing tree derives from github.com/julienschmidt/httprouter
// (BSD-3-Clause).
//
// TrailingSlashRedirect returns middleware that answers an unrouted request
// with a redirect to the same path with its trailing slash added or removed,
// whenever the route table recommends one. It runs before the rest of the
// chain, so such a redirect carries none of the headers the other middleware
// would add.
func TrailingSlashRedirect(table *router.Table) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			if request.Method == http.MethodConnect || request.URL.Path == "/" {
				return next(c)
			}

			routed, recommended := table.Lookup(request.Method, request.URL.Path)
			if routed || !recommended {
				return next(c)
			}

			target := request.URL.Path
			if prefix := path.Clean(request.Header.Get("X-Forwarded-Prefix")); prefix != "." {
				prefix = unsafePrefixCharacters.ReplaceAllString(prefix, "")
				prefix = repeatedSlashes.ReplaceAllString(prefix, "/")
				target = prefix + "/" + request.URL.Path
			}

			if length := len(target); length > 1 && target[length-1] == '/' {
				request.URL.Path = target[:length-1]
			} else {
				request.URL.Path = target + "/"
			}

			code := http.StatusMovedPermanently
			if request.Method != http.MethodGet {
				code = http.StatusTemporaryRedirect
			}
			http.Redirect(c.Response(), request, request.URL.String(), code)
			return nil
		}
	}
}

// RouteGuard returns middleware that answers a request the route table does
// not route with the not-found page, so that a path the table rejects is never
// carried into a handler by a looser match.
func RouteGuard(table *router.Table) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			if routed, _ := table.Lookup(request.Method, request.URL.Path); !routed {
				return echo.ErrNotFound
			}
			return next(c)
		}
	}
}

// DecodedPathRouting returns middleware that makes the router match on the
// decoded request path, so a percent encoded path segment reaches the same
// route as the character it stands for.
func DecodedPathRouting() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Request().URL.RawPath = ""
			return next(c)
		}
	}
}
