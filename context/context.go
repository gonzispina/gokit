package context

import (
	"context"
	"time"

	"github.com/gonzispina/gokit/uuid"
)

// Upgrade from the standard's lib context to our context
func Upgrade(ctx context.Context) Context {
	return &c{Context: ctx, trackingID: uuid.New()}
}

// Background context
func Background() Context {
	return &c{Context: context.Background(), trackingID: uuid.New()}
}

// WithID derives a context but keeps the tracking ID
func WithID(id string) Context {
	if id == "" {
		id = uuid.New()
	}
	return &c{Context: context.Background(), trackingID: id}
}

// CancelFunc shadow of the standard's lib
type CancelFunc = context.CancelFunc

// WithValue context
func WithValue(ctx Context, key interface{}, value interface{}) Context {
	n := context.WithValue(ctx, key, value)
	return &c{Context: n, trackingID: ctx.TrackingID()}
}

// WithCancel context
func WithCancel(ctx Context) (Context, CancelFunc) {
	n, f := context.WithCancel(ctx)
	return &c{Context: n, trackingID: ctx.TrackingID()}, f
}

// WithTimeout context
func WithTimeout(ctx Context, d time.Duration) (Context, context.CancelFunc) {
	n, f := context.WithTimeout(ctx, d)
	return &c{Context: n, trackingID: ctx.TrackingID()}, f
}

// Merge a context of the standard lib to an existing context
func Merge(newCtx context.Context, oldCtx Context) Context {
	ctx := oldCtx.(*c)
	return &c{Context: newCtx, trackingID: ctx.trackingID}
}

// Context interface
type Context interface {
	context.Context
	TrackingID() string
}

type c struct {
	context.Context
	trackingID string
}

// TrackingID to track everything
func (c c) TrackingID() string {
	return c.trackingID
}
