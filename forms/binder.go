package forms

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo/v4"
)

//JSONBinder decodes the request body as JSON, whatever the Content-Type is.
//The handlers bind JSON payloads exclusively, so the body is never negotiated
//by content type, and an empty body is reported as an error instead of being
//silently accepted.
type JSONBinder struct{}

var _ echo.Binder = &JSONBinder{}

//Bind ...
func (b *JSONBinder) Bind(obj interface{}, c echo.Context) error {

	req := c.Request()

	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}

	return json.NewDecoder(req.Body).Decode(obj)
}
