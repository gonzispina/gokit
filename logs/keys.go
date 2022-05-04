package logs

import (
	"github.com/gonzispina/gokit/context"
	"go.uber.org/zap"
)

// Field for logs
type Field = zap.Field

func addTrackingID(ctx context.Context, fields ...Field) []Field {
	res := []Field{zap.String("trackingId", ctx.TrackingID())}
	if len(fields) == 0 {
		return res
	}
	return append(res, fields...)
}

// Error transforms an error into a zap field
func Error(err error) Field {
	return zap.Error(err)
}

// Bytes receives the stack
func Bytes(value []byte) Field {
	return zap.String("bytes", string(value))
}

// UserID returns a zap field for logging
func UserID(value string) Field {
	return zap.String("userId", value)
}

// ReferenceID is a generic used to add information to logs
func ReferenceID(value string) Field {
	return zap.String("referenceId", value)
}
