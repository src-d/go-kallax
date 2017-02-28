// IMPORTANT! This is auto generated code by https://github.com/src-d/go-kallax
// Please, do not touch the code below, and if you do, do it under your own
// risk. Take into account that all the code you write here will be completely
// erased from earth the next time you generate the kallax models.
package benchmark

import (
	"database/sql"
	"fmt"

	"gopkg.in/src-d/go-kallax.v1"
	"gopkg.in/src-d/go-kallax.v1/types"
)

var _ types.SQLType
var _ fmt.Formatter

// NewPerson returns a new instance of Person.
func NewPerson() (record *Person) {
	return new(Person)
}

// GetID returns the primary key of the model.
func (r *Person) GetID() kallax.Identifier {
	return (*kallax.NumericID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *Person) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.NumericID)(&r.ID), nil
	case "name":
		return &r.Name, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Person: %s", col)
	}
}

// Value returns the value of the given column.
func (r *Person) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "name":
		return r.Name, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Person: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *Person) NewRelationshipRecord(field string) (kallax.Record, error) {
	switch field {
	case "Pets":
		return new(Pet), nil

	}
	return nil, fmt.Errorf("kallax: model Person has no relationship %s", field)
}

// SetRelationship sets the given relationship in the given field.
func (r *Person) SetRelationship(field string, rel interface{}) error {
	switch field {
	case "Pets":
		records, ok := rel.([]kallax.Record)
		if !ok {
			return fmt.Errorf("kallax: relationship field %s needs a collection of records, not %T", field, rel)
		}

		r.Pets = make([]*Pet, len(records))
		for i, record := range records {
			rel, ok := record.(*Pet)
			if !ok {
				return fmt.Errorf("kallax: element of type %T cannot be added to relationship %s", record, field)
			}
			r.Pets[i] = rel
		}
		return nil

	}
	return fmt.Errorf("kallax: model Person has no relationship %s", field)
}

// PersonStore is the entity to access the records of the type Person
// in the database.
type PersonStore struct {
	*kallax.Store
}

// NewPersonStore creates a new instance of PersonStore
// using a SQL database.
func NewPersonStore(db *sql.DB) *PersonStore {
	return &PersonStore{kallax.NewStore(db)}
}

func (s *PersonStore) relationshipRecords(record *Person) []kallax.RecordWithSchema {
	record.ClearVirtualColumns()
	var records []kallax.RecordWithSchema

	for _, rec := range record.Pets {
		rec.ClearVirtualColumns()
		rec.AddVirtualColumn("person_id", record.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema.Pet.BaseSchema,
			rec,
		})
	}

	return records
}

// Insert inserts a Person in the database. A non-persisted object is
// required for this operation.
func (s *PersonStore) Insert(record *Person) error {

	records := s.relationshipRecords(record)
	if len(records) > 0 {
		return s.Store.Transaction(func(s *kallax.Store) error {
			if err := s.Insert(Schema.Person.BaseSchema, record); err != nil {
				return err
			}

			for _, r := range records {
				if err := kallax.ApplyBeforeEvents(r.Record); err != nil {
					return err
				}
				persisted := r.Record.IsPersisted()

				if _, err := s.Save(r.Schema, r.Record); err != nil {
					return err
				}

				if err := kallax.ApplyAfterEvents(r.Record, persisted); err != nil {
					return err
				}
			}

			return nil
		})
	}

	return s.Store.Insert(Schema.Person.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *PersonStore) Update(record *Person, cols ...kallax.SchemaField) (updated int64, err error) {

	records := s.relationshipRecords(record)
	if len(records) > 0 {
		err = s.Store.Transaction(func(s *kallax.Store) error {
			updated, err = s.Update(Schema.Person.BaseSchema, record, cols...)
			if err != nil {
				return err
			}

			for _, r := range records {
				if err := kallax.ApplyBeforeEvents(r.Record); err != nil {
					return err
				}
				persisted := r.Record.IsPersisted()

				if _, err := s.Save(r.Schema, r.Record); err != nil {
					return err
				}

				if err := kallax.ApplyAfterEvents(r.Record, persisted); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return 0, err
		}

		return updated, nil
	}

	return s.Store.Update(Schema.Person.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *PersonStore) Save(record *Person) (updated bool, err error) {
	if !record.IsPersisted() {
		return false, s.Insert(record)
	}

	rowsUpdated, err := s.Update(record)
	if err != nil {
		return false, err
	}

	return rowsUpdated > 0, nil
}

// Delete removes the given record from the database.
func (s *PersonStore) Delete(record *Person) error {

	return s.Store.Delete(Schema.Person.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *PersonStore) Find(q *PersonQuery) (*PersonResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewPersonResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *PersonStore) MustFind(q *PersonQuery) *PersonResultSet {
	return NewPersonResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *PersonStore) Count(q *PersonQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *PersonStore) MustCount(q *PersonQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *PersonStore) FindOne(q *PersonQuery) (*Person, error) {
	q.Limit(1)
	q.Offset(0)
	rs, err := s.Find(q)
	if err != nil {
		return nil, err
	}

	if !rs.Next() {
		return nil, kallax.ErrNotFound
	}

	record, err := rs.Get()
	if err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return record, nil
}

// MustFindOne returns the first row retrieved by the given query. It panics
// if there is an error or if there are no rows.
func (s *PersonStore) MustFindOne(q *PersonQuery) *Person {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the Person with the data in the database and
// makes it writable.
func (s *PersonStore) Reload(record *Person) error {
	return s.Store.Reload(Schema.Person.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *PersonStore) Transaction(callback func(*PersonStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&PersonStore{store})
	})
}

// RemovePets removes the given items of the Pets field of the
// model. If no items are given, it removes all of them.
// The items will also be removed from the passed record inside this method.
func (s *PersonStore) RemovePets(record *Person, deleted ...*Pet) error {
	var updated []*Pet
	var clear bool
	if len(deleted) == 0 {
		clear = true
		deleted = record.Pets
		if len(deleted) == 0 {
			return nil
		}
	}

	if len(deleted) > 1 {
		err := s.Store.Transaction(func(s *kallax.Store) error {
			for _, d := range deleted {
				var r kallax.Record = d

				if beforeDeleter, ok := r.(kallax.BeforeDeleter); ok {
					if err := beforeDeleter.BeforeDelete(); err != nil {
						return err
					}
				}

				if err := s.Delete(Schema.Pet.BaseSchema, d); err != nil {
					return err
				}

				if afterDeleter, ok := r.(kallax.AfterDeleter); ok {
					if err := afterDeleter.AfterDelete(); err != nil {
						return err
					}
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		if clear {
			record.Pets = nil
			return nil
		}
	} else {
		var r kallax.Record = deleted[0]
		if beforeDeleter, ok := r.(kallax.BeforeDeleter); ok {
			if err := beforeDeleter.BeforeDelete(); err != nil {
				return err
			}
		}

		var err error
		if afterDeleter, ok := r.(kallax.AfterDeleter); ok {
			err = s.Store.Transaction(func(s *kallax.Store) error {
				err := s.Delete(Schema.Pet.BaseSchema, r)
				if err != nil {
					return err
				}

				return afterDeleter.AfterDelete()
			})
		} else {
			err = s.Store.Delete(Schema.Pet.BaseSchema, deleted[0])
		}

		if err != nil {
			return err
		}
	}

	for _, r := range record.Pets {
		var found bool
		for _, d := range deleted {
			if d.GetID().Equals(r.GetID()) {
				found = true
				break
			}
		}
		if !found {
			updated = append(updated, r)
		}
	}
	record.Pets = updated
	return nil
}

// PersonQuery is the object used to create queries for the Person
// entity.
type PersonQuery struct {
	*kallax.BaseQuery
}

// NewPersonQuery returns a new instance of PersonQuery.
func NewPersonQuery() *PersonQuery {
	return &PersonQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.Person.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *PersonQuery) Select(columns ...kallax.SchemaField) *PersonQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *PersonQuery) SelectNot(columns ...kallax.SchemaField) *PersonQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *PersonQuery) Copy() *PersonQuery {
	return &PersonQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *PersonQuery) Order(cols ...kallax.ColumnOrder) *PersonQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *PersonQuery) BatchSize(size uint64) *PersonQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *PersonQuery) Limit(n uint64) *PersonQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *PersonQuery) Offset(n uint64) *PersonQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *PersonQuery) Where(cond kallax.Condition) *PersonQuery {
	q.BaseQuery.Where(cond)
	return q
}

func (q *PersonQuery) WithPets(cond kallax.Condition) *PersonQuery {
	q.AddRelation(Schema.Pet.BaseSchema, "Pets", kallax.OneToMany, cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values, it will do nothing
func (q *PersonQuery) FindByID(v ...int64) *PersonQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.Person.ID, values...))
}

// FindByName adds a new filter to the query that will require that
// the Name property is equal to the passed value
func (q *PersonQuery) FindByName(v string) *PersonQuery {
	return q.Where(kallax.Eq(Schema.Person.Name, v))
}

// PersonResultSet is the set of results returned by a query to the
// database.
type PersonResultSet struct {
	ResultSet kallax.ResultSet
	last      *Person
	lastErr   error
}

// NewPersonResultSet creates a new result set for rows of the type
// Person.
func NewPersonResultSet(rs kallax.ResultSet) *PersonResultSet {
	return &PersonResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *PersonResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.Person.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*Person)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *Person")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *PersonResultSet) Get() (*Person, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *PersonResultSet) ForEach(fn func(*Person) error) error {
	for rs.Next() {
		record, err := rs.Get()
		if err != nil {
			return err
		}

		if err := fn(record); err != nil {
			if err == kallax.ErrStop {
				return rs.Close()
			}

			return err
		}
	}
	return nil
}

// All returns all records on the result set and closes the result set.
func (rs *PersonResultSet) All() ([]*Person, error) {
	var result []*Person
	for rs.Next() {
		record, err := rs.Get()
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

// One returns the first record on the result set and closes the result set.
func (rs *PersonResultSet) One() (*Person, error) {
	if !rs.Next() {
		return nil, kallax.ErrNotFound
	}

	record, err := rs.Get()
	if err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return record, nil
}

// Err returns the last error occurred.
func (rs *PersonResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *PersonResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewPet returns a new instance of Pet.
func NewPet() (record *Pet) {
	return new(Pet)
}

// GetID returns the primary key of the model.
func (r *Pet) GetID() kallax.Identifier {
	return (*kallax.NumericID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *Pet) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.NumericID)(&r.ID), nil
	case "name":
		return &r.Name, nil
	case "kind":
		return &r.Kind, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Pet: %s", col)
	}
}

// Value returns the value of the given column.
func (r *Pet) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "name":
		return r.Name, nil
	case "kind":
		return (string)(r.Kind), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Pet: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *Pet) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model Pet has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *Pet) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model Pet has no relationships")
}

// PetStore is the entity to access the records of the type Pet
// in the database.
type PetStore struct {
	*kallax.Store
}

// NewPetStore creates a new instance of PetStore
// using a SQL database.
func NewPetStore(db *sql.DB) *PetStore {
	return &PetStore{kallax.NewStore(db)}
}

// Insert inserts a Pet in the database. A non-persisted object is
// required for this operation.
func (s *PetStore) Insert(record *Pet) error {

	return s.Store.Insert(Schema.Pet.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *PetStore) Update(record *Pet, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.Pet.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *PetStore) Save(record *Pet) (updated bool, err error) {
	if !record.IsPersisted() {
		return false, s.Insert(record)
	}

	rowsUpdated, err := s.Update(record)
	if err != nil {
		return false, err
	}

	return rowsUpdated > 0, nil
}

// Delete removes the given record from the database.
func (s *PetStore) Delete(record *Pet) error {

	return s.Store.Delete(Schema.Pet.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *PetStore) Find(q *PetQuery) (*PetResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewPetResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *PetStore) MustFind(q *PetQuery) *PetResultSet {
	return NewPetResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *PetStore) Count(q *PetQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *PetStore) MustCount(q *PetQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *PetStore) FindOne(q *PetQuery) (*Pet, error) {
	q.Limit(1)
	q.Offset(0)
	rs, err := s.Find(q)
	if err != nil {
		return nil, err
	}

	if !rs.Next() {
		return nil, kallax.ErrNotFound
	}

	record, err := rs.Get()
	if err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return record, nil
}

// MustFindOne returns the first row retrieved by the given query. It panics
// if there is an error or if there are no rows.
func (s *PetStore) MustFindOne(q *PetQuery) *Pet {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the Pet with the data in the database and
// makes it writable.
func (s *PetStore) Reload(record *Pet) error {
	return s.Store.Reload(Schema.Pet.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *PetStore) Transaction(callback func(*PetStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&PetStore{store})
	})
}

// PetQuery is the object used to create queries for the Pet
// entity.
type PetQuery struct {
	*kallax.BaseQuery
}

// NewPetQuery returns a new instance of PetQuery.
func NewPetQuery() *PetQuery {
	return &PetQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.Pet.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *PetQuery) Select(columns ...kallax.SchemaField) *PetQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *PetQuery) SelectNot(columns ...kallax.SchemaField) *PetQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *PetQuery) Copy() *PetQuery {
	return &PetQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *PetQuery) Order(cols ...kallax.ColumnOrder) *PetQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *PetQuery) BatchSize(size uint64) *PetQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *PetQuery) Limit(n uint64) *PetQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *PetQuery) Offset(n uint64) *PetQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *PetQuery) Where(cond kallax.Condition) *PetQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values, it will do nothing
func (q *PetQuery) FindByID(v ...int64) *PetQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.Pet.ID, values...))
}

// FindByName adds a new filter to the query that will require that
// the Name property is equal to the passed value
func (q *PetQuery) FindByName(v string) *PetQuery {
	return q.Where(kallax.Eq(Schema.Pet.Name, v))
}

// FindByKind adds a new filter to the query that will require that
// the Kind property is equal to the passed value
func (q *PetQuery) FindByKind(v PetKind) *PetQuery {
	return q.Where(kallax.Eq(Schema.Pet.Kind, v))
}

// PetResultSet is the set of results returned by a query to the
// database.
type PetResultSet struct {
	ResultSet kallax.ResultSet
	last      *Pet
	lastErr   error
}

// NewPetResultSet creates a new result set for rows of the type
// Pet.
func NewPetResultSet(rs kallax.ResultSet) *PetResultSet {
	return &PetResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *PetResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.Pet.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*Pet)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *Pet")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *PetResultSet) Get() (*Pet, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *PetResultSet) ForEach(fn func(*Pet) error) error {
	for rs.Next() {
		record, err := rs.Get()
		if err != nil {
			return err
		}

		if err := fn(record); err != nil {
			if err == kallax.ErrStop {
				return rs.Close()
			}

			return err
		}
	}
	return nil
}

// All returns all records on the result set and closes the result set.
func (rs *PetResultSet) All() ([]*Pet, error) {
	var result []*Pet
	for rs.Next() {
		record, err := rs.Get()
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

// One returns the first record on the result set and closes the result set.
func (rs *PetResultSet) One() (*Pet, error) {
	if !rs.Next() {
		return nil, kallax.ErrNotFound
	}

	record, err := rs.Get()
	if err != nil {
		return nil, err
	}

	if err := rs.Close(); err != nil {
		return nil, err
	}

	return record, nil
}

// Err returns the last error occurred.
func (rs *PetResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *PetResultSet) Close() error {
	return rs.ResultSet.Close()
}

type schema struct {
	Person *schemaPerson
	Pet    *schemaPet
}

type schemaPerson struct {
	*kallax.BaseSchema
	ID   kallax.SchemaField
	Name kallax.SchemaField
}

type schemaPet struct {
	*kallax.BaseSchema
	ID   kallax.SchemaField
	Name kallax.SchemaField
	Kind kallax.SchemaField
}

var Schema = &schema{
	Person: &schemaPerson{
		BaseSchema: kallax.NewBaseSchema(
			"people",
			"__person",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{
				"Pets": kallax.NewForeignKey("person_id", false),
			},
			func() kallax.Record {
				return new(Person)
			},
			true,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("name"),
		),
		ID:   kallax.NewSchemaField("id"),
		Name: kallax.NewSchemaField("name"),
	},
	Pet: &schemaPet{
		BaseSchema: kallax.NewBaseSchema(
			"pets",
			"__pet",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(Pet)
			},
			true,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("name"),
			kallax.NewSchemaField("kind"),
		),
		ID:   kallax.NewSchemaField("id"),
		Name: kallax.NewSchemaField("name"),
		Kind: kallax.NewSchemaField("kind"),
	},
}
