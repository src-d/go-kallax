package kallax

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
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
	virtualColumns map[string]Identifier
	persisted      bool
	writable       bool
}

// NewModel creates a new Model that is writable and not persisted.
func NewModel() Model {
	return Model{
		persisted:      false,
		writable:       true,
		virtualColumns: make(map[string]Identifier),
	}
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

// ClearVirtualColumns clears all the previous virtual columns.
// This method is only intended for internal use. It is only exposed for
// technical reasons.
func (m *Model) ClearVirtualColumns() {
	m.virtualColumns = make(map[string]Identifier)
}

// AddVirtualColumn adds a new virtual column with the given name and value.
// This method is only intended for internal use. It is only exposed for
// technical reasons.
func (m *Model) AddVirtualColumn(name string, v Identifier) {
	if m.virtualColumns == nil {
		m.ClearVirtualColumns()
	}
	m.virtualColumns[name] = v
}

// VirtualColumn returns the value of the virtual column with the given column name.
// This method is only intended for internal use. It is only exposed for
// technical reasons.
func (m *Model) VirtualColumn(name string) Identifier {
	if m.virtualColumns == nil {
		m.ClearVirtualColumns()
	}
	return m.virtualColumns[name]
}

// Identifier is a type used to identify a model.
type Identifier interface {
	sql.Scanner
	driver.Valuer
	// Equals reports whether the identifier and the given one are equal.
	Equals(Identifier) bool
	// IsEmpty returns whether the ID is empty or not.
	IsEmpty() bool
	// Raw returns the internal value of the identifier.
	Raw() interface{}
}

// Identifiable must be implemented by those values that can be identified by an ID.
type Identifiable interface {
	// GetID returns the ID.
	GetID() Identifier
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
	AddVirtualColumn(string, Identifier)
	// VirtualColumn returns the virtual column with the given column name.
	VirtualColumn(string) Identifier
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

var randPool = &sync.Pool{
	New: func() interface{} {
		return rand.NewSource(time.Now().UnixNano())
	},
}

// ULID is an ID type provided by kallax that is a lexically sortable UUID.
// The internal representation is an ULID (https://github.com/oklog/ulid).
// It already implements sql.Scanner and driver.Valuer, so it's perfectly
// safe for database usage.
type ULID uuid.UUID

// NewULID returns a new ULID, which is a lexically sortable UUID.
func NewULID() ULID {
	entropy := randPool.Get().(rand.Source)
	id := ULID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(entropy)))
	randPool.Put(entropy)

	return id
}

// Scan implements the Scanner interface.
func (id *ULID) Scan(src interface{}) error {
	return (*uuid.UUID)(id).Scan(src)
}

// Value implements the Valuer interface.
func (id ULID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// IsEmpty returns whether the ID is empty or not. An empty ID means it has not
// been set yet.
func (id ULID) IsEmpty() bool {
	return uuid.Equal(uuid.UUID(id), uuid.Nil)
}

// String returns the string representation of the ID.
func (id ULID) String() string {
	return uuid.UUID(id).String()
}

// Equals reports whether the ID and the given one are equals.
func (id ULID) Equals(other Identifier) bool {
	v, ok := other.(*ULID)
	if !ok {
		return false
	}

	return uuid.Equal(uuid.UUID(id), uuid.UUID(*v))
}

// Raw returns the underlying raw value.
func (id ULID) Raw() interface{} {
	return id
}

// NumericID is a wrapper for int64 that implements the Identifier interface.
// You don't need to actually use this as a type in your model. They will be
// automatically converted to and from in the generated code.
type NumericID int64

// Scan implements the Scanner interface.
func (id *NumericID) Scan(src interface{}) error {
	switch src := src.(type) {
	case int64:
		*(*int64)(id) = src
	default:
		return fmt.Errorf("kallax: cannot scan value of type %T into a numeric ID", src)
	}

	return nil
}

// Value implements the Valuer interface.
func (id NumericID) Value() (driver.Value, error) {
	return int64(id), nil
}

// IsEmpty returns whether the ID is empty or not. An empty ID means it has not
// been set yet.
func (id NumericID) IsEmpty() bool {
	return int64(id) == 0
}

// String returns the string representation of the ID.
func (id NumericID) String() string {
	return fmt.Sprint(int64(id))
}

// Equals reports whether the ID and the given one are equals.
func (id NumericID) Equals(other Identifier) bool {
	v, ok := other.(*NumericID)
	if !ok {
		return false
	}

	return int64(id) == int64(*v)
}

// Raw returns the underlying raw value.
func (id NumericID) Raw() interface{} {
	return id
}

// UUID is a wrapper type for uuid.UUID that implements the Identifier
// interface.
// You don't need to actually use this as a type in your model. They will be
// automatically converted to and from in the generated code.
type UUID uuid.UUID

// Scan implements the Scanner interface.
func (id *UUID) Scan(src interface{}) error {
	return (*uuid.UUID)(id).Scan(src)
}

// Value implements the Valuer interface.
func (id UUID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// IsEmpty returns whether the ID is empty or not. An empty ID means it has not
// been set yet.
func (id UUID) IsEmpty() bool {
	return uuid.Equal(uuid.UUID(id), uuid.Nil)
}

// String returns the string representation of the ID.
func (id UUID) String() string {
	return uuid.UUID(id).String()
}

// Equals reports whether the ID and the given one are equals.
func (id UUID) Equals(other Identifier) bool {
	v, ok := other.(*UUID)
	if !ok {
		return false
	}

	return uuid.Equal(uuid.UUID(id), uuid.UUID(*v))
}

// Raw returns the underlying raw value.
func (id UUID) Raw() interface{} {
	return id
}

type virtualColumn struct {
	r   Record
	col string
	id  Identifier
}

// VirtualColumn returns a sql.Scanner that will scan the given column as a
// virtual column in the given record.
func VirtualColumn(col string, r Record, id Identifier) sql.Scanner {
	return &virtualColumn{r, col, id}
}

// Scan implements the scanner interface.
func (c *virtualColumn) Scan(src interface{}) error {
	if err := c.id.Scan(src); err != nil {
		return err
	}

	c.r.AddVirtualColumn(c.col, c.id)
	return nil
}
