package rest

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gonzispina/gokit/errors"
)

// ErrLib used in border cases
var ErrLib = errors.New("unexpected library behaviour", "internal_library_error")

// Validator representation
type Validator struct {
	validator *validator.Validate
}

// NewValidator retrieve a Validator pointer.
func NewValidator() *Validator {
	validate := validator.New()

	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validator: validate,
	}
}

// Value is used to validate an interface fields values.
func (v *Validator) Value(value reflect.Value) errors.Error {
	var err error

	kind := reflect.Indirect(value).Kind()
	s := value.Interface()

	switch kind {
	case reflect.Struct:
		err = v.validator.Struct(s)
	case reflect.Slice:
		err = v.validator.Var(s, "required,dive")
	default:
		return ErrLib
	}

	if err == nil {
		return nil
	}

	errs, ok := err.(validator.ValidationErrors)
	if !ok || len(errs) == 0 {
		return ErrLib
	}

	e := errs[0]
	var message string
	var code string
	switch e.Tag() {
	case "required":
		code = "param_is_required"
		message = fmt.Sprintf("'%s' is required", e.Field())
	case "len":
		code = "param_invalid_length"
		message = fmt.Sprintf("'%s' is must have a length equal to %s", e.Field(), e.Param())
	case "min":
		code = "param_length_below_minimum"
		message = fmt.Sprintf("'%s' is has a minimun length of %s", e.Field(), e.Param())
	case "max":
		code = "param_length_over_maximum"
		message = fmt.Sprintf("'%s' is has a maximum length of %s", e.Field(), e.Param())
	case "oneof":
		code = "param_is_not_present_in_enum"
		message = fmt.Sprintf("'%s' must be one of: '%s'", e.Field(), strings.Join(strings.Split(e.Param(), " "), "' '"))
	case "email":
		code = "param_is_not_an_email"
		message = fmt.Sprintf("'%s' must be a valid email address", e.Field())
	case "unique":
		code = "param_repeated_values"
		message = fmt.Sprintf("'%s' does not allow repeated values", e.Field())
	default:
		code = "param_is_invalid"
		message = fmt.Sprintf("'%s' is invalid, must meet the requirements of the tag %s", e.Field(), e.Tag())
	}

	return errors.New(message, code)
}
