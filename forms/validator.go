package forms

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

//DefaultValidator ...
type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ echo.Validator = &DefaultValidator{}

//Validate ...
func (v *DefaultValidator) Validate(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}

	return nil
}

func (v *DefaultValidator) lazyinit() {
	v.once.Do(func() {

		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// add any custom validations etc. here

		//Custom rule for user full name
		v.validate.RegisterValidation("fullName", ValidateFullName)
	})
}

func kindOfData(data interface{}) reflect.Kind {

	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

//ValidateFullName implements validator.Func
func ValidateFullName(fl validator.FieldLevel) bool {
	//Remove the extra space
	space := regexp.MustCompile(`\s+`)
	name := space.ReplaceAllString(fl.Field().String(), " ")

	//Remove trailing spaces
	name = strings.TrimSpace(name)

	//To support all possible languages
	matched, _ := regexp.Match(`^[^±!@£$%^&*_+§¡€#¢§¶•ªº«\\/<>?:;'"|=.,0123456789]{3,20}$`, []byte(name))
	return matched
}
