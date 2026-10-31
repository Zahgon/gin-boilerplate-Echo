package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// StaticFiles reproduces the file-serving behaviour of github.com/gin-gonic/gin v1.11.0
// (MIT License).
//
// StaticFiles returns a handler that serves the files below root under
// urlPrefix. A path that does not resolve to a file falls through to the
// not-found page, and directories are never listed, which leaves a directory
// request answered with the not-found status the empty listing is written
// under.
func StaticFiles(urlPrefix, root string) echo.HandlerFunc {
	files := onlyFiles{http.Dir(root)}
	server := http.StripPrefix(urlPrefix, http.FileServer(files))

	return func(c echo.Context) error {
		name := strings.TrimPrefix(c.Request().URL.Path, urlPrefix)
		if unescaped, err := url.QueryUnescape(name); err == nil {
			name = unescaped
		}

		file, err := files.Open(name)
		if err != nil {
			return echo.ErrNotFound
		}
		file.Close()

		response := c.Response()
		writer := &deferredWriter{
			ResponseWriter: response.Writer,
			status:         http.StatusNotFound,
		}
		server.ServeHTTP(writer, c.Request())
		writer.commit()

		response.Status = writer.status
		response.Committed = true
		return nil
	}
}

type onlyFiles struct {
	http.FileSystem
}

func (o onlyFiles) Open(name string) (http.File, error) {
	file, err := o.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return unlistedDirectory{file}, nil
}

type unlistedDirectory struct {
	http.File
}

func (unlistedDirectory) Readdir(int) ([]os.FileInfo, error) {
	return nil, nil
}
