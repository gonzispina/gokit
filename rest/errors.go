package rest

import (
	"fmt"
	"net/http"

	"github.com/gonzispina/gokit/errors"
)

var (
	errInvalidContentType = errors.New("invalid content type", "invalid_content_type")
)

// ErrInvalidStringParam error
func ErrInvalidStringParam(name string) error {
	return errors.New(fmt.Sprintf("'%s' must be a valid string", name), "invalid_param_type")
}

// ErrInvalidBoolParam error
func ErrInvalidBoolParam(name string) error {
	return errors.New(fmt.Sprintf("'%s' must be a valid bool value (true or false)", name), "invalid_param_type")
}

// ErrInvalidNumberParam error
func ErrInvalidNumberParam(name string) error {
	return errors.New(fmt.Sprintf("'%s' must be a valid number", name), "invalid_param_type")
}

// ErrInvalidArrayParam error
func ErrInvalidArrayParam(name string) error {
	return errors.New(fmt.Sprintf("'%s' is not valid", name), "invalid_param_value")
}

// NewError response
func NewError(statusCode int, err error) *Response {
	var code string
	var description string

	if err != nil {
		if sigiErr, ok := err.(errors.Error); ok {
			code = sigiErr.Code()
			description = err.Error()
		}
	}

	return &Response{
		Data:       nil,
		StatusCode: statusCode,
		Err:        description,
		Code:       code,
	}
}

// OK response
func OK(data interface{}) *Response {
	return NewResponse(http.StatusOK, data, nil)
}

// Created response
func Created(data interface{}) *Response {
	return NewResponse(http.StatusCreated, data, nil)
}

// NoContent response
func NoContent() *Response {
	return NewResponse(http.StatusNoContent, nil, nil)
}

// Found redirect response
func Found(redirectURL string) *Response {
	header := map[string]string{
		LocationHeader: redirectURL,
	}
	return NewResponse(http.StatusFound, nil, header)
}

// InternalServerError error response
func InternalServerError() *Response {
	return NewError(http.StatusInternalServerError, nil)
}

// BadRequest error response
func BadRequest(err error) *Response {
	return NewError(http.StatusBadRequest, err)
}

// NotFound error response
func NotFound(err error) *Response {
	return NewError(http.StatusNotFound, err)
}

// RequestEntityTooLarge error response
func RequestEntityTooLarge() *Response {
	return NewError(http.StatusRequestEntityTooLarge, nil)
}

// Unauthorized error response
func Unauthorized(err error) *Response {
	return NewError(http.StatusForbidden, err)
}

// Forbidden error response
func Forbidden() *Response {
	return NewError(http.StatusForbidden, nil)
}

// TooEarly error response
func TooEarly(err error) *Response {
	return NewError(http.StatusTooEarly, err)
}

// PageExpired error response
func PageExpired() *Response {
	return NewError(419, nil)
}
