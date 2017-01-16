package kallax

import (
	"database/sql"
	"database/sql/driver"

	uuid "github.com/satori/go.uuid"
)

type ID uuid.UUID

func (id *ID) Scan(src interface{}) error {
	return (*uuid.UUID)(id).Scan(src)
}

func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

var _ sql.Scanner = (*ID)(nil)
var _ driver.Valuer = (*ID)(nil)

func NewID() ID {
	return ID(uuid.NewV4())
}

func (id ID) IsEmpty() bool {
	return uuid.UUID(id) == uuid.Nil
}

type Model struct {
	ID        ID
	persisted bool
}

func (m *Model) GetID() ID {
	return m.ID
}

func (m *Model) SetID(id ID) {
	m.ID = id
}

func (m *Model) IsPersisted() bool {
	return m.persisted
}

func (m *Model) setPersisted(isPersisted bool) {
	m.persisted = isPersisted
}

type ColumnAddresser interface {
	ColumnAddress(string) (interface{}, error)
}

type Valuer interface {
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

type Identificable interface {
	GetID() ID
	SetID(ID)
}

type Persistable interface {
	IsPersisted() bool
	setPersisted(bool)
}

type Record interface {
	Identificable
	Persistable
	ColumnAddresser
	Valuer
}
