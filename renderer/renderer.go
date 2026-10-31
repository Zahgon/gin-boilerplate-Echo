package renderer

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// htmlContentType is echo.MIMETextHTMLCharsetUTF8 with the lower case charset
// the templates used to be served under.
const htmlContentType = "text/html; charset=utf-8"

// Template renders the HTML templates loaded from a glob pattern.
type Template struct {
	templates *template.Template
}

// New parses every template matching pattern and returns an echo.Renderer.
func New(pattern string) *Template {
	return &Template{templates: template.Must(template.ParseGlob(pattern))}
}

// Render implements echo.Renderer, executing the template registered under name.
// echo buffers the result before it writes the header, so the content type set
// here is the one that reaches the client.
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, htmlContentType)
	return t.templates.ExecuteTemplate(w, name, data)
}
