package common

import (
	"database/sql/driver"

	"github.com/satori/go.uuid"
)

// ID represents the Kallax identificator type. Its underlying type is an UUID
type ID uuid.UUID

// NewID returns a new kallax ID
func NewID() ID {
	return ID(uuid.NewV1())
}

// Scan sets the uuid value from the passed param
func (id *ID) Scan(src interface{}) error {
	return (*uuid.UUID)(id).Scan(src)
}

// Value returns the uuid value
func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// IsEmpty returns true if the ID is set
func (id ID) IsEmpty() bool {
	return uuid.Equal(uuid.UUID(id), uuid.Nil)
}
