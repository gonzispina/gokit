package mongo

import (
	"github.com/gonzispina/gokit/context"
	"go.mongodb.org/mongo-driver/mongo"
)

// Session shadow so the API remains the same
type Session = mongo.Session

// SessionContext shadow so the API remains the same
type SessionContext = context.Context

// Pipeline shadow so the API remains the same
type Pipeline = mongo.Pipeline

var (
	// ErrNoDocuments shadow so the API remains the same
	ErrNoDocuments = mongo.ErrNoDocuments
)

// SingleResult shadow so the API remains the same
type SingleResult = mongo.SingleResult
