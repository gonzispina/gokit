package uuid

import (
	"github.com/google/uuid"
)

// New string uuid
func New() string {
	return uuid.New().String()
}
