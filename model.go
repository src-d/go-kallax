package kallax

import (
	"database/sql"
	"database/sql/driver"

	uuid "github.com/satori/go.uuid"
)

// Model contains all the basic fields that make something a model, that is,
// the ID and some internal data used by kallax.
// To make a struct a model, it only needs to have Model embedded.
//
//	type MyModel struct {
//		kallax.Model
//		Foo string
//	}
//
// Custom name for the table can be specified using the struct tag `table` when
// embedding the Model.
//
//	type MyModel struct {
//		kallax.Model `table:"custom_name"`
//	}
//
// Otherwise, the default name of the table is the name of the model converted
// to lower snake case. E.g: MyModel => my_model.
// No pluralization is done right now, but might be done in the future, so
// please, set the name of the tables yourself.
type Model struct {
	ID             ID
	virtualColumns map[string]interface{}
	persisted      bool
	writable       bool
}

// NewModel creates a new Model that is writable, not persisted and identified
// with a newly generated ID.
func NewModel() Model {
	m := Model{
		persisted:      false,
		writable:       true,
		virtualColumns: make(map[string]interface{}),
	}
	m.ID = NewID()
	return m
}

// IsPersisted returns whether the Model has already been persisted to the
// database or not.
func (m *Model) IsPersisted() bool {
	return m.persisted
}

func (m *Model) setPersisted() {
	m.persisted = true
}

// IsWritable returns whether this Model can be saved into the database.
// For example, a model with partially retrieved data is not writable, so
// it is not saved by accident and the data is corrupted. For example, if
// you select only 2 columns out of all the ones the table has, it will not
// be writable.
func (m *Model) IsWritable() bool {
	return m.writable
}

func (m *Model) setWritable(w bool) {
	m.writable = w
}

// GetID returns the ID of the model.
func (m *Model) GetID() ID {
	return m.ID
}

// SetID sets the ID of the model.
// The ID should not be modified once it has been set and stored in the
// database, so use it with caution.
func (m *Model) SetID(id ID) {
	m.ID = id
}

// ClearVirtualColumns clears all the previous virtual columns.
// This method is only intended for internal use. It is only exposed for
// technical reasons.
func (m *Model) ClearVirtualColumns() {
	m.virtualColumns = make(map[string]interface{})
}

// AddVirtualColumn adds a new virtual column with the given name and value.
// This method is only intended for internal use. It is only exposed for
// technical reasons.
func (m *Model) AddVirtualColumn(name string, v interface{}) {
	if m.virtualColumns == nil {
		m.ClearVirtualColumns()
	}
	m.virtualColumns[name] = v
}

// VirtualColumn returns the value of the virtual column with the given column name.
// This method is only intended for internal use. It is only exposed for
// technical reasons.
func (m *Model) VirtualColumn(name string) interface{} {
	if m.virtualColumns == nil {
		m.ClearVirtualColumns()
	}
	return m.virtualColumns[name]
}

// Identifiable must be implemented by those values that can be identified by an ID.
type Identifiable interface {
	// GetID returns the ID.
	GetID() ID
	// SetID sets the ID.
	SetID(id ID)
}

// Persistable must be implemented by those values that can be persisted.
type Persistable interface {
	// IsPersisted returns whether this Model is new in the store or not.
	IsPersisted() bool
	setPersisted()
}

// Writable must be implemented by those values that defines internally
// if they can be sent back to the database to be stored with its changes.
type Writable interface {
	// IsWritable returns whether this Model can be saved into the database.
	IsWritable() bool
	setWritable(bool)
}

// ColumnAddresser provides the pointer addresses of columns.
type ColumnAddresser interface {
	// ColumnAddress returns the pointer to the column value of the given
	// column name, or an error if it does not exist in the model.
	ColumnAddress(string) (interface{}, error)
}

// Relationable can perform operations related to relationships of a record.
type Relationable interface {
	// NewRelationshipRecord returns a new Record for the relationship at the
	// given field.
	NewRelationshipRecord(string) (Record, error)
	// SetRelationship sets the relationship value at the given field.
	SetRelationship(string, interface{}) error
}

// Valuer provides the values for columns.
type Valuer interface {
	// Value returns the value of the given column, or an error if it does not
	// exist in the model.
	Value(string) (interface{}, error)
}

// VirtualColumnContainer contains a collection of virtual columns and
// manages them.
type VirtualColumnContainer interface {
	// ClearVirtualColumns removes all virtual columns.
	ClearVirtualColumns()
	// AddVirtualColumn adds a new virtual column with the given name and value
	AddVirtualColumn(string, interface{})
	// VirtualColumn returns the virtual column with the given column name.
	VirtualColumn(string) interface{}
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

// Record is something that can be stored as a row in the database.
type Record interface {
	Identifiable
	Persistable
	Writable
	Relationable
	ColumnAddresser
	Valuer
	VirtualColumnContainer
}

// ID is the identifier type used in models. It is an UUID, so it should
// be stored as `uuid` in PostgreSQL.
type ID uuid.UUID

// NewID returns a new ID.
func NewID() ID {
	return ID(uuid.NewV4())
}

// Scan implements the Scanner interface.
func (id *ID) Scan(src interface{}) error {
	return (*uuid.UUID)(id).Scan(src)
}

// Value implements the Valuer interface.
func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// IsEmpty returns whether the ID is empty or not. An empty ID means it has not
// been set yet.
func (id ID) IsEmpty() bool {
	return uuid.Equal(uuid.UUID(id), uuid.Nil)
}

// String returns the string representation of the ID.
func (id ID) String() string {
	return uuid.UUID(id).String()
}

type virtualColumn struct {
	r   Record
	col string
}

// VirtualColumn returns a sql.Scanner that will scan the given column as a
// virtual column in the given record.
func VirtualColumn(col string, r Record) sql.Scanner {
	return &virtualColumn{r, col}
}

// Scan implements the scanner interface.
func (c *virtualColumn) Scan(src interface{}) error {
	var id ID
	if err := (&id).Scan(src); err != nil {
		return err
	}

	c.r.AddVirtualColumn(c.col, id)
	return nil
}
