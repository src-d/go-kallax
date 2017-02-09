package kallax

import (
	"database/sql"
	"database/sql/driver"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
	uuid "github.com/satori/go.uuid"
)

// Model is the base type of the items that are stored
type Model struct {
	ID             ID
	virtualColumns map[string]interface{}
	persisted      bool
	writable       bool
}

// NewModel creates and return a new Model
func NewModel() Model {
	m := Model{
		persisted:      false,
		writable:       true,
		virtualColumns: make(map[string]interface{}),
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
func (m *Model) GetID() ID {
	return m.ID
}

// SetID overrides the ID.
// The ID should not be modified once it has been set and stored in the DB
// WARNING: Not to be used by final users!
func (m *Model) SetID(id ID) {
	m.ID = id
}

// ClearVirtualColumns clears all the previous virtual columns.
func (m *Model) ClearVirtualColumns() {
	m.virtualColumns = make(map[string]interface{})
}

// AddVirtualColumn adds a new virtual column with the given name and value.
func (m *Model) AddVirtualColumn(name string, v interface{}) {
	if m.virtualColumns == nil {
		m.ClearVirtualColumns()
	}
	m.virtualColumns[name] = v
}

// VirtualColumn returns the value of the virtual column with the given column name.
func (m *Model) VirtualColumn(name string) interface{} {
	if m.virtualColumns == nil {
		m.ClearVirtualColumns()
	}
	return m.virtualColumns[name]
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

// ColumnAddresser must be implemented by those values that expose their properties
// under pointers, identified by its property names
type ColumnAddresser interface {
	// ColumnAddress returns a pointer to the object property identified by the
	// column name or an error if that property does not exist
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

// Valuer must be implemented by those object that expose their properties
// identified by its property names
type Valuer interface {
	// Value returns the value under the object property identified by the passed
	// string or an error if that property does not exist
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

// Record is the interface that must be implemented by models that can be stored.
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

// ID is the Kallax identifier type.
type ID uuid.UUID

// NewID returns a new kallax ID, which is a lexically sortable UUID.
// The internal representation is an ULID (https://github.com/oklog/ulid).
func NewID() ID {
	entropy := randPool.Get().(rand.Source)
	id := ID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(entropy)))
	randPool.Put(entropy)

	return id
}

// Scan implements the Scanner interface.
func (id *ID) Scan(src interface{}) error {
	return (*uuid.UUID)(id).Scan(src)
}

// Value implements the Valuer interface.
func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

// IsEmpty returns true if the ID is not set
func (id ID) IsEmpty() bool {
	return uuid.Equal(uuid.UUID(id), uuid.Nil)
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

type virtualColumn struct {
	r   Record
	col string
}

func VirtualColumn(col string, r Record) sql.Scanner {
	return &virtualColumn{r, col}
}

func (c *virtualColumn) Scan(src interface{}) error {
	var id ID
	if err := (&id).Scan(src); err != nil {
		return err
	}

	c.r.AddVirtualColumn(c.col, id)
	return nil
}
