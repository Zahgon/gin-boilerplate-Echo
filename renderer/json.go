package renderer

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

// jsonContentType is echo.MIMEApplicationJSONCharsetUTF8 with the lower case
// charset the handlers used to send.
const jsonContentType = "application/json; charset=utf-8"

// JSONSerializer encodes JSON responses as a compact body without a trailing
// newline, under an explicit charset, which is what the handlers emitted before
// the migration.
type JSONSerializer struct{}

var _ echo.JSONSerializer = &JSONSerializer{}

// Serialize implements echo.JSONSerializer. The indent is ignored because the
// responses were never pretty printed.
func (s *JSONSerializer) Serialize(c echo.Context, data interface{}, indent string) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// The content type is rewritten before the first write, which is what
	// commits the response header.
	c.Response().Header().Set(echo.HeaderContentType, jsonContentType)

	_, err = c.Response().Write(encoded)
	return err
}

// Deserialize implements echo.JSONSerializer. Nothing calls it, since
// forms.JSONBinder reads the request body itself.
func (s *JSONSerializer) Deserialize(c echo.Context, data interface{}) error {
	return echo.DefaultJSONSerializer{}.Deserialize(c, data)
}
