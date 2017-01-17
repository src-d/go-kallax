package kallax

import (
	"github.com/src-d/go-kallax/behaviours"
	"github.com/src-d/go-kallax/common"
)

// Model is the base type of the items that are stored
type Model struct {
	behaviours.Identificator
	behaviours.TimestampDates
	persisted bool
	writable  bool
}

// NewModel creates and return a new Model
func NewModel() Model {
	m := Model{
		persisted: false,
		writable:  true,
	}
	m.ID = common.NewID()
	return m
}

// IsPersisted returns whether this Model is new in the store or not.
func (m *Model) IsPersisted() bool {
	return m.persisted
}

func (m *Model) setPersisted() {
	m.persisted = true
}

// IsWritable returns whether this Model can be sent back to the database to be stored with its changes.
func (m *Model) IsWritable() bool {
	return m.writable
}

func (m *Model) setWritable(w bool) {
	m.writable = w
}

// Persistable must be implemented by those values that can be persisted
type Persistable interface {
	// IsPersisted returns whether this Model is new in the store or not.
	IsPersisted() bool
	setPersisted()
}

// Writable must be implemented by those values that defines internally
//  if they can be sent back to the database to be stored with its changes.
type Writable interface {
	IsWritable() bool
	setWritable(bool)
}

// ColumnAddresser must be implemented by those values that exposes its properties
//  under pointers, identified by its property names
type ColumnAddresser interface {
	// ColumnAddress returns a pointer to the object property identified by the passed string
	//  or an error if that property does not exist
	ColumnAddress(string) (interface{}, error)
}

// Valuer must be implemented by those object that exposes its properties
//  identified by its property names
type Valuer interface {
	// Value returns the value under the object property identified by the passed string
	//  or an error if that property does not exist
	Value(string) (interface{}, error)
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
	behaviours.Identifiable
	behaviours.Timestampable
	Persistable
	Writable
	ColumnAddresser
	Valuer
}
