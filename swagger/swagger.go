// Package swagger serves the Swagger UI distribution and the generated API
// definition over HTTP.
package swagger

import (
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/labstack/echo/v4"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/swag"
)

type uiConfig struct {
	URL                      string
	DocExpansion             string
	Title                    string
	Oauth2RedirectURL        template.JS
	DefaultModelsExpandDepth int
	DeepLinking              bool
	PersistAuthorization     bool
	Oauth2DefaultClientID    string
}

var defaultConfig = uiConfig{
	URL:          "doc.json",
	DocExpansion: "list",
	Title:        "Swagger UI",
	Oauth2RedirectURL: "`${window.location.protocol}//${window.location.host}$" +
		"{window.location.pathname.split('/').slice(0, window.location.pathname.split('/').length - 1).join('/')}" +
		"/oauth2-redirect.html`",
	DefaultModelsExpandDepth: 1,
	DeepLinking:              true,
}

var servedFile = regexp.MustCompile(`(.*)(index\.html|doc\.json|favicon-16x16\.png|favicon-32x32\.png|/oauth2-redirect\.html|swagger-ui\.css|swagger-ui\.css\.map|swagger-ui\.js|swagger-ui\.js\.map|swagger-ui-bundle\.js|swagger-ui-bundle\.js\.map|swagger-ui-standalone-preset\.js|swagger-ui-standalone-preset\.js\.map)[?|.]*`)

const plainTextContentType = "text/plain; charset=utf-8"

var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript",
	".png":  "image/png",
	".json": "application/json; charset=utf-8",
}

// Handler returns a handler that serves the Swagger UI, its assets and the
// generated definition, for requests routed under a /swagger/ prefix.
func Handler() echo.HandlerFunc {
	index := template.Must(template.New("swagger_index.html").Parse(indexTemplate))
	assets := swaggerFiles.Handler
	var once sync.Once

	return func(c echo.Context) error {
		response := c.Response()

		matches := servedFile.FindStringSubmatch(c.Request().RequestURI)
		if len(matches) != 3 {
			response.Header().Set(echo.HeaderContentType, plainTextContentType)
			response.WriteHeader(http.StatusNotFound)
			_, err := response.Write([]byte(http.StatusText(http.StatusNotFound)))
			return err
		}

		name := matches[2]
		once.Do(func() { assets.Prefix = matches[1] })

		if contentType, ok := contentTypes[filepath.Ext(name)]; ok {
			response.Header().Set(echo.HeaderContentType, contentType)
		}

		switch name {
		case "index.html":
			return index.Execute(response, defaultConfig)
		case "doc.json":
			document, err := swag.ReadDoc(swag.Name)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			_, err = response.Write([]byte(document))
			return err
		default:
			assets.ServeHTTP(response, c.Request())
			return nil
		}
	}
}
