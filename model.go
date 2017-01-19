package kallax

import (
	"database/sql/driver"

	uuid "github.com/satori/go.uuid"
)

// Model is the base type of the items that are stored
type Model struct {
	ID        ID
	persisted bool
	writable  bool
}

// NewModel creates and return a new Model
func NewModel() Model {
	m := Model{
		persisted: false,
		writable:  true,
	}
	m.ID = NewID()
	return m
}

// IsPersisted returns whether this Model is new in the store or not.
func (m *Model) IsPersisted() bool {
	return m.persisted
}

func (m *Model) setPersisted() {
	m.persisted = true
}

// IsWritable returns whether this Model can be sent back to the database to be
// stored with its changes.
func (m *Model) IsWritable() bool {
	return m.writable
}

func (m *Model) setWritable(w bool) {
	m.writable = w
}

// GetID returns the ID.
func (i *Model) GetID() ID {
	return i.ID
}

// SetID overrides the ID.
// The ID should not be modified once it has been set and stored in the DB
func (i *Model) SetID(id ID) {
	i.ID = id
}

// Identifiable must be implemented by those values that can be identified by an ID
type Identifiable interface {
	// GetID returns the ID.
	GetID() ID
	// SetID overrides the ID.
	SetID(id ID)
}

// Persistable must be implemented by those values that can be persisted
type Persistable interface {
	// IsPersisted returns whether this Model is new in the store or not.
	IsPersisted() bool
	setPersisted()
}

// Writable must be implemented by those values that defines internally
// if they can be sent back to the database to be stored with its changes.
type Writable interface {
	// IsWritable returns whether this Model can be sent back to the database
	// to be stored with its changes.
	IsWritable() bool
	setWritable(bool)
}

// ColumnAddresser must be implemented by those values that exposes its properties
// under pointers, identified by its property names
type ColumnAddresser interface {
	// ColumnAddress returns a pointer to the object property identified by the
	// passed string or an error if that property does not exist
	ColumnAddress(string) (interface{}, error)
}

// Valuer must be implemented by those object that exposes its properties
// identified by its property names
type Valuer interface {
	// Value returns the value under the object property identified by the passed
	// string or an error if that property does not exist
	Value(string) (driver.Value, error)
}

// RecordValues returns the values of a record at the given columns in the same
// order as the columns.
func RecordValues(record Valuer, columns ...string) ([]interface{}, error) {
	var values = make([]interface{}, len(columns))
	for i, col := range columns {
		v, err := record.Value(col)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

// Record is the interface that must be implemented for items that can be stored
type Record interface {
	Identifiable
	Persistable
	Writable
	ColumnAddresser
	Valuer
}

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

func (id ID) String() string {
	return uuid.UUID(id).String()
}
