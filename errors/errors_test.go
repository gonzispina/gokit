package errors_test

import (
	"testing"

	"github.com/gonzispina/gokit/errors"
	"github.com/stretchr/testify/require"
)

func TestErrors(main *testing.T) {
	main.Run("New returns an interface that is compatible with standard library 'error' interface", func(t *testing.T) {
		err := errors.New("An error message", "an_error_code")
		_, ok := err.(error)
		require.True(t, ok)
	})

	main.Run("New lowers all messages", func(t *testing.T) {
		err := errors.New("An error message", "an_error_code")
		require.Equal(t, "an error message", err.Error())
	})

	main.Run("Wrap encapsulates the error type", func(t *testing.T) {
		encapsulatedErr := errors.New("error encapsulated", "an_error_code")
		superEncapsulatedErr := errors.New("super encapsulated error", "an_error_code")
		superDuperEncapsulatedErr := errors.New("super duper encapsulated error", "an_error_code")
		err := errors.New("An error message", "an_error_code").Wrap(encapsulatedErr).Wrap(superEncapsulatedErr).Wrap(superDuperEncapsulatedErr)
		require.True(t, errors.Is(err, encapsulatedErr))
		require.True(t, errors.Is(err, superEncapsulatedErr))
		require.True(t, errors.Is(err, superDuperEncapsulatedErr))
	})

	main.Run("NewWithErr encapsulates the err", func(t *testing.T) {
		encapsulatedErr := errors.New("error encapsulated", "an_error_code")
		err := errors.NewWithErr("an error", "code", encapsulatedErr)
		require.True(t, errors.Is(err, encapsulatedErr))
	})

	main.Run("IsOnly returns true when the error is unique", func(t *testing.T) {
		err := errors.ErrUnknown
		require.True(t, errors.IsOnly(err, err))
	})

	main.Run("IsOnly returns false when there is another error wrapper", func(t *testing.T) {
		encapsulatedErr := errors.New("error encapsulated", "an_error_code")
		err := errors.NewWithErr("an error", "code", encapsulatedErr)
		require.False(t, errors.IsOnly(err, err))
		require.False(t, errors.IsOnly(err, encapsulatedErr))
	})
}
