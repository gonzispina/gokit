package errors

import (
	"errors"
	"reflect"
	"strings"
)

// ErrUnknown is used primarily by tests
var ErrUnknown = New("unknown error", "errors_unknown")

// Is calls to standard's library errors.Is()
// This is done to avoid double imports
var Is = errors.Is

// As calls to standard's library errors.As()
// This is done to avoid double imports
var As = errors.As

// OneOf to tell if the error exists in the array
func OneOf(err error, targets ...error) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		if Is(err, target) {
			return true
		}
	}
	return false
}

// IsOnly tells if the error is one specific type only
func IsOnly(err, target error) bool {
	if target == nil {
		return err == target
	}
	isComparable := reflect.TypeOf(target).Comparable()
	if isComparable && err != target {
		return false
	}
	if x, ok := err.(interface{ Is(error) bool }); ok && !x.Is(target) {
		return false
	}
	if err = errors.Unwrap(err); err != nil {
		return false
	}
	return true
}

// Error representation
type Error interface {
	Error() string
	Code() string
	Wrap(err error) Error
	Unwrap() error
	Is(error) bool
}

// New creates a new error
func New(msg string, code string) Error {
	return err{msg: msg, code: code}
}

// NewWithErr so it is not necessary to call New(msg).Wrap(err)
func NewWithErr(msg, code string, ext error) Error {
	return New(msg, code).Wrap(ext)
}

type err struct {
	msg  string
	code string
	err  Error
}

// Error message
func (e err) Error() string {
	return strings.ToLower(e.msg)
}

// Code of the error
func (e err) Code() string {
	return strings.ToLower(e.code)
}

// Wrap an error
func (e err) Wrap(err error) Error {
	if err == nil {
		return e
	}
	if e.err != nil {
		e.err = e.err.Wrap(err)
		return e
	}
	n := New(err.Error(), e.Code())
	e.err = n
	return e
}

// Unwrap the error
func (e err) Unwrap() error {
	return e.err
}

// Is tells if the error is or not equal
func (e err) Is(err error) bool {
	return e.msg == err.Error()
}
