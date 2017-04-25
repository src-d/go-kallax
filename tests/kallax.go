// IMPORTANT! This is auto generated code by https://github.com/src-d/go-kallax
// Please, do not touch the code below, and if you do, do it under your own
// risk. Take into account that all the code you write here will be completely
// erased from earth the next time you generate the kallax models.
package tests

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"gopkg.in/src-d/go-kallax.v1"
	"gopkg.in/src-d/go-kallax.v1/tests/fixtures"
	"gopkg.in/src-d/go-kallax.v1/types"
)

var _ types.SQLType
var _ fmt.Formatter

// NewCar returns a new instance of Car.
func NewCar(model string, owner *Person) (record *Car) {
	return newCar(model, owner)
}

// GetID returns the primary key of the model.
func (r *Car) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *Car) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "owner_id":
		return types.Nullable(kallax.VirtualColumn("owner_id", r, new(kallax.NumericID))), nil
	case "model_name":
		return &r.ModelName, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Car: %s", col)
	}
}

// Value returns the value of the given column.
func (r *Car) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "owner_id":
		return r.Model.VirtualColumn(col), nil
	case "model_name":
		return r.ModelName, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Car: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *Car) NewRelationshipRecord(field string) (kallax.Record, error) {
	switch field {
	case "Owner":
		return new(Person), nil

	}
	return nil, fmt.Errorf("kallax: model Car has no relationship %s", field)
}

// SetRelationship sets the given relationship in the given field.
func (r *Car) SetRelationship(field string, rel interface{}) error {
	switch field {
	case "Owner":
		val, ok := rel.(*Person)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Owner", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Owner = val
		}

		return nil

	}
	return fmt.Errorf("kallax: model Car has no relationship %s", field)
}

// CarStore is the entity to access the records of the type Car
// in the database.
type CarStore struct {
	*kallax.Store
}

// NewCarStore creates a new instance of CarStore
// using a SQL database.
func NewCarStore(db *sql.DB) *CarStore {
	return &CarStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *CarStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *CarStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

func (s *CarStore) inverseRecords(record *Car) []kallax.RecordWithSchema {
	record.ClearVirtualColumns()
	var records []kallax.RecordWithSchema

	if record.Owner != nil {
		record.AddVirtualColumn("owner_id", record.Owner.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.Person.BaseSchema,
			Record: record.Owner,
		})
	}

	return records
}

// Insert inserts a Car in the database. A non-persisted object is
// required for this operation.
func (s *CarStore) Insert(record *Car) error {

	if err := record.BeforeSave(); err != nil {
		return err
	}

	inverseRecords := s.inverseRecords(record)

	if len(inverseRecords) > 0 {
		return s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			if err := s.Insert(Schema.Car.BaseSchema, record); err != nil {
				return err
			}

			if err := record.AfterSave(); err != nil {
				return err
			}

			return nil
		})
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		if err := s.Insert(Schema.Car.BaseSchema, record); err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *CarStore) Update(record *Car, cols ...kallax.SchemaField) (updated int64, err error) {

	if err := record.BeforeSave(); err != nil {
		return 0, err
	}

	inverseRecords := s.inverseRecords(record)

	if len(inverseRecords) > 0 {
		err = s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			updated, err = s.Update(Schema.Car.BaseSchema, record, cols...)
			if err != nil {
				return err
			}

			if err := record.AfterSave(); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return 0, err
		}

		return updated, nil
	}

	err = s.Store.Transaction(func(s *kallax.Store) error {
		updated, err = s.Update(Schema.Car.BaseSchema, record, cols...)
		if err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return updated, nil

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *CarStore) Save(record *Car) (updated bool, err error) {
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
func (s *CarStore) Delete(record *Car) error {

	if err := record.BeforeDelete(); err != nil {
		return err
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		err := s.Delete(Schema.Car.BaseSchema, record)
		if err != nil {
			return err
		}

		return record.AfterDelete()
	})

}

// Find returns the set of results for the given query.
func (s *CarStore) Find(q *CarQuery) (*CarResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewCarResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *CarStore) MustFind(q *CarQuery) *CarResultSet {
	return NewCarResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *CarStore) Count(q *CarQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *CarStore) MustCount(q *CarQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *CarStore) FindOne(q *CarQuery) (*Car, error) {
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
func (s *CarStore) MustFindOne(q *CarQuery) *Car {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the Car with the data in the database and
// makes it writable.
func (s *CarStore) Reload(record *Car) error {
	return s.Store.Reload(Schema.Car.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *CarStore) Transaction(callback func(*CarStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&CarStore{store})
	})
}

// CarQuery is the object used to create queries for the Car
// entity.
type CarQuery struct {
	*kallax.BaseQuery
}

// NewCarQuery returns a new instance of CarQuery.
func NewCarQuery() *CarQuery {
	return &CarQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.Car.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *CarQuery) Select(columns ...kallax.SchemaField) *CarQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *CarQuery) SelectNot(columns ...kallax.SchemaField) *CarQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *CarQuery) Copy() *CarQuery {
	return &CarQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *CarQuery) Order(cols ...kallax.ColumnOrder) *CarQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *CarQuery) BatchSize(size uint64) *CarQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *CarQuery) Limit(n uint64) *CarQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *CarQuery) Offset(n uint64) *CarQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *CarQuery) Where(cond kallax.Condition) *CarQuery {
	q.BaseQuery.Where(cond)
	return q
}

func (q *CarQuery) WithOwner() *CarQuery {
	q.AddRelation(Schema.Person.BaseSchema, "Owner", kallax.OneToOne, nil)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *CarQuery) FindByID(v ...kallax.ULID) *CarQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.Car.ID, values...))
}

// FindByOwner adds a new filter to the query that will require that
// the foreign key of Owner is equal to the passed value.
func (q *CarQuery) FindByOwner(v int64) *CarQuery {
	return q.Where(kallax.Eq(Schema.Car.OwnerFK, v))
}

// FindByModelName adds a new filter to the query that will require that
// the ModelName property is equal to the passed value.
func (q *CarQuery) FindByModelName(v string) *CarQuery {
	return q.Where(kallax.Eq(Schema.Car.ModelName, v))
}

// CarResultSet is the set of results returned by a query to the
// database.
type CarResultSet struct {
	ResultSet kallax.ResultSet
	last      *Car
	lastErr   error
}

// NewCarResultSet creates a new result set for rows of the type
// Car.
func NewCarResultSet(rs kallax.ResultSet) *CarResultSet {
	return &CarResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *CarResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.Car.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*Car)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *Car")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *CarResultSet) Get() (*Car, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *CarResultSet) ForEach(fn func(*Car) error) error {
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
func (rs *CarResultSet) All() ([]*Car, error) {
	var result []*Car
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
func (rs *CarResultSet) One() (*Car, error) {
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
func (rs *CarResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *CarResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewEventsAllFixture returns a new instance of EventsAllFixture.
func NewEventsAllFixture() (record *EventsAllFixture) {
	return newEventsAllFixture()
}

// GetID returns the primary key of the model.
func (r *EventsAllFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *EventsAllFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "checks":
		return types.JSON(&r.Checks), nil
	case "must_fail_before":
		return types.JSON(&r.MustFailBefore), nil
	case "must_fail_after":
		return types.JSON(&r.MustFailAfter), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in EventsAllFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *EventsAllFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "checks":
		return types.JSON(r.Checks), nil
	case "must_fail_before":
		return types.JSON(r.MustFailBefore), nil
	case "must_fail_after":
		return types.JSON(r.MustFailAfter), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in EventsAllFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *EventsAllFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model EventsAllFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *EventsAllFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model EventsAllFixture has no relationships")
}

// EventsAllFixtureStore is the entity to access the records of the type EventsAllFixture
// in the database.
type EventsAllFixtureStore struct {
	*kallax.Store
}

// NewEventsAllFixtureStore creates a new instance of EventsAllFixtureStore
// using a SQL database.
func NewEventsAllFixtureStore(db *sql.DB) *EventsAllFixtureStore {
	return &EventsAllFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *EventsAllFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *EventsAllFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a EventsAllFixture in the database. A non-persisted object is
// required for this operation.
func (s *EventsAllFixtureStore) Insert(record *EventsAllFixture) error {

	if err := record.BeforeSave(); err != nil {
		return err
	}

	if err := record.BeforeInsert(); err != nil {
		return err
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		if err := s.Insert(Schema.EventsAllFixture.BaseSchema, record); err != nil {
			return err
		}

		if err := record.AfterInsert(); err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *EventsAllFixtureStore) Update(record *EventsAllFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	if err := record.BeforeSave(); err != nil {
		return 0, err
	}

	if err := record.BeforeUpdate(); err != nil {
		return 0, err
	}

	err = s.Store.Transaction(func(s *kallax.Store) error {
		updated, err = s.Update(Schema.EventsAllFixture.BaseSchema, record, cols...)
		if err != nil {
			return err
		}

		if err := record.AfterUpdate(); err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return updated, nil

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *EventsAllFixtureStore) Save(record *EventsAllFixture) (updated bool, err error) {
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
func (s *EventsAllFixtureStore) Delete(record *EventsAllFixture) error {

	return s.Store.Delete(Schema.EventsAllFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *EventsAllFixtureStore) Find(q *EventsAllFixtureQuery) (*EventsAllFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewEventsAllFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *EventsAllFixtureStore) MustFind(q *EventsAllFixtureQuery) *EventsAllFixtureResultSet {
	return NewEventsAllFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *EventsAllFixtureStore) Count(q *EventsAllFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *EventsAllFixtureStore) MustCount(q *EventsAllFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *EventsAllFixtureStore) FindOne(q *EventsAllFixtureQuery) (*EventsAllFixture, error) {
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
func (s *EventsAllFixtureStore) MustFindOne(q *EventsAllFixtureQuery) *EventsAllFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the EventsAllFixture with the data in the database and
// makes it writable.
func (s *EventsAllFixtureStore) Reload(record *EventsAllFixture) error {
	return s.Store.Reload(Schema.EventsAllFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *EventsAllFixtureStore) Transaction(callback func(*EventsAllFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&EventsAllFixtureStore{store})
	})
}

// EventsAllFixtureQuery is the object used to create queries for the EventsAllFixture
// entity.
type EventsAllFixtureQuery struct {
	*kallax.BaseQuery
}

// NewEventsAllFixtureQuery returns a new instance of EventsAllFixtureQuery.
func NewEventsAllFixtureQuery() *EventsAllFixtureQuery {
	return &EventsAllFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.EventsAllFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *EventsAllFixtureQuery) Select(columns ...kallax.SchemaField) *EventsAllFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *EventsAllFixtureQuery) SelectNot(columns ...kallax.SchemaField) *EventsAllFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *EventsAllFixtureQuery) Copy() *EventsAllFixtureQuery {
	return &EventsAllFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *EventsAllFixtureQuery) Order(cols ...kallax.ColumnOrder) *EventsAllFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *EventsAllFixtureQuery) BatchSize(size uint64) *EventsAllFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *EventsAllFixtureQuery) Limit(n uint64) *EventsAllFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *EventsAllFixtureQuery) Offset(n uint64) *EventsAllFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *EventsAllFixtureQuery) Where(cond kallax.Condition) *EventsAllFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *EventsAllFixtureQuery) FindByID(v ...kallax.ULID) *EventsAllFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.EventsAllFixture.ID, values...))
}

// EventsAllFixtureResultSet is the set of results returned by a query to the
// database.
type EventsAllFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *EventsAllFixture
	lastErr   error
}

// NewEventsAllFixtureResultSet creates a new result set for rows of the type
// EventsAllFixture.
func NewEventsAllFixtureResultSet(rs kallax.ResultSet) *EventsAllFixtureResultSet {
	return &EventsAllFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *EventsAllFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.EventsAllFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*EventsAllFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *EventsAllFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *EventsAllFixtureResultSet) Get() (*EventsAllFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *EventsAllFixtureResultSet) ForEach(fn func(*EventsAllFixture) error) error {
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
func (rs *EventsAllFixtureResultSet) All() ([]*EventsAllFixture, error) {
	var result []*EventsAllFixture
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
func (rs *EventsAllFixtureResultSet) One() (*EventsAllFixture, error) {
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
func (rs *EventsAllFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *EventsAllFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewEventsFixture returns a new instance of EventsFixture.
func NewEventsFixture() (record *EventsFixture) {
	return newEventsFixture()
}

// GetID returns the primary key of the model.
func (r *EventsFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *EventsFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "checks":
		return types.JSON(&r.Checks), nil
	case "must_fail_before":
		return types.JSON(&r.MustFailBefore), nil
	case "must_fail_after":
		return types.JSON(&r.MustFailAfter), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in EventsFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *EventsFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "checks":
		return types.JSON(r.Checks), nil
	case "must_fail_before":
		return types.JSON(r.MustFailBefore), nil
	case "must_fail_after":
		return types.JSON(r.MustFailAfter), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in EventsFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *EventsFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model EventsFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *EventsFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model EventsFixture has no relationships")
}

// EventsFixtureStore is the entity to access the records of the type EventsFixture
// in the database.
type EventsFixtureStore struct {
	*kallax.Store
}

// NewEventsFixtureStore creates a new instance of EventsFixtureStore
// using a SQL database.
func NewEventsFixtureStore(db *sql.DB) *EventsFixtureStore {
	return &EventsFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *EventsFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *EventsFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a EventsFixture in the database. A non-persisted object is
// required for this operation.
func (s *EventsFixtureStore) Insert(record *EventsFixture) error {

	if err := record.BeforeInsert(); err != nil {
		return err
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		if err := s.Insert(Schema.EventsFixture.BaseSchema, record); err != nil {
			return err
		}

		if err := record.AfterInsert(); err != nil {
			return err
		}

		return nil
	})

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *EventsFixtureStore) Update(record *EventsFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	if err := record.BeforeUpdate(); err != nil {
		return 0, err
	}

	err = s.Store.Transaction(func(s *kallax.Store) error {
		updated, err = s.Update(Schema.EventsFixture.BaseSchema, record, cols...)
		if err != nil {
			return err
		}

		if err := record.AfterUpdate(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return updated, nil

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *EventsFixtureStore) Save(record *EventsFixture) (updated bool, err error) {
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
func (s *EventsFixtureStore) Delete(record *EventsFixture) error {

	return s.Store.Delete(Schema.EventsFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *EventsFixtureStore) Find(q *EventsFixtureQuery) (*EventsFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewEventsFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *EventsFixtureStore) MustFind(q *EventsFixtureQuery) *EventsFixtureResultSet {
	return NewEventsFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *EventsFixtureStore) Count(q *EventsFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *EventsFixtureStore) MustCount(q *EventsFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *EventsFixtureStore) FindOne(q *EventsFixtureQuery) (*EventsFixture, error) {
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
func (s *EventsFixtureStore) MustFindOne(q *EventsFixtureQuery) *EventsFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the EventsFixture with the data in the database and
// makes it writable.
func (s *EventsFixtureStore) Reload(record *EventsFixture) error {
	return s.Store.Reload(Schema.EventsFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *EventsFixtureStore) Transaction(callback func(*EventsFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&EventsFixtureStore{store})
	})
}

// EventsFixtureQuery is the object used to create queries for the EventsFixture
// entity.
type EventsFixtureQuery struct {
	*kallax.BaseQuery
}

// NewEventsFixtureQuery returns a new instance of EventsFixtureQuery.
func NewEventsFixtureQuery() *EventsFixtureQuery {
	return &EventsFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.EventsFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *EventsFixtureQuery) Select(columns ...kallax.SchemaField) *EventsFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *EventsFixtureQuery) SelectNot(columns ...kallax.SchemaField) *EventsFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *EventsFixtureQuery) Copy() *EventsFixtureQuery {
	return &EventsFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *EventsFixtureQuery) Order(cols ...kallax.ColumnOrder) *EventsFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *EventsFixtureQuery) BatchSize(size uint64) *EventsFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *EventsFixtureQuery) Limit(n uint64) *EventsFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *EventsFixtureQuery) Offset(n uint64) *EventsFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *EventsFixtureQuery) Where(cond kallax.Condition) *EventsFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *EventsFixtureQuery) FindByID(v ...kallax.ULID) *EventsFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.EventsFixture.ID, values...))
}

// EventsFixtureResultSet is the set of results returned by a query to the
// database.
type EventsFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *EventsFixture
	lastErr   error
}

// NewEventsFixtureResultSet creates a new result set for rows of the type
// EventsFixture.
func NewEventsFixtureResultSet(rs kallax.ResultSet) *EventsFixtureResultSet {
	return &EventsFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *EventsFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.EventsFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*EventsFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *EventsFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *EventsFixtureResultSet) Get() (*EventsFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *EventsFixtureResultSet) ForEach(fn func(*EventsFixture) error) error {
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
func (rs *EventsFixtureResultSet) All() ([]*EventsFixture, error) {
	var result []*EventsFixture
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
func (rs *EventsFixtureResultSet) One() (*EventsFixture, error) {
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
func (rs *EventsFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *EventsFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewEventsSaveFixture returns a new instance of EventsSaveFixture.
func NewEventsSaveFixture() (record *EventsSaveFixture) {
	return newEventsSaveFixture()
}

// GetID returns the primary key of the model.
func (r *EventsSaveFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *EventsSaveFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "checks":
		return types.JSON(&r.Checks), nil
	case "must_fail_before":
		return types.JSON(&r.MustFailBefore), nil
	case "must_fail_after":
		return types.JSON(&r.MustFailAfter), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in EventsSaveFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *EventsSaveFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "checks":
		return types.JSON(r.Checks), nil
	case "must_fail_before":
		return types.JSON(r.MustFailBefore), nil
	case "must_fail_after":
		return types.JSON(r.MustFailAfter), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in EventsSaveFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *EventsSaveFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model EventsSaveFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *EventsSaveFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model EventsSaveFixture has no relationships")
}

// EventsSaveFixtureStore is the entity to access the records of the type EventsSaveFixture
// in the database.
type EventsSaveFixtureStore struct {
	*kallax.Store
}

// NewEventsSaveFixtureStore creates a new instance of EventsSaveFixtureStore
// using a SQL database.
func NewEventsSaveFixtureStore(db *sql.DB) *EventsSaveFixtureStore {
	return &EventsSaveFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *EventsSaveFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *EventsSaveFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a EventsSaveFixture in the database. A non-persisted object is
// required for this operation.
func (s *EventsSaveFixtureStore) Insert(record *EventsSaveFixture) error {

	if err := record.BeforeSave(); err != nil {
		return err
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		if err := s.Insert(Schema.EventsSaveFixture.BaseSchema, record); err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *EventsSaveFixtureStore) Update(record *EventsSaveFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	if err := record.BeforeSave(); err != nil {
		return 0, err
	}

	err = s.Store.Transaction(func(s *kallax.Store) error {
		updated, err = s.Update(Schema.EventsSaveFixture.BaseSchema, record, cols...)
		if err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return updated, nil

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *EventsSaveFixtureStore) Save(record *EventsSaveFixture) (updated bool, err error) {
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
func (s *EventsSaveFixtureStore) Delete(record *EventsSaveFixture) error {

	return s.Store.Delete(Schema.EventsSaveFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *EventsSaveFixtureStore) Find(q *EventsSaveFixtureQuery) (*EventsSaveFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewEventsSaveFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *EventsSaveFixtureStore) MustFind(q *EventsSaveFixtureQuery) *EventsSaveFixtureResultSet {
	return NewEventsSaveFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *EventsSaveFixtureStore) Count(q *EventsSaveFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *EventsSaveFixtureStore) MustCount(q *EventsSaveFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *EventsSaveFixtureStore) FindOne(q *EventsSaveFixtureQuery) (*EventsSaveFixture, error) {
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
func (s *EventsSaveFixtureStore) MustFindOne(q *EventsSaveFixtureQuery) *EventsSaveFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the EventsSaveFixture with the data in the database and
// makes it writable.
func (s *EventsSaveFixtureStore) Reload(record *EventsSaveFixture) error {
	return s.Store.Reload(Schema.EventsSaveFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *EventsSaveFixtureStore) Transaction(callback func(*EventsSaveFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&EventsSaveFixtureStore{store})
	})
}

// EventsSaveFixtureQuery is the object used to create queries for the EventsSaveFixture
// entity.
type EventsSaveFixtureQuery struct {
	*kallax.BaseQuery
}

// NewEventsSaveFixtureQuery returns a new instance of EventsSaveFixtureQuery.
func NewEventsSaveFixtureQuery() *EventsSaveFixtureQuery {
	return &EventsSaveFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.EventsSaveFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *EventsSaveFixtureQuery) Select(columns ...kallax.SchemaField) *EventsSaveFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *EventsSaveFixtureQuery) SelectNot(columns ...kallax.SchemaField) *EventsSaveFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *EventsSaveFixtureQuery) Copy() *EventsSaveFixtureQuery {
	return &EventsSaveFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *EventsSaveFixtureQuery) Order(cols ...kallax.ColumnOrder) *EventsSaveFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *EventsSaveFixtureQuery) BatchSize(size uint64) *EventsSaveFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *EventsSaveFixtureQuery) Limit(n uint64) *EventsSaveFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *EventsSaveFixtureQuery) Offset(n uint64) *EventsSaveFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *EventsSaveFixtureQuery) Where(cond kallax.Condition) *EventsSaveFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *EventsSaveFixtureQuery) FindByID(v ...kallax.ULID) *EventsSaveFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.EventsSaveFixture.ID, values...))
}

// EventsSaveFixtureResultSet is the set of results returned by a query to the
// database.
type EventsSaveFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *EventsSaveFixture
	lastErr   error
}

// NewEventsSaveFixtureResultSet creates a new result set for rows of the type
// EventsSaveFixture.
func NewEventsSaveFixtureResultSet(rs kallax.ResultSet) *EventsSaveFixtureResultSet {
	return &EventsSaveFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *EventsSaveFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.EventsSaveFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*EventsSaveFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *EventsSaveFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *EventsSaveFixtureResultSet) Get() (*EventsSaveFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *EventsSaveFixtureResultSet) ForEach(fn func(*EventsSaveFixture) error) error {
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
func (rs *EventsSaveFixtureResultSet) All() ([]*EventsSaveFixture, error) {
	var result []*EventsSaveFixture
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
func (rs *EventsSaveFixtureResultSet) One() (*EventsSaveFixture, error) {
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
func (rs *EventsSaveFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *EventsSaveFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewJSONModel returns a new instance of JSONModel.
func NewJSONModel() (record *JSONModel) {
	return newJSONModel()
}

// GetID returns the primary key of the model.
func (r *JSONModel) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *JSONModel) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "foo":
		return &r.Foo, nil
	case "bar":
		if r.Bar == nil {
			r.Bar = new(Bar)
		}
		return types.JSON(r.Bar), nil
	case "baz_slice":
		return types.JSON(&r.BazSlice), nil
	case "baz":
		return types.JSON(&r.Baz), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in JSONModel: %s", col)
	}
}

// Value returns the value of the given column.
func (r *JSONModel) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "foo":
		return r.Foo, nil
	case "bar":
		if r.Bar == (*Bar)(nil) {
			return nil, nil
		}
		return types.JSON(r.Bar), nil
	case "baz_slice":
		return types.JSON(r.BazSlice), nil
	case "baz":
		return types.JSON(r.Baz), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in JSONModel: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *JSONModel) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model JSONModel has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *JSONModel) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model JSONModel has no relationships")
}

// JSONModelStore is the entity to access the records of the type JSONModel
// in the database.
type JSONModelStore struct {
	*kallax.Store
}

// NewJSONModelStore creates a new instance of JSONModelStore
// using a SQL database.
func NewJSONModelStore(db *sql.DB) *JSONModelStore {
	return &JSONModelStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *JSONModelStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *JSONModelStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a JSONModel in the database. A non-persisted object is
// required for this operation.
func (s *JSONModelStore) Insert(record *JSONModel) error {

	return s.Store.Insert(Schema.JSONModel.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *JSONModelStore) Update(record *JSONModel, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.JSONModel.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *JSONModelStore) Save(record *JSONModel) (updated bool, err error) {
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
func (s *JSONModelStore) Delete(record *JSONModel) error {

	return s.Store.Delete(Schema.JSONModel.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *JSONModelStore) Find(q *JSONModelQuery) (*JSONModelResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewJSONModelResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *JSONModelStore) MustFind(q *JSONModelQuery) *JSONModelResultSet {
	return NewJSONModelResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *JSONModelStore) Count(q *JSONModelQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *JSONModelStore) MustCount(q *JSONModelQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *JSONModelStore) FindOne(q *JSONModelQuery) (*JSONModel, error) {
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
func (s *JSONModelStore) MustFindOne(q *JSONModelQuery) *JSONModel {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the JSONModel with the data in the database and
// makes it writable.
func (s *JSONModelStore) Reload(record *JSONModel) error {
	return s.Store.Reload(Schema.JSONModel.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *JSONModelStore) Transaction(callback func(*JSONModelStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&JSONModelStore{store})
	})
}

// JSONModelQuery is the object used to create queries for the JSONModel
// entity.
type JSONModelQuery struct {
	*kallax.BaseQuery
}

// NewJSONModelQuery returns a new instance of JSONModelQuery.
func NewJSONModelQuery() *JSONModelQuery {
	return &JSONModelQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.JSONModel.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *JSONModelQuery) Select(columns ...kallax.SchemaField) *JSONModelQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *JSONModelQuery) SelectNot(columns ...kallax.SchemaField) *JSONModelQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *JSONModelQuery) Copy() *JSONModelQuery {
	return &JSONModelQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *JSONModelQuery) Order(cols ...kallax.ColumnOrder) *JSONModelQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *JSONModelQuery) BatchSize(size uint64) *JSONModelQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *JSONModelQuery) Limit(n uint64) *JSONModelQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *JSONModelQuery) Offset(n uint64) *JSONModelQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *JSONModelQuery) Where(cond kallax.Condition) *JSONModelQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *JSONModelQuery) FindByID(v ...kallax.ULID) *JSONModelQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.JSONModel.ID, values...))
}

// FindByFoo adds a new filter to the query that will require that
// the Foo property is equal to the passed value.
func (q *JSONModelQuery) FindByFoo(v string) *JSONModelQuery {
	return q.Where(kallax.Eq(Schema.JSONModel.Foo, v))
}

// JSONModelResultSet is the set of results returned by a query to the
// database.
type JSONModelResultSet struct {
	ResultSet kallax.ResultSet
	last      *JSONModel
	lastErr   error
}

// NewJSONModelResultSet creates a new result set for rows of the type
// JSONModel.
func NewJSONModelResultSet(rs kallax.ResultSet) *JSONModelResultSet {
	return &JSONModelResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *JSONModelResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.JSONModel.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*JSONModel)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *JSONModel")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *JSONModelResultSet) Get() (*JSONModel, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *JSONModelResultSet) ForEach(fn func(*JSONModel) error) error {
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
func (rs *JSONModelResultSet) All() ([]*JSONModel, error) {
	var result []*JSONModel
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
func (rs *JSONModelResultSet) One() (*JSONModel, error) {
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
func (rs *JSONModelResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *JSONModelResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewMultiKeySortFixture returns a new instance of MultiKeySortFixture.
func NewMultiKeySortFixture() (record *MultiKeySortFixture) {
	return newMultiKeySortFixture()
}

// GetID returns the primary key of the model.
func (r *MultiKeySortFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *MultiKeySortFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "name":
		return &r.Name, nil
	case "start":
		return &r.Start, nil
	case "_end":
		return &r.End, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in MultiKeySortFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *MultiKeySortFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "name":
		return r.Name, nil
	case "start":
		return r.Start, nil
	case "_end":
		return r.End, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in MultiKeySortFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *MultiKeySortFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model MultiKeySortFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *MultiKeySortFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model MultiKeySortFixture has no relationships")
}

// MultiKeySortFixtureStore is the entity to access the records of the type MultiKeySortFixture
// in the database.
type MultiKeySortFixtureStore struct {
	*kallax.Store
}

// NewMultiKeySortFixtureStore creates a new instance of MultiKeySortFixtureStore
// using a SQL database.
func NewMultiKeySortFixtureStore(db *sql.DB) *MultiKeySortFixtureStore {
	return &MultiKeySortFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *MultiKeySortFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *MultiKeySortFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a MultiKeySortFixture in the database. A non-persisted object is
// required for this operation.
func (s *MultiKeySortFixtureStore) Insert(record *MultiKeySortFixture) error {

	return s.Store.Insert(Schema.MultiKeySortFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *MultiKeySortFixtureStore) Update(record *MultiKeySortFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.MultiKeySortFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *MultiKeySortFixtureStore) Save(record *MultiKeySortFixture) (updated bool, err error) {
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
func (s *MultiKeySortFixtureStore) Delete(record *MultiKeySortFixture) error {

	return s.Store.Delete(Schema.MultiKeySortFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *MultiKeySortFixtureStore) Find(q *MultiKeySortFixtureQuery) (*MultiKeySortFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewMultiKeySortFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *MultiKeySortFixtureStore) MustFind(q *MultiKeySortFixtureQuery) *MultiKeySortFixtureResultSet {
	return NewMultiKeySortFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *MultiKeySortFixtureStore) Count(q *MultiKeySortFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *MultiKeySortFixtureStore) MustCount(q *MultiKeySortFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *MultiKeySortFixtureStore) FindOne(q *MultiKeySortFixtureQuery) (*MultiKeySortFixture, error) {
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
func (s *MultiKeySortFixtureStore) MustFindOne(q *MultiKeySortFixtureQuery) *MultiKeySortFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the MultiKeySortFixture with the data in the database and
// makes it writable.
func (s *MultiKeySortFixtureStore) Reload(record *MultiKeySortFixture) error {
	return s.Store.Reload(Schema.MultiKeySortFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *MultiKeySortFixtureStore) Transaction(callback func(*MultiKeySortFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&MultiKeySortFixtureStore{store})
	})
}

// MultiKeySortFixtureQuery is the object used to create queries for the MultiKeySortFixture
// entity.
type MultiKeySortFixtureQuery struct {
	*kallax.BaseQuery
}

// NewMultiKeySortFixtureQuery returns a new instance of MultiKeySortFixtureQuery.
func NewMultiKeySortFixtureQuery() *MultiKeySortFixtureQuery {
	return &MultiKeySortFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.MultiKeySortFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *MultiKeySortFixtureQuery) Select(columns ...kallax.SchemaField) *MultiKeySortFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *MultiKeySortFixtureQuery) SelectNot(columns ...kallax.SchemaField) *MultiKeySortFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *MultiKeySortFixtureQuery) Copy() *MultiKeySortFixtureQuery {
	return &MultiKeySortFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *MultiKeySortFixtureQuery) Order(cols ...kallax.ColumnOrder) *MultiKeySortFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *MultiKeySortFixtureQuery) BatchSize(size uint64) *MultiKeySortFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *MultiKeySortFixtureQuery) Limit(n uint64) *MultiKeySortFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *MultiKeySortFixtureQuery) Offset(n uint64) *MultiKeySortFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *MultiKeySortFixtureQuery) Where(cond kallax.Condition) *MultiKeySortFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *MultiKeySortFixtureQuery) FindByID(v ...kallax.ULID) *MultiKeySortFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.MultiKeySortFixture.ID, values...))
}

// FindByName adds a new filter to the query that will require that
// the Name property is equal to the passed value.
func (q *MultiKeySortFixtureQuery) FindByName(v string) *MultiKeySortFixtureQuery {
	return q.Where(kallax.Eq(Schema.MultiKeySortFixture.Name, v))
}

// FindByStart adds a new filter to the query that will require that
// the Start property is equal to the passed value.
func (q *MultiKeySortFixtureQuery) FindByStart(cond kallax.ScalarCond, v time.Time) *MultiKeySortFixtureQuery {
	return q.Where(cond(Schema.MultiKeySortFixture.Start, v))
}

// FindByEnd adds a new filter to the query that will require that
// the End property is equal to the passed value.
func (q *MultiKeySortFixtureQuery) FindByEnd(cond kallax.ScalarCond, v time.Time) *MultiKeySortFixtureQuery {
	return q.Where(cond(Schema.MultiKeySortFixture.End, v))
}

// MultiKeySortFixtureResultSet is the set of results returned by a query to the
// database.
type MultiKeySortFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *MultiKeySortFixture
	lastErr   error
}

// NewMultiKeySortFixtureResultSet creates a new result set for rows of the type
// MultiKeySortFixture.
func NewMultiKeySortFixtureResultSet(rs kallax.ResultSet) *MultiKeySortFixtureResultSet {
	return &MultiKeySortFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *MultiKeySortFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.MultiKeySortFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*MultiKeySortFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *MultiKeySortFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *MultiKeySortFixtureResultSet) Get() (*MultiKeySortFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *MultiKeySortFixtureResultSet) ForEach(fn func(*MultiKeySortFixture) error) error {
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
func (rs *MultiKeySortFixtureResultSet) All() ([]*MultiKeySortFixture, error) {
	var result []*MultiKeySortFixture
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
func (rs *MultiKeySortFixtureResultSet) One() (*MultiKeySortFixture, error) {
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
func (rs *MultiKeySortFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *MultiKeySortFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewNullable returns a new instance of Nullable.
func NewNullable() (record *Nullable) {
	return new(Nullable)
}

// GetID returns the primary key of the model.
func (r *Nullable) GetID() kallax.Identifier {
	return (*kallax.NumericID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *Nullable) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.NumericID)(&r.ID), nil
	case "t":
		return types.Nullable(&r.T), nil
	case "some_json":
		if r.SomeJSON == nil {
			r.SomeJSON = new(SomeJSON)
		}
		return types.JSON(r.SomeJSON), nil
	case "scanner":
		if r.Scanner == nil {
			r.Scanner = new(kallax.ULID)
		}
		return types.Nullable(r.Scanner), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Nullable: %s", col)
	}
}

// Value returns the value of the given column.
func (r *Nullable) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "t":
		if r.T == (*time.Time)(nil) {
			return nil, nil
		}
		return r.T, nil
	case "some_json":
		if r.SomeJSON == (*SomeJSON)(nil) {
			return nil, nil
		}
		return types.JSON(r.SomeJSON), nil
	case "scanner":
		if r.Scanner == (*kallax.ULID)(nil) {
			return nil, nil
		}
		return r.Scanner, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Nullable: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *Nullable) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model Nullable has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *Nullable) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model Nullable has no relationships")
}

// NullableStore is the entity to access the records of the type Nullable
// in the database.
type NullableStore struct {
	*kallax.Store
}

// NewNullableStore creates a new instance of NullableStore
// using a SQL database.
func NewNullableStore(db *sql.DB) *NullableStore {
	return &NullableStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *NullableStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *NullableStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a Nullable in the database. A non-persisted object is
// required for this operation.
func (s *NullableStore) Insert(record *Nullable) error {

	return s.Store.Insert(Schema.Nullable.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *NullableStore) Update(record *Nullable, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.Nullable.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *NullableStore) Save(record *Nullable) (updated bool, err error) {
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
func (s *NullableStore) Delete(record *Nullable) error {

	return s.Store.Delete(Schema.Nullable.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *NullableStore) Find(q *NullableQuery) (*NullableResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewNullableResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *NullableStore) MustFind(q *NullableQuery) *NullableResultSet {
	return NewNullableResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *NullableStore) Count(q *NullableQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *NullableStore) MustCount(q *NullableQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *NullableStore) FindOne(q *NullableQuery) (*Nullable, error) {
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
func (s *NullableStore) MustFindOne(q *NullableQuery) *Nullable {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the Nullable with the data in the database and
// makes it writable.
func (s *NullableStore) Reload(record *Nullable) error {
	return s.Store.Reload(Schema.Nullable.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *NullableStore) Transaction(callback func(*NullableStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&NullableStore{store})
	})
}

// NullableQuery is the object used to create queries for the Nullable
// entity.
type NullableQuery struct {
	*kallax.BaseQuery
}

// NewNullableQuery returns a new instance of NullableQuery.
func NewNullableQuery() *NullableQuery {
	return &NullableQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.Nullable.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *NullableQuery) Select(columns ...kallax.SchemaField) *NullableQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *NullableQuery) SelectNot(columns ...kallax.SchemaField) *NullableQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *NullableQuery) Copy() *NullableQuery {
	return &NullableQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *NullableQuery) Order(cols ...kallax.ColumnOrder) *NullableQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *NullableQuery) BatchSize(size uint64) *NullableQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *NullableQuery) Limit(n uint64) *NullableQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *NullableQuery) Offset(n uint64) *NullableQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *NullableQuery) Where(cond kallax.Condition) *NullableQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *NullableQuery) FindByID(v ...int64) *NullableQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.Nullable.ID, values...))
}

// FindByT adds a new filter to the query that will require that
// the T property is equal to the passed value.
func (q *NullableQuery) FindByT(cond kallax.ScalarCond, v time.Time) *NullableQuery {
	return q.Where(cond(Schema.Nullable.T, v))
}

// FindByScanner adds a new filter to the query that will require that
// the Scanner property is equal to the passed value.
func (q *NullableQuery) FindByScanner(v kallax.ULID) *NullableQuery {
	return q.Where(kallax.Eq(Schema.Nullable.Scanner, v))
}

// NullableResultSet is the set of results returned by a query to the
// database.
type NullableResultSet struct {
	ResultSet kallax.ResultSet
	last      *Nullable
	lastErr   error
}

// NewNullableResultSet creates a new result set for rows of the type
// Nullable.
func NewNullableResultSet(rs kallax.ResultSet) *NullableResultSet {
	return &NullableResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *NullableResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.Nullable.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*Nullable)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *Nullable")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *NullableResultSet) Get() (*Nullable, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *NullableResultSet) ForEach(fn func(*Nullable) error) error {
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
func (rs *NullableResultSet) All() ([]*Nullable, error) {
	var result []*Nullable
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
func (rs *NullableResultSet) One() (*Nullable, error) {
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
func (rs *NullableResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *NullableResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewPerson returns a new instance of Person.
func NewPerson(name string) (record *Person) {
	return newPerson(name)
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
	case "Car":
		return new(Car), nil

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
	case "Car":
		val, ok := rel.(*Car)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Car", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Car = val
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

// GenericStore returns the generic store of this store.
func (s *PersonStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *PersonStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

func (s *PersonStore) relationshipRecords(record *Person) []kallax.RecordWithSchema {
	var records []kallax.RecordWithSchema

	for _, rec := range record.Pets {
		rec.ClearVirtualColumns()
		rec.AddVirtualColumn("owner_id", record.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.Pet.BaseSchema,
			Record: rec,
		})
	}

	if record.Car != nil {
		record.Car.ClearVirtualColumns()
		record.Car.AddVirtualColumn("owner_id", record.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.Car.BaseSchema,
			Record: record.Car,
		})
	}

	return records
}

// Insert inserts a Person in the database. A non-persisted object is
// required for this operation.
func (s *PersonStore) Insert(record *Person) error {

	if err := record.BeforeSave(); err != nil {
		return err
	}

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

			if err := record.AfterSave(); err != nil {
				return err
			}

			return nil
		})
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		if err := s.Insert(Schema.Person.BaseSchema, record); err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *PersonStore) Update(record *Person, cols ...kallax.SchemaField) (updated int64, err error) {

	if err := record.BeforeSave(); err != nil {
		return 0, err
	}

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

			if err := record.AfterSave(); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return 0, err
		}

		return updated, nil
	}

	err = s.Store.Transaction(func(s *kallax.Store) error {
		updated, err = s.Update(Schema.Person.BaseSchema, record, cols...)
		if err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return updated, nil

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

	if err := record.BeforeDelete(); err != nil {
		return err
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		err := s.Delete(Schema.Person.BaseSchema, record)
		if err != nil {
			return err
		}

		return record.AfterDelete()
	})

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

// RemoveCar removes from the database the given relationship of the
// model. It also resets the field Car of the model.
func (s *PersonStore) RemoveCar(record *Person) error {
	var r kallax.Record = record.Car
	if beforeDeleter, ok := r.(kallax.BeforeDeleter); ok {
		if err := beforeDeleter.BeforeDelete(); err != nil {
			return err
		}
	}

	var err error
	if afterDeleter, ok := r.(kallax.AfterDeleter); ok {
		err = s.Store.Transaction(func(s *kallax.Store) error {
			err := s.Delete(Schema.Car.BaseSchema, r)
			if err != nil {
				return err
			}

			return afterDeleter.AfterDelete()
		})
	} else {
		err = s.Store.Delete(Schema.Car.BaseSchema, r)
	}
	if err != nil {
		return err
	}

	record.Car = nil
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

func (q *PersonQuery) WithCar() *PersonQuery {
	q.AddRelation(Schema.Car.BaseSchema, "Car", kallax.OneToOne, nil)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
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
// the Name property is equal to the passed value.
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
func NewPet(name string, kind string, owner *Person) (record *Pet) {
	return newPet(name, kind, owner)
}

// GetID returns the primary key of the model.
func (r *Pet) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *Pet) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "name":
		return &r.Name, nil
	case "kind":
		return &r.Kind, nil
	case "owner_id":
		return types.Nullable(kallax.VirtualColumn("owner_id", r, new(kallax.NumericID))), nil

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
		return r.Kind, nil
	case "owner_id":
		return r.Model.VirtualColumn(col), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in Pet: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *Pet) NewRelationshipRecord(field string) (kallax.Record, error) {
	switch field {
	case "Owner":
		return new(Person), nil

	}
	return nil, fmt.Errorf("kallax: model Pet has no relationship %s", field)
}

// SetRelationship sets the given relationship in the given field.
func (r *Pet) SetRelationship(field string, rel interface{}) error {
	switch field {
	case "Owner":
		val, ok := rel.(*Person)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Owner", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Owner = val
		}

		return nil

	}
	return fmt.Errorf("kallax: model Pet has no relationship %s", field)
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

// GenericStore returns the generic store of this store.
func (s *PetStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *PetStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

func (s *PetStore) inverseRecords(record *Pet) []kallax.RecordWithSchema {
	record.ClearVirtualColumns()
	var records []kallax.RecordWithSchema

	if record.Owner != nil {
		record.AddVirtualColumn("owner_id", record.Owner.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.Person.BaseSchema,
			Record: record.Owner,
		})
	}

	return records
}

// Insert inserts a Pet in the database. A non-persisted object is
// required for this operation.
func (s *PetStore) Insert(record *Pet) error {

	if err := record.BeforeSave(); err != nil {
		return err
	}

	inverseRecords := s.inverseRecords(record)

	if len(inverseRecords) > 0 {
		return s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			if err := s.Insert(Schema.Pet.BaseSchema, record); err != nil {
				return err
			}

			if err := record.AfterSave(); err != nil {
				return err
			}

			return nil
		})
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		if err := s.Insert(Schema.Pet.BaseSchema, record); err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *PetStore) Update(record *Pet, cols ...kallax.SchemaField) (updated int64, err error) {

	if err := record.BeforeSave(); err != nil {
		return 0, err
	}

	inverseRecords := s.inverseRecords(record)

	if len(inverseRecords) > 0 {
		err = s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			updated, err = s.Update(Schema.Pet.BaseSchema, record, cols...)
			if err != nil {
				return err
			}

			if err := record.AfterSave(); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return 0, err
		}

		return updated, nil
	}

	err = s.Store.Transaction(func(s *kallax.Store) error {
		updated, err = s.Update(Schema.Pet.BaseSchema, record, cols...)
		if err != nil {
			return err
		}

		if err := record.AfterSave(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return updated, nil

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

	if err := record.BeforeDelete(); err != nil {
		return err
	}

	return s.Store.Transaction(func(s *kallax.Store) error {
		err := s.Delete(Schema.Pet.BaseSchema, record)
		if err != nil {
			return err
		}

		return record.AfterDelete()
	})

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

func (q *PetQuery) WithOwner() *PetQuery {
	q.AddRelation(Schema.Person.BaseSchema, "Owner", kallax.OneToOne, nil)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *PetQuery) FindByID(v ...kallax.ULID) *PetQuery {
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
// the Name property is equal to the passed value.
func (q *PetQuery) FindByName(v string) *PetQuery {
	return q.Where(kallax.Eq(Schema.Pet.Name, v))
}

// FindByKind adds a new filter to the query that will require that
// the Kind property is equal to the passed value.
func (q *PetQuery) FindByKind(v string) *PetQuery {
	return q.Where(kallax.Eq(Schema.Pet.Kind, v))
}

// FindByOwner adds a new filter to the query that will require that
// the foreign key of Owner is equal to the passed value.
func (q *PetQuery) FindByOwner(v int64) *PetQuery {
	return q.Where(kallax.Eq(Schema.Pet.OwnerFK, v))
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

// NewQueryFixture returns a new instance of QueryFixture.
func NewQueryFixture(f string) (record *QueryFixture) {
	return newQueryFixture(f)
}

// GetID returns the primary key of the model.
func (r *QueryFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *QueryFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "inverse_id":
		return types.Nullable(kallax.VirtualColumn("inverse_id", r, new(kallax.ULID))), nil
	case "embedded":
		return types.JSON(&r.Embedded), nil
	case "inline":
		return &r.Inline.Inline, nil
	case "map_of_string":
		return types.JSON(&r.MapOfString), nil
	case "map_of_interface":
		return types.JSON(&r.MapOfInterface), nil
	case "map_of_some_type":
		return types.JSON(&r.MapOfSomeType), nil
	case "foo":
		return &r.Foo, nil
	case "string_property":
		return &r.StringProperty, nil
	case "integer":
		return &r.Integer, nil
	case "integer64":
		return &r.Integer64, nil
	case "float32":
		return &r.Float32, nil
	case "boolean":
		return &r.Boolean, nil
	case "array_param":
		return types.Array(&r.ArrayParam, 3), nil
	case "slice_param":
		return types.Slice(&r.SliceParam), nil
	case "alias_array_param":
		return types.Array(&r.AliasArrayParam, 3), nil
	case "alias_slice_param":
		return types.Slice((*[]string)(&r.AliasSliceParam)), nil
	case "alias_string_param":
		return (*string)(&r.AliasStringParam), nil
	case "alias_int_param":
		return (*int)(&r.AliasIntParam), nil
	case "dummy_param":
		return types.JSON(&r.DummyParam), nil
	case "alias_dummy_param":
		return types.JSON(&r.AliasDummyParam), nil
	case "slice_dummy_param":
		return types.JSON(&r.SliceDummyParam), nil
	case "idproperty_param":
		return &r.IDPropertyParam, nil
	case "interface_prop_param":
		return &r.InterfacePropParam, nil
	case "urlparam":
		return (*types.URL)(&r.URLParam), nil
	case "time_param":
		return &r.TimeParam, nil
	case "alias_arr_alias_string_param":
		return types.Slice(&r.AliasArrAliasStringParam), nil
	case "alias_here_array_param":
		return types.Array(&r.AliasHereArrayParam, 3), nil
	case "array_alias_here_string_param":
		return types.Slice(&r.ArrayAliasHereStringParam), nil
	case "scanner_valuer_param":
		return &r.ScannerValuerParam, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in QueryFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *QueryFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "inverse_id":
		return r.Model.VirtualColumn(col), nil
	case "embedded":
		return types.JSON(r.Embedded), nil
	case "inline":
		return r.Inline.Inline, nil
	case "map_of_string":
		return types.JSON(r.MapOfString), nil
	case "map_of_interface":
		return types.JSON(r.MapOfInterface), nil
	case "map_of_some_type":
		return types.JSON(r.MapOfSomeType), nil
	case "foo":
		return r.Foo, nil
	case "string_property":
		return r.StringProperty, nil
	case "integer":
		return r.Integer, nil
	case "integer64":
		return r.Integer64, nil
	case "float32":
		return r.Float32, nil
	case "boolean":
		return r.Boolean, nil
	case "array_param":
		return types.Array(&r.ArrayParam, 3), nil
	case "slice_param":
		return types.Slice(r.SliceParam), nil
	case "alias_array_param":
		return types.Array(&r.AliasArrayParam, 3), nil
	case "alias_slice_param":
		return types.Slice(r.AliasSliceParam), nil
	case "alias_string_param":
		return (string)(r.AliasStringParam), nil
	case "alias_int_param":
		return (int)(r.AliasIntParam), nil
	case "dummy_param":
		return types.JSON(r.DummyParam), nil
	case "alias_dummy_param":
		return types.JSON(r.AliasDummyParam), nil
	case "slice_dummy_param":
		return types.JSON(r.SliceDummyParam), nil
	case "idproperty_param":
		return r.IDPropertyParam, nil
	case "interface_prop_param":
		return r.InterfacePropParam, nil
	case "urlparam":
		return (*types.URL)(&r.URLParam), nil
	case "time_param":
		return r.TimeParam, nil
	case "alias_arr_alias_string_param":
		return types.Slice(r.AliasArrAliasStringParam), nil
	case "alias_here_array_param":
		return types.Array(&r.AliasHereArrayParam, 3), nil
	case "array_alias_here_string_param":
		return types.Slice(r.ArrayAliasHereStringParam), nil
	case "scanner_valuer_param":
		return r.ScannerValuerParam, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in QueryFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *QueryFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	switch field {
	case "Relation":
		return new(QueryRelationFixture), nil
	case "Inverse":
		return new(QueryRelationFixture), nil
	case "NRelation":
		return new(QueryRelationFixture), nil

	}
	return nil, fmt.Errorf("kallax: model QueryFixture has no relationship %s", field)
}

// SetRelationship sets the given relationship in the given field.
func (r *QueryFixture) SetRelationship(field string, rel interface{}) error {
	switch field {
	case "Relation":
		val, ok := rel.(*QueryRelationFixture)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Relation", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Relation = val
		}

		return nil
	case "Inverse":
		val, ok := rel.(*QueryRelationFixture)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Inverse", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Inverse = val
		}

		return nil
	case "NRelation":
		records, ok := rel.([]kallax.Record)
		if !ok {
			return fmt.Errorf("kallax: relationship field %s needs a collection of records, not %T", field, rel)
		}

		r.NRelation = make([]*QueryRelationFixture, len(records))
		for i, record := range records {
			rel, ok := record.(*QueryRelationFixture)
			if !ok {
				return fmt.Errorf("kallax: element of type %T cannot be added to relationship %s", record, field)
			}
			r.NRelation[i] = rel
		}
		return nil

	}
	return fmt.Errorf("kallax: model QueryFixture has no relationship %s", field)
}

// QueryFixtureStore is the entity to access the records of the type QueryFixture
// in the database.
type QueryFixtureStore struct {
	*kallax.Store
}

// NewQueryFixtureStore creates a new instance of QueryFixtureStore
// using a SQL database.
func NewQueryFixtureStore(db *sql.DB) *QueryFixtureStore {
	return &QueryFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *QueryFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *QueryFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

func (s *QueryFixtureStore) relationshipRecords(record *QueryFixture) []kallax.RecordWithSchema {
	var records []kallax.RecordWithSchema

	if record.Relation != nil {
		record.Relation.ClearVirtualColumns()
		record.Relation.AddVirtualColumn("owner_id", record.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.QueryRelationFixture.BaseSchema,
			Record: record.Relation,
		})
	}

	for _, rec := range record.NRelation {
		rec.ClearVirtualColumns()
		rec.AddVirtualColumn("owner_id", record.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.QueryRelationFixture.BaseSchema,
			Record: rec,
		})
	}

	return records
}

func (s *QueryFixtureStore) inverseRecords(record *QueryFixture) []kallax.RecordWithSchema {
	record.ClearVirtualColumns()
	var records []kallax.RecordWithSchema

	if record.Inverse != nil {
		record.AddVirtualColumn("inverse_id", record.Inverse.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.QueryRelationFixture.BaseSchema,
			Record: record.Inverse,
		})
	}

	return records
}

// Insert inserts a QueryFixture in the database. A non-persisted object is
// required for this operation.
func (s *QueryFixtureStore) Insert(record *QueryFixture) error {

	records := s.relationshipRecords(record)

	inverseRecords := s.inverseRecords(record)

	if len(records) > 0 && len(inverseRecords) > 0 {
		return s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			if err := s.Insert(Schema.QueryFixture.BaseSchema, record); err != nil {
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

	return s.Store.Insert(Schema.QueryFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *QueryFixtureStore) Update(record *QueryFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	records := s.relationshipRecords(record)

	inverseRecords := s.inverseRecords(record)

	if len(records) > 0 && len(inverseRecords) > 0 {
		err = s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			updated, err = s.Update(Schema.QueryFixture.BaseSchema, record, cols...)
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

	return s.Store.Update(Schema.QueryFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *QueryFixtureStore) Save(record *QueryFixture) (updated bool, err error) {
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
func (s *QueryFixtureStore) Delete(record *QueryFixture) error {

	return s.Store.Delete(Schema.QueryFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *QueryFixtureStore) Find(q *QueryFixtureQuery) (*QueryFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewQueryFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *QueryFixtureStore) MustFind(q *QueryFixtureQuery) *QueryFixtureResultSet {
	return NewQueryFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *QueryFixtureStore) Count(q *QueryFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *QueryFixtureStore) MustCount(q *QueryFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *QueryFixtureStore) FindOne(q *QueryFixtureQuery) (*QueryFixture, error) {
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
func (s *QueryFixtureStore) MustFindOne(q *QueryFixtureQuery) *QueryFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the QueryFixture with the data in the database and
// makes it writable.
func (s *QueryFixtureStore) Reload(record *QueryFixture) error {
	return s.Store.Reload(Schema.QueryFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *QueryFixtureStore) Transaction(callback func(*QueryFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&QueryFixtureStore{store})
	})
}

// RemoveRelation removes from the database the given relationship of the
// model. It also resets the field Relation of the model.
func (s *QueryFixtureStore) RemoveRelation(record *QueryFixture) error {
	var r kallax.Record = record.Relation
	if beforeDeleter, ok := r.(kallax.BeforeDeleter); ok {
		if err := beforeDeleter.BeforeDelete(); err != nil {
			return err
		}
	}

	var err error
	if afterDeleter, ok := r.(kallax.AfterDeleter); ok {
		err = s.Store.Transaction(func(s *kallax.Store) error {
			err := s.Delete(Schema.QueryRelationFixture.BaseSchema, r)
			if err != nil {
				return err
			}

			return afterDeleter.AfterDelete()
		})
	} else {
		err = s.Store.Delete(Schema.QueryRelationFixture.BaseSchema, r)
	}
	if err != nil {
		return err
	}

	record.Relation = nil
	return nil
}

// RemoveNRelation removes the given items of the NRelation field of the
// model. If no items are given, it removes all of them.
// The items will also be removed from the passed record inside this method.
func (s *QueryFixtureStore) RemoveNRelation(record *QueryFixture, deleted ...*QueryRelationFixture) error {
	var updated []*QueryRelationFixture
	var clear bool
	if len(deleted) == 0 {
		clear = true
		deleted = record.NRelation
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

				if err := s.Delete(Schema.QueryRelationFixture.BaseSchema, d); err != nil {
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
			record.NRelation = nil
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
				err := s.Delete(Schema.QueryRelationFixture.BaseSchema, r)
				if err != nil {
					return err
				}

				return afterDeleter.AfterDelete()
			})
		} else {
			err = s.Store.Delete(Schema.QueryRelationFixture.BaseSchema, deleted[0])
		}

		if err != nil {
			return err
		}
	}

	for _, r := range record.NRelation {
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
	record.NRelation = updated
	return nil
}

// QueryFixtureQuery is the object used to create queries for the QueryFixture
// entity.
type QueryFixtureQuery struct {
	*kallax.BaseQuery
}

// NewQueryFixtureQuery returns a new instance of QueryFixtureQuery.
func NewQueryFixtureQuery() *QueryFixtureQuery {
	return &QueryFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.QueryFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *QueryFixtureQuery) Select(columns ...kallax.SchemaField) *QueryFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *QueryFixtureQuery) SelectNot(columns ...kallax.SchemaField) *QueryFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *QueryFixtureQuery) Copy() *QueryFixtureQuery {
	return &QueryFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *QueryFixtureQuery) Order(cols ...kallax.ColumnOrder) *QueryFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *QueryFixtureQuery) BatchSize(size uint64) *QueryFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *QueryFixtureQuery) Limit(n uint64) *QueryFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *QueryFixtureQuery) Offset(n uint64) *QueryFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *QueryFixtureQuery) Where(cond kallax.Condition) *QueryFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

func (q *QueryFixtureQuery) WithRelation() *QueryFixtureQuery {
	q.AddRelation(Schema.QueryRelationFixture.BaseSchema, "Relation", kallax.OneToOne, nil)
	return q
}

func (q *QueryFixtureQuery) WithInverse() *QueryFixtureQuery {
	q.AddRelation(Schema.QueryRelationFixture.BaseSchema, "Inverse", kallax.OneToOne, nil)
	return q
}

func (q *QueryFixtureQuery) WithNRelation(cond kallax.Condition) *QueryFixtureQuery {
	q.AddRelation(Schema.QueryRelationFixture.BaseSchema, "NRelation", kallax.OneToMany, cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByID(v ...kallax.ULID) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.QueryFixture.ID, values...))
}

// FindByInverse adds a new filter to the query that will require that
// the foreign key of Inverse is equal to the passed value.
func (q *QueryFixtureQuery) FindByInverse(v kallax.ULID) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.InverseFK, v))
}

// FindByInline adds a new filter to the query that will require that
// the Inline property is equal to the passed value.
func (q *QueryFixtureQuery) FindByInline(v string) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.Inline, v))
}

// FindByFoo adds a new filter to the query that will require that
// the Foo property is equal to the passed value.
func (q *QueryFixtureQuery) FindByFoo(v string) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.Foo, v))
}

// FindByStringProperty adds a new filter to the query that will require that
// the StringProperty property is equal to the passed value.
func (q *QueryFixtureQuery) FindByStringProperty(v string) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.StringProperty, v))
}

// FindByInteger adds a new filter to the query that will require that
// the Integer property is equal to the passed value.
func (q *QueryFixtureQuery) FindByInteger(cond kallax.ScalarCond, v int) *QueryFixtureQuery {
	return q.Where(cond(Schema.QueryFixture.Integer, v))
}

// FindByInteger64 adds a new filter to the query that will require that
// the Integer64 property is equal to the passed value.
func (q *QueryFixtureQuery) FindByInteger64(cond kallax.ScalarCond, v int64) *QueryFixtureQuery {
	return q.Where(cond(Schema.QueryFixture.Integer64, v))
}

// FindByFloat32 adds a new filter to the query that will require that
// the Float32 property is equal to the passed value.
func (q *QueryFixtureQuery) FindByFloat32(cond kallax.ScalarCond, v float32) *QueryFixtureQuery {
	return q.Where(cond(Schema.QueryFixture.Float32, v))
}

// FindByBoolean adds a new filter to the query that will require that
// the Boolean property is equal to the passed value.
func (q *QueryFixtureQuery) FindByBoolean(v bool) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.Boolean, v))
}

// FindByArrayParam adds a new filter to the query that will require that
// the ArrayParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByArrayParam(v ...string) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.ArrayParam, values...))
}

// FindBySliceParam adds a new filter to the query that will require that
// the SliceParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindBySliceParam(v ...string) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.SliceParam, values...))
}

// FindByAliasArrayParam adds a new filter to the query that will require that
// the AliasArrayParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByAliasArrayParam(v ...string) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.AliasArrayParam, values...))
}

// FindByAliasSliceParam adds a new filter to the query that will require that
// the AliasSliceParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByAliasSliceParam(v ...string) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.AliasSliceParam, values...))
}

// FindByAliasStringParam adds a new filter to the query that will require that
// the AliasStringParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByAliasStringParam(v fixtures.AliasString) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.AliasStringParam, v))
}

// FindByAliasIntParam adds a new filter to the query that will require that
// the AliasIntParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByAliasIntParam(cond kallax.ScalarCond, v fixtures.AliasInt) *QueryFixtureQuery {
	return q.Where(cond(Schema.QueryFixture.AliasIntParam, v))
}

// FindByIDPropertyParam adds a new filter to the query that will require that
// the IDPropertyParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByIDPropertyParam(v kallax.ULID) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.IDPropertyParam, v))
}

// FindByInterfacePropParam adds a new filter to the query that will require that
// the InterfacePropParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByInterfacePropParam(v fixtures.InterfaceImplementation) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.InterfacePropParam, v))
}

// FindByURLParam adds a new filter to the query that will require that
// the URLParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByURLParam(v url.URL) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.URLParam, v))
}

// FindByTimeParam adds a new filter to the query that will require that
// the TimeParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByTimeParam(cond kallax.ScalarCond, v time.Time) *QueryFixtureQuery {
	return q.Where(cond(Schema.QueryFixture.TimeParam, v))
}

// FindByAliasArrAliasStringParam adds a new filter to the query that will require that
// the AliasArrAliasStringParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByAliasArrAliasStringParam(v ...fixtures.AliasString) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.AliasArrAliasStringParam, values...))
}

// FindByAliasHereArrayParam adds a new filter to the query that will require that
// the AliasHereArrayParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByAliasHereArrayParam(v ...string) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.AliasHereArrayParam, values...))
}

// FindByArrayAliasHereStringParam adds a new filter to the query that will require that
// the ArrayAliasHereStringParam property contains all the passed values; if no passed values,
// it will do nothing.
func (q *QueryFixtureQuery) FindByArrayAliasHereStringParam(v ...AliasHereString) *QueryFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.QueryFixture.ArrayAliasHereStringParam, values...))
}

// FindByScannerValuerParam adds a new filter to the query that will require that
// the ScannerValuerParam property is equal to the passed value.
func (q *QueryFixtureQuery) FindByScannerValuerParam(v ScannerValuer) *QueryFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryFixture.ScannerValuerParam, v))
}

// QueryFixtureResultSet is the set of results returned by a query to the
// database.
type QueryFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *QueryFixture
	lastErr   error
}

// NewQueryFixtureResultSet creates a new result set for rows of the type
// QueryFixture.
func NewQueryFixtureResultSet(rs kallax.ResultSet) *QueryFixtureResultSet {
	return &QueryFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *QueryFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.QueryFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*QueryFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *QueryFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *QueryFixtureResultSet) Get() (*QueryFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *QueryFixtureResultSet) ForEach(fn func(*QueryFixture) error) error {
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
func (rs *QueryFixtureResultSet) All() ([]*QueryFixture, error) {
	var result []*QueryFixture
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
func (rs *QueryFixtureResultSet) One() (*QueryFixture, error) {
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
func (rs *QueryFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *QueryFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewQueryRelationFixture returns a new instance of QueryRelationFixture.
func NewQueryRelationFixture() (record *QueryRelationFixture) {
	return new(QueryRelationFixture)
}

// GetID returns the primary key of the model.
func (r *QueryRelationFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *QueryRelationFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "name":
		return &r.Name, nil
	case "owner_id":
		return types.Nullable(kallax.VirtualColumn("owner_id", r, new(kallax.ULID))), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in QueryRelationFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *QueryRelationFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "name":
		return r.Name, nil
	case "owner_id":
		return r.Model.VirtualColumn(col), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in QueryRelationFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *QueryRelationFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	switch field {
	case "Owner":
		return new(QueryFixture), nil

	}
	return nil, fmt.Errorf("kallax: model QueryRelationFixture has no relationship %s", field)
}

// SetRelationship sets the given relationship in the given field.
func (r *QueryRelationFixture) SetRelationship(field string, rel interface{}) error {
	switch field {
	case "Owner":
		val, ok := rel.(*QueryFixture)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Owner", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Owner = val
		}

		return nil

	}
	return fmt.Errorf("kallax: model QueryRelationFixture has no relationship %s", field)
}

// QueryRelationFixtureStore is the entity to access the records of the type QueryRelationFixture
// in the database.
type QueryRelationFixtureStore struct {
	*kallax.Store
}

// NewQueryRelationFixtureStore creates a new instance of QueryRelationFixtureStore
// using a SQL database.
func NewQueryRelationFixtureStore(db *sql.DB) *QueryRelationFixtureStore {
	return &QueryRelationFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *QueryRelationFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *QueryRelationFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

func (s *QueryRelationFixtureStore) inverseRecords(record *QueryRelationFixture) []kallax.RecordWithSchema {
	record.ClearVirtualColumns()
	var records []kallax.RecordWithSchema

	if record.Owner != nil {
		record.AddVirtualColumn("owner_id", record.Owner.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.QueryFixture.BaseSchema,
			Record: record.Owner,
		})
	}

	return records
}

// Insert inserts a QueryRelationFixture in the database. A non-persisted object is
// required for this operation.
func (s *QueryRelationFixtureStore) Insert(record *QueryRelationFixture) error {

	inverseRecords := s.inverseRecords(record)

	if len(inverseRecords) > 0 {
		return s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			if err := s.Insert(Schema.QueryRelationFixture.BaseSchema, record); err != nil {
				return err
			}

			return nil
		})
	}

	return s.Store.Insert(Schema.QueryRelationFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *QueryRelationFixtureStore) Update(record *QueryRelationFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	inverseRecords := s.inverseRecords(record)

	if len(inverseRecords) > 0 {
		err = s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			updated, err = s.Update(Schema.QueryRelationFixture.BaseSchema, record, cols...)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return 0, err
		}

		return updated, nil
	}

	return s.Store.Update(Schema.QueryRelationFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *QueryRelationFixtureStore) Save(record *QueryRelationFixture) (updated bool, err error) {
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
func (s *QueryRelationFixtureStore) Delete(record *QueryRelationFixture) error {

	return s.Store.Delete(Schema.QueryRelationFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *QueryRelationFixtureStore) Find(q *QueryRelationFixtureQuery) (*QueryRelationFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewQueryRelationFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *QueryRelationFixtureStore) MustFind(q *QueryRelationFixtureQuery) *QueryRelationFixtureResultSet {
	return NewQueryRelationFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *QueryRelationFixtureStore) Count(q *QueryRelationFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *QueryRelationFixtureStore) MustCount(q *QueryRelationFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *QueryRelationFixtureStore) FindOne(q *QueryRelationFixtureQuery) (*QueryRelationFixture, error) {
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
func (s *QueryRelationFixtureStore) MustFindOne(q *QueryRelationFixtureQuery) *QueryRelationFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the QueryRelationFixture with the data in the database and
// makes it writable.
func (s *QueryRelationFixtureStore) Reload(record *QueryRelationFixture) error {
	return s.Store.Reload(Schema.QueryRelationFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *QueryRelationFixtureStore) Transaction(callback func(*QueryRelationFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&QueryRelationFixtureStore{store})
	})
}

// QueryRelationFixtureQuery is the object used to create queries for the QueryRelationFixture
// entity.
type QueryRelationFixtureQuery struct {
	*kallax.BaseQuery
}

// NewQueryRelationFixtureQuery returns a new instance of QueryRelationFixtureQuery.
func NewQueryRelationFixtureQuery() *QueryRelationFixtureQuery {
	return &QueryRelationFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.QueryRelationFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *QueryRelationFixtureQuery) Select(columns ...kallax.SchemaField) *QueryRelationFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *QueryRelationFixtureQuery) SelectNot(columns ...kallax.SchemaField) *QueryRelationFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *QueryRelationFixtureQuery) Copy() *QueryRelationFixtureQuery {
	return &QueryRelationFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *QueryRelationFixtureQuery) Order(cols ...kallax.ColumnOrder) *QueryRelationFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *QueryRelationFixtureQuery) BatchSize(size uint64) *QueryRelationFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *QueryRelationFixtureQuery) Limit(n uint64) *QueryRelationFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *QueryRelationFixtureQuery) Offset(n uint64) *QueryRelationFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *QueryRelationFixtureQuery) Where(cond kallax.Condition) *QueryRelationFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

func (q *QueryRelationFixtureQuery) WithOwner() *QueryRelationFixtureQuery {
	q.AddRelation(Schema.QueryFixture.BaseSchema, "Owner", kallax.OneToOne, nil)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *QueryRelationFixtureQuery) FindByID(v ...kallax.ULID) *QueryRelationFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.QueryRelationFixture.ID, values...))
}

// FindByName adds a new filter to the query that will require that
// the Name property is equal to the passed value.
func (q *QueryRelationFixtureQuery) FindByName(v string) *QueryRelationFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryRelationFixture.Name, v))
}

// FindByOwner adds a new filter to the query that will require that
// the foreign key of Owner is equal to the passed value.
func (q *QueryRelationFixtureQuery) FindByOwner(v kallax.ULID) *QueryRelationFixtureQuery {
	return q.Where(kallax.Eq(Schema.QueryRelationFixture.OwnerFK, v))
}

// QueryRelationFixtureResultSet is the set of results returned by a query to the
// database.
type QueryRelationFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *QueryRelationFixture
	lastErr   error
}

// NewQueryRelationFixtureResultSet creates a new result set for rows of the type
// QueryRelationFixture.
func NewQueryRelationFixtureResultSet(rs kallax.ResultSet) *QueryRelationFixtureResultSet {
	return &QueryRelationFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *QueryRelationFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.QueryRelationFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*QueryRelationFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *QueryRelationFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *QueryRelationFixtureResultSet) Get() (*QueryRelationFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *QueryRelationFixtureResultSet) ForEach(fn func(*QueryRelationFixture) error) error {
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
func (rs *QueryRelationFixtureResultSet) All() ([]*QueryRelationFixture, error) {
	var result []*QueryRelationFixture
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
func (rs *QueryRelationFixtureResultSet) One() (*QueryRelationFixture, error) {
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
func (rs *QueryRelationFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *QueryRelationFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewResultSetFixture returns a new instance of ResultSetFixture.
func NewResultSetFixture(f string) (record *ResultSetFixture) {
	return newResultSetFixture(f)
}

// GetID returns the primary key of the model.
func (r *ResultSetFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *ResultSetFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "foo":
		return &r.Foo, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in ResultSetFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *ResultSetFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "foo":
		return r.Foo, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in ResultSetFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *ResultSetFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model ResultSetFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *ResultSetFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model ResultSetFixture has no relationships")
}

// ResultSetFixtureStore is the entity to access the records of the type ResultSetFixture
// in the database.
type ResultSetFixtureStore struct {
	*kallax.Store
}

// NewResultSetFixtureStore creates a new instance of ResultSetFixtureStore
// using a SQL database.
func NewResultSetFixtureStore(db *sql.DB) *ResultSetFixtureStore {
	return &ResultSetFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *ResultSetFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *ResultSetFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a ResultSetFixture in the database. A non-persisted object is
// required for this operation.
func (s *ResultSetFixtureStore) Insert(record *ResultSetFixture) error {

	return s.Store.Insert(Schema.ResultSetFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *ResultSetFixtureStore) Update(record *ResultSetFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.ResultSetFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *ResultSetFixtureStore) Save(record *ResultSetFixture) (updated bool, err error) {
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
func (s *ResultSetFixtureStore) Delete(record *ResultSetFixture) error {

	return s.Store.Delete(Schema.ResultSetFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *ResultSetFixtureStore) Find(q *ResultSetFixtureQuery) (*ResultSetFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewResultSetFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *ResultSetFixtureStore) MustFind(q *ResultSetFixtureQuery) *ResultSetFixtureResultSet {
	return NewResultSetFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *ResultSetFixtureStore) Count(q *ResultSetFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *ResultSetFixtureStore) MustCount(q *ResultSetFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *ResultSetFixtureStore) FindOne(q *ResultSetFixtureQuery) (*ResultSetFixture, error) {
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
func (s *ResultSetFixtureStore) MustFindOne(q *ResultSetFixtureQuery) *ResultSetFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the ResultSetFixture with the data in the database and
// makes it writable.
func (s *ResultSetFixtureStore) Reload(record *ResultSetFixture) error {
	return s.Store.Reload(Schema.ResultSetFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *ResultSetFixtureStore) Transaction(callback func(*ResultSetFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&ResultSetFixtureStore{store})
	})
}

// ResultSetFixtureQuery is the object used to create queries for the ResultSetFixture
// entity.
type ResultSetFixtureQuery struct {
	*kallax.BaseQuery
}

// NewResultSetFixtureQuery returns a new instance of ResultSetFixtureQuery.
func NewResultSetFixtureQuery() *ResultSetFixtureQuery {
	return &ResultSetFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.ResultSetFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *ResultSetFixtureQuery) Select(columns ...kallax.SchemaField) *ResultSetFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *ResultSetFixtureQuery) SelectNot(columns ...kallax.SchemaField) *ResultSetFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *ResultSetFixtureQuery) Copy() *ResultSetFixtureQuery {
	return &ResultSetFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *ResultSetFixtureQuery) Order(cols ...kallax.ColumnOrder) *ResultSetFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *ResultSetFixtureQuery) BatchSize(size uint64) *ResultSetFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *ResultSetFixtureQuery) Limit(n uint64) *ResultSetFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *ResultSetFixtureQuery) Offset(n uint64) *ResultSetFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *ResultSetFixtureQuery) Where(cond kallax.Condition) *ResultSetFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *ResultSetFixtureQuery) FindByID(v ...kallax.ULID) *ResultSetFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.ResultSetFixture.ID, values...))
}

// FindByFoo adds a new filter to the query that will require that
// the Foo property is equal to the passed value.
func (q *ResultSetFixtureQuery) FindByFoo(v string) *ResultSetFixtureQuery {
	return q.Where(kallax.Eq(Schema.ResultSetFixture.Foo, v))
}

// ResultSetFixtureResultSet is the set of results returned by a query to the
// database.
type ResultSetFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *ResultSetFixture
	lastErr   error
}

// NewResultSetFixtureResultSet creates a new result set for rows of the type
// ResultSetFixture.
func NewResultSetFixtureResultSet(rs kallax.ResultSet) *ResultSetFixtureResultSet {
	return &ResultSetFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *ResultSetFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.ResultSetFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*ResultSetFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *ResultSetFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *ResultSetFixtureResultSet) Get() (*ResultSetFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *ResultSetFixtureResultSet) ForEach(fn func(*ResultSetFixture) error) error {
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
func (rs *ResultSetFixtureResultSet) All() ([]*ResultSetFixture, error) {
	var result []*ResultSetFixture
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
func (rs *ResultSetFixtureResultSet) One() (*ResultSetFixture, error) {
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
func (rs *ResultSetFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *ResultSetFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewSchemaFixture returns a new instance of SchemaFixture.
func NewSchemaFixture() (record *SchemaFixture) {
	return newSchemaFixture()
}

// GetID returns the primary key of the model.
func (r *SchemaFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *SchemaFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "string":
		return &r.String, nil
	case "int":
		return &r.Int, nil
	case "inline":
		return &r.Inline.Inline, nil
	case "map_of_string":
		return types.JSON(&r.MapOfString), nil
	case "map_of_interface":
		return types.JSON(&r.MapOfInterface), nil
	case "map_of_some_type":
		return types.JSON(&r.MapOfSomeType), nil
	case "rel_id":
		return types.Nullable(kallax.VirtualColumn("rel_id", r, new(kallax.ULID))), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in SchemaFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *SchemaFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "string":
		return r.String, nil
	case "int":
		return r.Int, nil
	case "inline":
		return r.Inline.Inline, nil
	case "map_of_string":
		return types.JSON(r.MapOfString), nil
	case "map_of_interface":
		return types.JSON(r.MapOfInterface), nil
	case "map_of_some_type":
		return types.JSON(r.MapOfSomeType), nil
	case "rel_id":
		return r.Model.VirtualColumn(col), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in SchemaFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *SchemaFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	switch field {
	case "Nested":
		return new(SchemaFixture), nil
	case "Inverse":
		return new(SchemaRelationshipFixture), nil

	}
	return nil, fmt.Errorf("kallax: model SchemaFixture has no relationship %s", field)
}

// SetRelationship sets the given relationship in the given field.
func (r *SchemaFixture) SetRelationship(field string, rel interface{}) error {
	switch field {
	case "Nested":
		val, ok := rel.(*SchemaFixture)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Nested", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Nested = val
		}

		return nil
	case "Inverse":
		val, ok := rel.(*SchemaRelationshipFixture)
		if !ok {
			return fmt.Errorf("kallax: record of type %t can't be assigned to relationship Inverse", rel)
		}
		if !val.GetID().IsEmpty() {
			r.Inverse = val
		}

		return nil

	}
	return fmt.Errorf("kallax: model SchemaFixture has no relationship %s", field)
}

// SchemaFixtureStore is the entity to access the records of the type SchemaFixture
// in the database.
type SchemaFixtureStore struct {
	*kallax.Store
}

// NewSchemaFixtureStore creates a new instance of SchemaFixtureStore
// using a SQL database.
func NewSchemaFixtureStore(db *sql.DB) *SchemaFixtureStore {
	return &SchemaFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *SchemaFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *SchemaFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

func (s *SchemaFixtureStore) relationshipRecords(record *SchemaFixture) []kallax.RecordWithSchema {
	var records []kallax.RecordWithSchema

	if record.Nested != nil {
		record.Nested.ClearVirtualColumns()
		record.Nested.AddVirtualColumn("schema_fixture_id", record.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.SchemaFixture.BaseSchema,
			Record: record.Nested,
		})
	}

	return records
}

func (s *SchemaFixtureStore) inverseRecords(record *SchemaFixture) []kallax.RecordWithSchema {
	record.ClearVirtualColumns()
	var records []kallax.RecordWithSchema

	if record.Inverse != nil {
		record.AddVirtualColumn("rel_id", record.Inverse.GetID())
		records = append(records, kallax.RecordWithSchema{
			Schema: Schema.SchemaRelationshipFixture.BaseSchema,
			Record: record.Inverse,
		})
	}

	return records
}

// Insert inserts a SchemaFixture in the database. A non-persisted object is
// required for this operation.
func (s *SchemaFixtureStore) Insert(record *SchemaFixture) error {

	records := s.relationshipRecords(record)

	inverseRecords := s.inverseRecords(record)

	if len(records) > 0 && len(inverseRecords) > 0 {
		return s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			if err := s.Insert(Schema.SchemaFixture.BaseSchema, record); err != nil {
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

	return s.Store.Insert(Schema.SchemaFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *SchemaFixtureStore) Update(record *SchemaFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	records := s.relationshipRecords(record)

	inverseRecords := s.inverseRecords(record)

	if len(records) > 0 && len(inverseRecords) > 0 {
		err = s.Store.Transaction(func(s *kallax.Store) error {

			for _, r := range inverseRecords {
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

			updated, err = s.Update(Schema.SchemaFixture.BaseSchema, record, cols...)
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

	return s.Store.Update(Schema.SchemaFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *SchemaFixtureStore) Save(record *SchemaFixture) (updated bool, err error) {
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
func (s *SchemaFixtureStore) Delete(record *SchemaFixture) error {

	return s.Store.Delete(Schema.SchemaFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *SchemaFixtureStore) Find(q *SchemaFixtureQuery) (*SchemaFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewSchemaFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *SchemaFixtureStore) MustFind(q *SchemaFixtureQuery) *SchemaFixtureResultSet {
	return NewSchemaFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *SchemaFixtureStore) Count(q *SchemaFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *SchemaFixtureStore) MustCount(q *SchemaFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *SchemaFixtureStore) FindOne(q *SchemaFixtureQuery) (*SchemaFixture, error) {
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
func (s *SchemaFixtureStore) MustFindOne(q *SchemaFixtureQuery) *SchemaFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the SchemaFixture with the data in the database and
// makes it writable.
func (s *SchemaFixtureStore) Reload(record *SchemaFixture) error {
	return s.Store.Reload(Schema.SchemaFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *SchemaFixtureStore) Transaction(callback func(*SchemaFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&SchemaFixtureStore{store})
	})
}

// RemoveNested removes from the database the given relationship of the
// model. It also resets the field Nested of the model.
func (s *SchemaFixtureStore) RemoveNested(record *SchemaFixture) error {
	var r kallax.Record = record.Nested
	if beforeDeleter, ok := r.(kallax.BeforeDeleter); ok {
		if err := beforeDeleter.BeforeDelete(); err != nil {
			return err
		}
	}

	var err error
	if afterDeleter, ok := r.(kallax.AfterDeleter); ok {
		err = s.Store.Transaction(func(s *kallax.Store) error {
			err := s.Delete(Schema.SchemaFixture.BaseSchema, r)
			if err != nil {
				return err
			}

			return afterDeleter.AfterDelete()
		})
	} else {
		err = s.Store.Delete(Schema.SchemaFixture.BaseSchema, r)
	}
	if err != nil {
		return err
	}

	record.Nested = nil
	return nil
}

// SchemaFixtureQuery is the object used to create queries for the SchemaFixture
// entity.
type SchemaFixtureQuery struct {
	*kallax.BaseQuery
}

// NewSchemaFixtureQuery returns a new instance of SchemaFixtureQuery.
func NewSchemaFixtureQuery() *SchemaFixtureQuery {
	return &SchemaFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.SchemaFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *SchemaFixtureQuery) Select(columns ...kallax.SchemaField) *SchemaFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *SchemaFixtureQuery) SelectNot(columns ...kallax.SchemaField) *SchemaFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *SchemaFixtureQuery) Copy() *SchemaFixtureQuery {
	return &SchemaFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *SchemaFixtureQuery) Order(cols ...kallax.ColumnOrder) *SchemaFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *SchemaFixtureQuery) BatchSize(size uint64) *SchemaFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *SchemaFixtureQuery) Limit(n uint64) *SchemaFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *SchemaFixtureQuery) Offset(n uint64) *SchemaFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *SchemaFixtureQuery) Where(cond kallax.Condition) *SchemaFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

func (q *SchemaFixtureQuery) WithNested() *SchemaFixtureQuery {
	q.AddRelation(Schema.SchemaFixture.BaseSchema, "Nested", kallax.OneToOne, nil)
	return q
}

func (q *SchemaFixtureQuery) WithInverse() *SchemaFixtureQuery {
	q.AddRelation(Schema.SchemaRelationshipFixture.BaseSchema, "Inverse", kallax.OneToOne, nil)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *SchemaFixtureQuery) FindByID(v ...kallax.ULID) *SchemaFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.SchemaFixture.ID, values...))
}

// FindByString adds a new filter to the query that will require that
// the String property is equal to the passed value.
func (q *SchemaFixtureQuery) FindByString(v string) *SchemaFixtureQuery {
	return q.Where(kallax.Eq(Schema.SchemaFixture.String, v))
}

// FindByInt adds a new filter to the query that will require that
// the Int property is equal to the passed value.
func (q *SchemaFixtureQuery) FindByInt(cond kallax.ScalarCond, v int) *SchemaFixtureQuery {
	return q.Where(cond(Schema.SchemaFixture.Int, v))
}

// FindByInline adds a new filter to the query that will require that
// the Inline property is equal to the passed value.
func (q *SchemaFixtureQuery) FindByInline(v string) *SchemaFixtureQuery {
	return q.Where(kallax.Eq(Schema.SchemaFixture.Inline, v))
}

// FindByInverse adds a new filter to the query that will require that
// the foreign key of Inverse is equal to the passed value.
func (q *SchemaFixtureQuery) FindByInverse(v kallax.ULID) *SchemaFixtureQuery {
	return q.Where(kallax.Eq(Schema.SchemaFixture.InverseFK, v))
}

// SchemaFixtureResultSet is the set of results returned by a query to the
// database.
type SchemaFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *SchemaFixture
	lastErr   error
}

// NewSchemaFixtureResultSet creates a new result set for rows of the type
// SchemaFixture.
func NewSchemaFixtureResultSet(rs kallax.ResultSet) *SchemaFixtureResultSet {
	return &SchemaFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *SchemaFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.SchemaFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*SchemaFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *SchemaFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *SchemaFixtureResultSet) Get() (*SchemaFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *SchemaFixtureResultSet) ForEach(fn func(*SchemaFixture) error) error {
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
func (rs *SchemaFixtureResultSet) All() ([]*SchemaFixture, error) {
	var result []*SchemaFixture
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
func (rs *SchemaFixtureResultSet) One() (*SchemaFixture, error) {
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
func (rs *SchemaFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *SchemaFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewSchemaRelationshipFixture returns a new instance of SchemaRelationshipFixture.
func NewSchemaRelationshipFixture() (record *SchemaRelationshipFixture) {
	return new(SchemaRelationshipFixture)
}

// GetID returns the primary key of the model.
func (r *SchemaRelationshipFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *SchemaRelationshipFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in SchemaRelationshipFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *SchemaRelationshipFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in SchemaRelationshipFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *SchemaRelationshipFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model SchemaRelationshipFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *SchemaRelationshipFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model SchemaRelationshipFixture has no relationships")
}

// SchemaRelationshipFixtureStore is the entity to access the records of the type SchemaRelationshipFixture
// in the database.
type SchemaRelationshipFixtureStore struct {
	*kallax.Store
}

// NewSchemaRelationshipFixtureStore creates a new instance of SchemaRelationshipFixtureStore
// using a SQL database.
func NewSchemaRelationshipFixtureStore(db *sql.DB) *SchemaRelationshipFixtureStore {
	return &SchemaRelationshipFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *SchemaRelationshipFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *SchemaRelationshipFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a SchemaRelationshipFixture in the database. A non-persisted object is
// required for this operation.
func (s *SchemaRelationshipFixtureStore) Insert(record *SchemaRelationshipFixture) error {

	return s.Store.Insert(Schema.SchemaRelationshipFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *SchemaRelationshipFixtureStore) Update(record *SchemaRelationshipFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.SchemaRelationshipFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *SchemaRelationshipFixtureStore) Save(record *SchemaRelationshipFixture) (updated bool, err error) {
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
func (s *SchemaRelationshipFixtureStore) Delete(record *SchemaRelationshipFixture) error {

	return s.Store.Delete(Schema.SchemaRelationshipFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *SchemaRelationshipFixtureStore) Find(q *SchemaRelationshipFixtureQuery) (*SchemaRelationshipFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewSchemaRelationshipFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *SchemaRelationshipFixtureStore) MustFind(q *SchemaRelationshipFixtureQuery) *SchemaRelationshipFixtureResultSet {
	return NewSchemaRelationshipFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *SchemaRelationshipFixtureStore) Count(q *SchemaRelationshipFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *SchemaRelationshipFixtureStore) MustCount(q *SchemaRelationshipFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *SchemaRelationshipFixtureStore) FindOne(q *SchemaRelationshipFixtureQuery) (*SchemaRelationshipFixture, error) {
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
func (s *SchemaRelationshipFixtureStore) MustFindOne(q *SchemaRelationshipFixtureQuery) *SchemaRelationshipFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the SchemaRelationshipFixture with the data in the database and
// makes it writable.
func (s *SchemaRelationshipFixtureStore) Reload(record *SchemaRelationshipFixture) error {
	return s.Store.Reload(Schema.SchemaRelationshipFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *SchemaRelationshipFixtureStore) Transaction(callback func(*SchemaRelationshipFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&SchemaRelationshipFixtureStore{store})
	})
}

// SchemaRelationshipFixtureQuery is the object used to create queries for the SchemaRelationshipFixture
// entity.
type SchemaRelationshipFixtureQuery struct {
	*kallax.BaseQuery
}

// NewSchemaRelationshipFixtureQuery returns a new instance of SchemaRelationshipFixtureQuery.
func NewSchemaRelationshipFixtureQuery() *SchemaRelationshipFixtureQuery {
	return &SchemaRelationshipFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.SchemaRelationshipFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *SchemaRelationshipFixtureQuery) Select(columns ...kallax.SchemaField) *SchemaRelationshipFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *SchemaRelationshipFixtureQuery) SelectNot(columns ...kallax.SchemaField) *SchemaRelationshipFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *SchemaRelationshipFixtureQuery) Copy() *SchemaRelationshipFixtureQuery {
	return &SchemaRelationshipFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *SchemaRelationshipFixtureQuery) Order(cols ...kallax.ColumnOrder) *SchemaRelationshipFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *SchemaRelationshipFixtureQuery) BatchSize(size uint64) *SchemaRelationshipFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *SchemaRelationshipFixtureQuery) Limit(n uint64) *SchemaRelationshipFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *SchemaRelationshipFixtureQuery) Offset(n uint64) *SchemaRelationshipFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *SchemaRelationshipFixtureQuery) Where(cond kallax.Condition) *SchemaRelationshipFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *SchemaRelationshipFixtureQuery) FindByID(v ...kallax.ULID) *SchemaRelationshipFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.SchemaRelationshipFixture.ID, values...))
}

// SchemaRelationshipFixtureResultSet is the set of results returned by a query to the
// database.
type SchemaRelationshipFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *SchemaRelationshipFixture
	lastErr   error
}

// NewSchemaRelationshipFixtureResultSet creates a new result set for rows of the type
// SchemaRelationshipFixture.
func NewSchemaRelationshipFixtureResultSet(rs kallax.ResultSet) *SchemaRelationshipFixtureResultSet {
	return &SchemaRelationshipFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *SchemaRelationshipFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.SchemaRelationshipFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*SchemaRelationshipFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *SchemaRelationshipFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *SchemaRelationshipFixtureResultSet) Get() (*SchemaRelationshipFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *SchemaRelationshipFixtureResultSet) ForEach(fn func(*SchemaRelationshipFixture) error) error {
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
func (rs *SchemaRelationshipFixtureResultSet) All() ([]*SchemaRelationshipFixture, error) {
	var result []*SchemaRelationshipFixture
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
func (rs *SchemaRelationshipFixtureResultSet) One() (*SchemaRelationshipFixture, error) {
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
func (rs *SchemaRelationshipFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *SchemaRelationshipFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewStoreFixture returns a new instance of StoreFixture.
func NewStoreFixture() (record *StoreFixture) {
	return newStoreFixture()
}

// GetID returns the primary key of the model.
func (r *StoreFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *StoreFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "foo":
		return &r.Foo, nil
	case "slice_prop":
		return types.Slice(&r.SliceProp), nil
	case "alias_slice_prop":
		return types.Slice((*[]string)(&r.AliasSliceProp)), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in StoreFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *StoreFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "foo":
		return r.Foo, nil
	case "slice_prop":
		return types.Slice(r.SliceProp), nil
	case "alias_slice_prop":
		return types.Slice(r.AliasSliceProp), nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in StoreFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *StoreFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model StoreFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *StoreFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model StoreFixture has no relationships")
}

// StoreFixtureStore is the entity to access the records of the type StoreFixture
// in the database.
type StoreFixtureStore struct {
	*kallax.Store
}

// NewStoreFixtureStore creates a new instance of StoreFixtureStore
// using a SQL database.
func NewStoreFixtureStore(db *sql.DB) *StoreFixtureStore {
	return &StoreFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *StoreFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *StoreFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a StoreFixture in the database. A non-persisted object is
// required for this operation.
func (s *StoreFixtureStore) Insert(record *StoreFixture) error {

	return s.Store.Insert(Schema.StoreFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *StoreFixtureStore) Update(record *StoreFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.StoreFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *StoreFixtureStore) Save(record *StoreFixture) (updated bool, err error) {
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
func (s *StoreFixtureStore) Delete(record *StoreFixture) error {

	return s.Store.Delete(Schema.StoreFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *StoreFixtureStore) Find(q *StoreFixtureQuery) (*StoreFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewStoreFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *StoreFixtureStore) MustFind(q *StoreFixtureQuery) *StoreFixtureResultSet {
	return NewStoreFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *StoreFixtureStore) Count(q *StoreFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *StoreFixtureStore) MustCount(q *StoreFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *StoreFixtureStore) FindOne(q *StoreFixtureQuery) (*StoreFixture, error) {
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
func (s *StoreFixtureStore) MustFindOne(q *StoreFixtureQuery) *StoreFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the StoreFixture with the data in the database and
// makes it writable.
func (s *StoreFixtureStore) Reload(record *StoreFixture) error {
	return s.Store.Reload(Schema.StoreFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *StoreFixtureStore) Transaction(callback func(*StoreFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&StoreFixtureStore{store})
	})
}

// StoreFixtureQuery is the object used to create queries for the StoreFixture
// entity.
type StoreFixtureQuery struct {
	*kallax.BaseQuery
}

// NewStoreFixtureQuery returns a new instance of StoreFixtureQuery.
func NewStoreFixtureQuery() *StoreFixtureQuery {
	return &StoreFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.StoreFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *StoreFixtureQuery) Select(columns ...kallax.SchemaField) *StoreFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *StoreFixtureQuery) SelectNot(columns ...kallax.SchemaField) *StoreFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *StoreFixtureQuery) Copy() *StoreFixtureQuery {
	return &StoreFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *StoreFixtureQuery) Order(cols ...kallax.ColumnOrder) *StoreFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *StoreFixtureQuery) BatchSize(size uint64) *StoreFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *StoreFixtureQuery) Limit(n uint64) *StoreFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *StoreFixtureQuery) Offset(n uint64) *StoreFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *StoreFixtureQuery) Where(cond kallax.Condition) *StoreFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *StoreFixtureQuery) FindByID(v ...kallax.ULID) *StoreFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.StoreFixture.ID, values...))
}

// FindByFoo adds a new filter to the query that will require that
// the Foo property is equal to the passed value.
func (q *StoreFixtureQuery) FindByFoo(v string) *StoreFixtureQuery {
	return q.Where(kallax.Eq(Schema.StoreFixture.Foo, v))
}

// FindBySliceProp adds a new filter to the query that will require that
// the SliceProp property contains all the passed values; if no passed values,
// it will do nothing.
func (q *StoreFixtureQuery) FindBySliceProp(v ...string) *StoreFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.StoreFixture.SliceProp, values...))
}

// FindByAliasSliceProp adds a new filter to the query that will require that
// the AliasSliceProp property contains all the passed values; if no passed values,
// it will do nothing.
func (q *StoreFixtureQuery) FindByAliasSliceProp(v ...string) *StoreFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.ArrayContains(Schema.StoreFixture.AliasSliceProp, values...))
}

// StoreFixtureResultSet is the set of results returned by a query to the
// database.
type StoreFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *StoreFixture
	lastErr   error
}

// NewStoreFixtureResultSet creates a new result set for rows of the type
// StoreFixture.
func NewStoreFixtureResultSet(rs kallax.ResultSet) *StoreFixtureResultSet {
	return &StoreFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *StoreFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.StoreFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*StoreFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *StoreFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *StoreFixtureResultSet) Get() (*StoreFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *StoreFixtureResultSet) ForEach(fn func(*StoreFixture) error) error {
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
func (rs *StoreFixtureResultSet) All() ([]*StoreFixture, error) {
	var result []*StoreFixture
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
func (rs *StoreFixtureResultSet) One() (*StoreFixture, error) {
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
func (rs *StoreFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *StoreFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewStoreWithConstructFixture returns a new instance of StoreWithConstructFixture.
func NewStoreWithConstructFixture(f string) (record *StoreWithConstructFixture) {
	return newStoreWithConstructFixture(f)
}

// GetID returns the primary key of the model.
func (r *StoreWithConstructFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *StoreWithConstructFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "foo":
		return &r.Foo, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in StoreWithConstructFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *StoreWithConstructFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "foo":
		return r.Foo, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in StoreWithConstructFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *StoreWithConstructFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model StoreWithConstructFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *StoreWithConstructFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model StoreWithConstructFixture has no relationships")
}

// StoreWithConstructFixtureStore is the entity to access the records of the type StoreWithConstructFixture
// in the database.
type StoreWithConstructFixtureStore struct {
	*kallax.Store
}

// NewStoreWithConstructFixtureStore creates a new instance of StoreWithConstructFixtureStore
// using a SQL database.
func NewStoreWithConstructFixtureStore(db *sql.DB) *StoreWithConstructFixtureStore {
	return &StoreWithConstructFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *StoreWithConstructFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *StoreWithConstructFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a StoreWithConstructFixture in the database. A non-persisted object is
// required for this operation.
func (s *StoreWithConstructFixtureStore) Insert(record *StoreWithConstructFixture) error {

	return s.Store.Insert(Schema.StoreWithConstructFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *StoreWithConstructFixtureStore) Update(record *StoreWithConstructFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.StoreWithConstructFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *StoreWithConstructFixtureStore) Save(record *StoreWithConstructFixture) (updated bool, err error) {
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
func (s *StoreWithConstructFixtureStore) Delete(record *StoreWithConstructFixture) error {

	return s.Store.Delete(Schema.StoreWithConstructFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *StoreWithConstructFixtureStore) Find(q *StoreWithConstructFixtureQuery) (*StoreWithConstructFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewStoreWithConstructFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *StoreWithConstructFixtureStore) MustFind(q *StoreWithConstructFixtureQuery) *StoreWithConstructFixtureResultSet {
	return NewStoreWithConstructFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *StoreWithConstructFixtureStore) Count(q *StoreWithConstructFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *StoreWithConstructFixtureStore) MustCount(q *StoreWithConstructFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *StoreWithConstructFixtureStore) FindOne(q *StoreWithConstructFixtureQuery) (*StoreWithConstructFixture, error) {
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
func (s *StoreWithConstructFixtureStore) MustFindOne(q *StoreWithConstructFixtureQuery) *StoreWithConstructFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the StoreWithConstructFixture with the data in the database and
// makes it writable.
func (s *StoreWithConstructFixtureStore) Reload(record *StoreWithConstructFixture) error {
	return s.Store.Reload(Schema.StoreWithConstructFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *StoreWithConstructFixtureStore) Transaction(callback func(*StoreWithConstructFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&StoreWithConstructFixtureStore{store})
	})
}

// StoreWithConstructFixtureQuery is the object used to create queries for the StoreWithConstructFixture
// entity.
type StoreWithConstructFixtureQuery struct {
	*kallax.BaseQuery
}

// NewStoreWithConstructFixtureQuery returns a new instance of StoreWithConstructFixtureQuery.
func NewStoreWithConstructFixtureQuery() *StoreWithConstructFixtureQuery {
	return &StoreWithConstructFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.StoreWithConstructFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *StoreWithConstructFixtureQuery) Select(columns ...kallax.SchemaField) *StoreWithConstructFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *StoreWithConstructFixtureQuery) SelectNot(columns ...kallax.SchemaField) *StoreWithConstructFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *StoreWithConstructFixtureQuery) Copy() *StoreWithConstructFixtureQuery {
	return &StoreWithConstructFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *StoreWithConstructFixtureQuery) Order(cols ...kallax.ColumnOrder) *StoreWithConstructFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *StoreWithConstructFixtureQuery) BatchSize(size uint64) *StoreWithConstructFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *StoreWithConstructFixtureQuery) Limit(n uint64) *StoreWithConstructFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *StoreWithConstructFixtureQuery) Offset(n uint64) *StoreWithConstructFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *StoreWithConstructFixtureQuery) Where(cond kallax.Condition) *StoreWithConstructFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *StoreWithConstructFixtureQuery) FindByID(v ...kallax.ULID) *StoreWithConstructFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.StoreWithConstructFixture.ID, values...))
}

// FindByFoo adds a new filter to the query that will require that
// the Foo property is equal to the passed value.
func (q *StoreWithConstructFixtureQuery) FindByFoo(v string) *StoreWithConstructFixtureQuery {
	return q.Where(kallax.Eq(Schema.StoreWithConstructFixture.Foo, v))
}

// StoreWithConstructFixtureResultSet is the set of results returned by a query to the
// database.
type StoreWithConstructFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *StoreWithConstructFixture
	lastErr   error
}

// NewStoreWithConstructFixtureResultSet creates a new result set for rows of the type
// StoreWithConstructFixture.
func NewStoreWithConstructFixtureResultSet(rs kallax.ResultSet) *StoreWithConstructFixtureResultSet {
	return &StoreWithConstructFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *StoreWithConstructFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.StoreWithConstructFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*StoreWithConstructFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *StoreWithConstructFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *StoreWithConstructFixtureResultSet) Get() (*StoreWithConstructFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *StoreWithConstructFixtureResultSet) ForEach(fn func(*StoreWithConstructFixture) error) error {
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
func (rs *StoreWithConstructFixtureResultSet) All() ([]*StoreWithConstructFixture, error) {
	var result []*StoreWithConstructFixture
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
func (rs *StoreWithConstructFixtureResultSet) One() (*StoreWithConstructFixture, error) {
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
func (rs *StoreWithConstructFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *StoreWithConstructFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

// NewStoreWithNewFixture returns a new instance of StoreWithNewFixture.
func NewStoreWithNewFixture() (record *StoreWithNewFixture) {
	return newStoreWithNewFixture()
}

// GetID returns the primary key of the model.
func (r *StoreWithNewFixture) GetID() kallax.Identifier {
	return (*kallax.ULID)(&r.ID)
}

// ColumnAddress returns the pointer to the value of the given column.
func (r *StoreWithNewFixture) ColumnAddress(col string) (interface{}, error) {
	switch col {
	case "id":
		return (*kallax.ULID)(&r.ID), nil
	case "foo":
		return &r.Foo, nil
	case "bar":
		return &r.Bar, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in StoreWithNewFixture: %s", col)
	}
}

// Value returns the value of the given column.
func (r *StoreWithNewFixture) Value(col string) (interface{}, error) {
	switch col {
	case "id":
		return r.ID, nil
	case "foo":
		return r.Foo, nil
	case "bar":
		return r.Bar, nil

	default:
		return nil, fmt.Errorf("kallax: invalid column in StoreWithNewFixture: %s", col)
	}
}

// NewRelationshipRecord returns a new record for the relatiobship in the given
// field.
func (r *StoreWithNewFixture) NewRelationshipRecord(field string) (kallax.Record, error) {
	return nil, fmt.Errorf("kallax: model StoreWithNewFixture has no relationships")
}

// SetRelationship sets the given relationship in the given field.
func (r *StoreWithNewFixture) SetRelationship(field string, rel interface{}) error {
	return fmt.Errorf("kallax: model StoreWithNewFixture has no relationships")
}

// StoreWithNewFixtureStore is the entity to access the records of the type StoreWithNewFixture
// in the database.
type StoreWithNewFixtureStore struct {
	*kallax.Store
}

// NewStoreWithNewFixtureStore creates a new instance of StoreWithNewFixtureStore
// using a SQL database.
func NewStoreWithNewFixtureStore(db *sql.DB) *StoreWithNewFixtureStore {
	return &StoreWithNewFixtureStore{kallax.NewStore(db)}
}

// GenericStore returns the generic store of this store.
func (s *StoreWithNewFixtureStore) GenericStore() *kallax.Store {
	return s.Store
}

// SetGenericStore changes the generic store of this store.
func (s *StoreWithNewFixtureStore) SetGenericStore(store *kallax.Store) {
	s.Store = store
}

// Insert inserts a StoreWithNewFixture in the database. A non-persisted object is
// required for this operation.
func (s *StoreWithNewFixtureStore) Insert(record *StoreWithNewFixture) error {

	return s.Store.Insert(Schema.StoreWithNewFixture.BaseSchema, record)

}

// Update updates the given record on the database. If the columns are given,
// only these columns will be updated. Otherwise all of them will be.
// Be very careful with this, as you will have a potentially different object
// in memory but not on the database.
// Only writable records can be updated. Writable objects are those that have
// been just inserted or retrieved using a query with no custom select fields.
func (s *StoreWithNewFixtureStore) Update(record *StoreWithNewFixture, cols ...kallax.SchemaField) (updated int64, err error) {

	return s.Store.Update(Schema.StoreWithNewFixture.BaseSchema, record, cols...)

}

// Save inserts the object if the record is not persisted, otherwise it updates
// it. Same rules of Update and Insert apply depending on the case.
func (s *StoreWithNewFixtureStore) Save(record *StoreWithNewFixture) (updated bool, err error) {
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
func (s *StoreWithNewFixtureStore) Delete(record *StoreWithNewFixture) error {

	return s.Store.Delete(Schema.StoreWithNewFixture.BaseSchema, record)

}

// Find returns the set of results for the given query.
func (s *StoreWithNewFixtureStore) Find(q *StoreWithNewFixtureQuery) (*StoreWithNewFixtureResultSet, error) {
	rs, err := s.Store.Find(q)
	if err != nil {
		return nil, err
	}

	return NewStoreWithNewFixtureResultSet(rs), nil
}

// MustFind returns the set of results for the given query, but panics if there
// is any error.
func (s *StoreWithNewFixtureStore) MustFind(q *StoreWithNewFixtureQuery) *StoreWithNewFixtureResultSet {
	return NewStoreWithNewFixtureResultSet(s.Store.MustFind(q))
}

// Count returns the number of rows that would be retrieved with the given
// query.
func (s *StoreWithNewFixtureStore) Count(q *StoreWithNewFixtureQuery) (int64, error) {
	return s.Store.Count(q)
}

// MustCount returns the number of rows that would be retrieved with the given
// query, but panics if there is an error.
func (s *StoreWithNewFixtureStore) MustCount(q *StoreWithNewFixtureQuery) int64 {
	return s.Store.MustCount(q)
}

// FindOne returns the first row returned by the given query.
// `ErrNotFound` is returned if there are no results.
func (s *StoreWithNewFixtureStore) FindOne(q *StoreWithNewFixtureQuery) (*StoreWithNewFixture, error) {
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
func (s *StoreWithNewFixtureStore) MustFindOne(q *StoreWithNewFixtureQuery) *StoreWithNewFixture {
	record, err := s.FindOne(q)
	if err != nil {
		panic(err)
	}
	return record
}

// Reload refreshes the StoreWithNewFixture with the data in the database and
// makes it writable.
func (s *StoreWithNewFixtureStore) Reload(record *StoreWithNewFixture) error {
	return s.Store.Reload(Schema.StoreWithNewFixture.BaseSchema, record)
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *StoreWithNewFixtureStore) Transaction(callback func(*StoreWithNewFixtureStore) error) error {
	if callback == nil {
		return kallax.ErrInvalidTxCallback
	}

	return s.Store.Transaction(func(store *kallax.Store) error {
		return callback(&StoreWithNewFixtureStore{store})
	})
}

// StoreWithNewFixtureQuery is the object used to create queries for the StoreWithNewFixture
// entity.
type StoreWithNewFixtureQuery struct {
	*kallax.BaseQuery
}

// NewStoreWithNewFixtureQuery returns a new instance of StoreWithNewFixtureQuery.
func NewStoreWithNewFixtureQuery() *StoreWithNewFixtureQuery {
	return &StoreWithNewFixtureQuery{
		BaseQuery: kallax.NewBaseQuery(Schema.StoreWithNewFixture.BaseSchema),
	}
}

// Select adds columns to select in the query.
func (q *StoreWithNewFixtureQuery) Select(columns ...kallax.SchemaField) *StoreWithNewFixtureQuery {
	if len(columns) == 0 {
		return q
	}
	q.BaseQuery.Select(columns...)
	return q
}

// SelectNot excludes columns from being selected in the query.
func (q *StoreWithNewFixtureQuery) SelectNot(columns ...kallax.SchemaField) *StoreWithNewFixtureQuery {
	q.BaseQuery.SelectNot(columns...)
	return q
}

// Copy returns a new identical copy of the query. Remember queries are mutable
// so make a copy any time you need to reuse them.
func (q *StoreWithNewFixtureQuery) Copy() *StoreWithNewFixtureQuery {
	return &StoreWithNewFixtureQuery{
		BaseQuery: q.BaseQuery.Copy(),
	}
}

// Order adds order clauses to the query for the given columns.
func (q *StoreWithNewFixtureQuery) Order(cols ...kallax.ColumnOrder) *StoreWithNewFixtureQuery {
	q.BaseQuery.Order(cols...)
	return q
}

// BatchSize sets the number of items to fetch per batch when there are 1:N
// relationships selected in the query.
func (q *StoreWithNewFixtureQuery) BatchSize(size uint64) *StoreWithNewFixtureQuery {
	q.BaseQuery.BatchSize(size)
	return q
}

// Limit sets the max number of items to retrieve.
func (q *StoreWithNewFixtureQuery) Limit(n uint64) *StoreWithNewFixtureQuery {
	q.BaseQuery.Limit(n)
	return q
}

// Offset sets the number of items to skip from the result set of items.
func (q *StoreWithNewFixtureQuery) Offset(n uint64) *StoreWithNewFixtureQuery {
	q.BaseQuery.Offset(n)
	return q
}

// Where adds a condition to the query. All conditions added are concatenated
// using a logical AND.
func (q *StoreWithNewFixtureQuery) Where(cond kallax.Condition) *StoreWithNewFixtureQuery {
	q.BaseQuery.Where(cond)
	return q
}

// FindByID adds a new filter to the query that will require that
// the ID property is equal to one of the passed values; if no passed values,
// it will do nothing.
func (q *StoreWithNewFixtureQuery) FindByID(v ...kallax.ULID) *StoreWithNewFixtureQuery {
	if len(v) == 0 {
		return q
	}
	values := make([]interface{}, len(v))
	for i, val := range v {
		values[i] = val
	}
	return q.Where(kallax.In(Schema.StoreWithNewFixture.ID, values...))
}

// FindByFoo adds a new filter to the query that will require that
// the Foo property is equal to the passed value.
func (q *StoreWithNewFixtureQuery) FindByFoo(v string) *StoreWithNewFixtureQuery {
	return q.Where(kallax.Eq(Schema.StoreWithNewFixture.Foo, v))
}

// FindByBar adds a new filter to the query that will require that
// the Bar property is equal to the passed value.
func (q *StoreWithNewFixtureQuery) FindByBar(v string) *StoreWithNewFixtureQuery {
	return q.Where(kallax.Eq(Schema.StoreWithNewFixture.Bar, v))
}

// StoreWithNewFixtureResultSet is the set of results returned by a query to the
// database.
type StoreWithNewFixtureResultSet struct {
	ResultSet kallax.ResultSet
	last      *StoreWithNewFixture
	lastErr   error
}

// NewStoreWithNewFixtureResultSet creates a new result set for rows of the type
// StoreWithNewFixture.
func NewStoreWithNewFixtureResultSet(rs kallax.ResultSet) *StoreWithNewFixtureResultSet {
	return &StoreWithNewFixtureResultSet{ResultSet: rs}
}

// Next fetches the next item in the result set and returns true if there is
// a next item.
// The result set is closed automatically when there are no more items.
func (rs *StoreWithNewFixtureResultSet) Next() bool {
	if !rs.ResultSet.Next() {
		rs.lastErr = rs.ResultSet.Close()
		rs.last = nil
		return false
	}

	var record kallax.Record
	record, rs.lastErr = rs.ResultSet.Get(Schema.StoreWithNewFixture.BaseSchema)
	if rs.lastErr != nil {
		rs.last = nil
	} else {
		var ok bool
		rs.last, ok = record.(*StoreWithNewFixture)
		if !ok {
			rs.lastErr = fmt.Errorf("kallax: unable to convert record to *StoreWithNewFixture")
			rs.last = nil
		}
	}

	return true
}

// Get retrieves the last fetched item from the result set and the last error.
func (rs *StoreWithNewFixtureResultSet) Get() (*StoreWithNewFixture, error) {
	return rs.last, rs.lastErr
}

// ForEach iterates over the complete result set passing every record found to
// the given callback. It is possible to stop the iteration by returning
// `kallax.ErrStop` in the callback.
// Result set is always closed at the end.
func (rs *StoreWithNewFixtureResultSet) ForEach(fn func(*StoreWithNewFixture) error) error {
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
func (rs *StoreWithNewFixtureResultSet) All() ([]*StoreWithNewFixture, error) {
	var result []*StoreWithNewFixture
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
func (rs *StoreWithNewFixtureResultSet) One() (*StoreWithNewFixture, error) {
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
func (rs *StoreWithNewFixtureResultSet) Err() error {
	return rs.lastErr
}

// Close closes the result set.
func (rs *StoreWithNewFixtureResultSet) Close() error {
	return rs.ResultSet.Close()
}

type schema struct {
	Car                       *schemaCar
	EventsAllFixture          *schemaEventsAllFixture
	EventsFixture             *schemaEventsFixture
	EventsSaveFixture         *schemaEventsSaveFixture
	JSONModel                 *schemaJSONModel
	MultiKeySortFixture       *schemaMultiKeySortFixture
	Nullable                  *schemaNullable
	Person                    *schemaPerson
	Pet                       *schemaPet
	QueryFixture              *schemaQueryFixture
	QueryRelationFixture      *schemaQueryRelationFixture
	ResultSetFixture          *schemaResultSetFixture
	SchemaFixture             *schemaSchemaFixture
	SchemaRelationshipFixture *schemaSchemaRelationshipFixture
	StoreFixture              *schemaStoreFixture
	StoreWithConstructFixture *schemaStoreWithConstructFixture
	StoreWithNewFixture       *schemaStoreWithNewFixture
}

type schemaCar struct {
	*kallax.BaseSchema
	ID        kallax.SchemaField
	OwnerFK   kallax.SchemaField
	ModelName kallax.SchemaField
}

type schemaEventsAllFixture struct {
	*kallax.BaseSchema
	ID             kallax.SchemaField
	Checks         kallax.SchemaField
	MustFailBefore kallax.SchemaField
	MustFailAfter  kallax.SchemaField
}

type schemaEventsFixture struct {
	*kallax.BaseSchema
	ID             kallax.SchemaField
	Checks         kallax.SchemaField
	MustFailBefore kallax.SchemaField
	MustFailAfter  kallax.SchemaField
}

type schemaEventsSaveFixture struct {
	*kallax.BaseSchema
	ID             kallax.SchemaField
	Checks         kallax.SchemaField
	MustFailBefore kallax.SchemaField
	MustFailAfter  kallax.SchemaField
}

type schemaJSONModel struct {
	*kallax.BaseSchema
	ID       kallax.SchemaField
	Foo      kallax.SchemaField
	Bar      *schemaJSONModelBar
	BazSlice *schemaJSONModelBazSlice
	Baz      kallax.SchemaField
}

type schemaMultiKeySortFixture struct {
	*kallax.BaseSchema
	ID    kallax.SchemaField
	Name  kallax.SchemaField
	Start kallax.SchemaField
	End   kallax.SchemaField
}

type schemaNullable struct {
	*kallax.BaseSchema
	ID       kallax.SchemaField
	T        kallax.SchemaField
	SomeJSON *schemaNullableSomeJSON
	Scanner  kallax.SchemaField
}

type schemaPerson struct {
	*kallax.BaseSchema
	ID   kallax.SchemaField
	Name kallax.SchemaField
}

type schemaPet struct {
	*kallax.BaseSchema
	ID      kallax.SchemaField
	Name    kallax.SchemaField
	Kind    kallax.SchemaField
	OwnerFK kallax.SchemaField
}

type schemaQueryFixture struct {
	*kallax.BaseSchema
	ID                        kallax.SchemaField
	InverseFK                 kallax.SchemaField
	Embedded                  kallax.SchemaField
	Inline                    kallax.SchemaField
	MapOfString               kallax.SchemaField
	MapOfInterface            kallax.SchemaField
	MapOfSomeType             kallax.SchemaField
	Foo                       kallax.SchemaField
	StringProperty            kallax.SchemaField
	Integer                   kallax.SchemaField
	Integer64                 kallax.SchemaField
	Float32                   kallax.SchemaField
	Boolean                   kallax.SchemaField
	ArrayParam                kallax.SchemaField
	SliceParam                kallax.SchemaField
	AliasArrayParam           kallax.SchemaField
	AliasSliceParam           kallax.SchemaField
	AliasStringParam          kallax.SchemaField
	AliasIntParam             kallax.SchemaField
	DummyParam                kallax.SchemaField
	AliasDummyParam           kallax.SchemaField
	SliceDummyParam           kallax.SchemaField
	IDPropertyParam           kallax.SchemaField
	InterfacePropParam        kallax.SchemaField
	URLParam                  kallax.SchemaField
	TimeParam                 kallax.SchemaField
	AliasArrAliasStringParam  kallax.SchemaField
	AliasHereArrayParam       kallax.SchemaField
	ArrayAliasHereStringParam kallax.SchemaField
	ScannerValuerParam        kallax.SchemaField
}

type schemaQueryRelationFixture struct {
	*kallax.BaseSchema
	ID      kallax.SchemaField
	Name    kallax.SchemaField
	OwnerFK kallax.SchemaField
}

type schemaResultSetFixture struct {
	*kallax.BaseSchema
	ID  kallax.SchemaField
	Foo kallax.SchemaField
}

type schemaSchemaFixture struct {
	*kallax.BaseSchema
	ID             kallax.SchemaField
	String         kallax.SchemaField
	Int            kallax.SchemaField
	Inline         kallax.SchemaField
	MapOfString    kallax.SchemaField
	MapOfInterface kallax.SchemaField
	MapOfSomeType  kallax.SchemaField
	InverseFK      kallax.SchemaField
}

type schemaSchemaRelationshipFixture struct {
	*kallax.BaseSchema
	ID kallax.SchemaField
}

type schemaStoreFixture struct {
	*kallax.BaseSchema
	ID             kallax.SchemaField
	Foo            kallax.SchemaField
	SliceProp      kallax.SchemaField
	AliasSliceProp kallax.SchemaField
}

type schemaStoreWithConstructFixture struct {
	*kallax.BaseSchema
	ID  kallax.SchemaField
	Foo kallax.SchemaField
}

type schemaStoreWithNewFixture struct {
	*kallax.BaseSchema
	ID  kallax.SchemaField
	Foo kallax.SchemaField
	Bar kallax.SchemaField
}

type schemaJSONModelBar struct {
	*kallax.BaseSchemaField
	Qux *schemaJSONModelBarQux
	Mux kallax.SchemaField
}

type schemaJSONModelBarQux struct {
	*kallax.JSONSchemaArray
	Schnooga kallax.SchemaField
	Balooga  kallax.SchemaField
	Boo      kallax.SchemaField
}

func (s *schemaJSONModelBarQux) At(n int) *schemaJSONModelBarQux {
	return &schemaJSONModelBarQux{
		JSONSchemaArray: kallax.NewJSONSchemaArray("bar", "Qux"),
		Schnooga:        kallax.NewJSONSchemaKey(kallax.JSONText, "bar", "Qux", fmt.Sprint(n), "Schnooga"),
		Balooga:         kallax.NewJSONSchemaKey(kallax.JSONInt, "bar", "Qux", fmt.Sprint(n), "Balooga"),
		Boo:             kallax.NewJSONSchemaKey(kallax.JSONFloat, "bar", "Qux", fmt.Sprint(n), "Boo"),
	}
}

type schemaJSONModelBazSlice struct {
	*kallax.BaseSchemaField
	Mux kallax.SchemaField
}

func (s *schemaJSONModelBazSlice) At(n int) *schemaJSONModelBazSlice {
	return &schemaJSONModelBazSlice{
		BaseSchemaField: kallax.NewSchemaField("baz_slice").(*kallax.BaseSchemaField),
		Mux:             kallax.NewJSONSchemaKey(kallax.JSONText, "baz_slice", fmt.Sprint(n), "Mux"),
	}
}

type schemaNullableSomeJSON struct {
	*kallax.BaseSchemaField
	Foo kallax.SchemaField
}

var Schema = &schema{
	Car: &schemaCar{
		BaseSchema: kallax.NewBaseSchema(
			"cars",
			"__car",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{
				"Owner": kallax.NewForeignKey("owner_id", true),
			},
			func() kallax.Record {
				return new(Car)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("owner_id"),
			kallax.NewSchemaField("model_name"),
		),
		ID:        kallax.NewSchemaField("id"),
		OwnerFK:   kallax.NewSchemaField("owner_id"),
		ModelName: kallax.NewSchemaField("model_name"),
	},
	EventsAllFixture: &schemaEventsAllFixture{
		BaseSchema: kallax.NewBaseSchema(
			"event",
			"__eventsallfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(EventsAllFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("checks"),
			kallax.NewSchemaField("must_fail_before"),
			kallax.NewSchemaField("must_fail_after"),
		),
		ID:             kallax.NewSchemaField("id"),
		Checks:         kallax.NewSchemaField("checks"),
		MustFailBefore: kallax.NewSchemaField("must_fail_before"),
		MustFailAfter:  kallax.NewSchemaField("must_fail_after"),
	},
	EventsFixture: &schemaEventsFixture{
		BaseSchema: kallax.NewBaseSchema(
			"event",
			"__eventsfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(EventsFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("checks"),
			kallax.NewSchemaField("must_fail_before"),
			kallax.NewSchemaField("must_fail_after"),
		),
		ID:             kallax.NewSchemaField("id"),
		Checks:         kallax.NewSchemaField("checks"),
		MustFailBefore: kallax.NewSchemaField("must_fail_before"),
		MustFailAfter:  kallax.NewSchemaField("must_fail_after"),
	},
	EventsSaveFixture: &schemaEventsSaveFixture{
		BaseSchema: kallax.NewBaseSchema(
			"event",
			"__eventssavefixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(EventsSaveFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("checks"),
			kallax.NewSchemaField("must_fail_before"),
			kallax.NewSchemaField("must_fail_after"),
		),
		ID:             kallax.NewSchemaField("id"),
		Checks:         kallax.NewSchemaField("checks"),
		MustFailBefore: kallax.NewSchemaField("must_fail_before"),
		MustFailAfter:  kallax.NewSchemaField("must_fail_after"),
	},
	JSONModel: &schemaJSONModel{
		BaseSchema: kallax.NewBaseSchema(
			"jsons",
			"__jsonmodel",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(JSONModel)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("foo"),
			kallax.NewSchemaField("bar"),
			kallax.NewSchemaField("baz_slice"),
			kallax.NewSchemaField("baz"),
		),
		ID:  kallax.NewSchemaField("id"),
		Foo: kallax.NewSchemaField("foo"),
		Bar: &schemaJSONModelBar{
			BaseSchemaField: kallax.NewSchemaField("bar").(*kallax.BaseSchemaField),
			Qux: &schemaJSONModelBarQux{
				JSONSchemaArray: kallax.NewJSONSchemaArray("bar", "Qux"),
				Schnooga:        kallax.NewJSONSchemaKey(kallax.JSONText, "bar", "Qux", "Schnooga"),
				Balooga:         kallax.NewJSONSchemaKey(kallax.JSONInt, "bar", "Qux", "Balooga"),
				Boo:             kallax.NewJSONSchemaKey(kallax.JSONFloat, "bar", "Qux", "Boo"),
			},
			Mux: kallax.NewJSONSchemaKey(kallax.JSONText, "bar", "Mux"),
		},
		BazSlice: &schemaJSONModelBazSlice{
			BaseSchemaField: kallax.NewSchemaField("baz_slice").(*kallax.BaseSchemaField),
			Mux:             kallax.NewJSONSchemaKey(kallax.JSONText, "baz_slice", "Mux"),
		},
		Baz: kallax.NewSchemaField("baz"),
	},
	MultiKeySortFixture: &schemaMultiKeySortFixture{
		BaseSchema: kallax.NewBaseSchema(
			"query",
			"__multikeysortfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(MultiKeySortFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("name"),
			kallax.NewSchemaField("start"),
			kallax.NewSchemaField("_end"),
		),
		ID:    kallax.NewSchemaField("id"),
		Name:  kallax.NewSchemaField("name"),
		Start: kallax.NewSchemaField("start"),
		End:   kallax.NewSchemaField("_end"),
	},
	Nullable: &schemaNullable{
		BaseSchema: kallax.NewBaseSchema(
			"nullable",
			"__nullable",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(Nullable)
			},
			true,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("t"),
			kallax.NewSchemaField("some_json"),
			kallax.NewSchemaField("scanner"),
		),
		ID: kallax.NewSchemaField("id"),
		T:  kallax.NewSchemaField("t"),
		SomeJSON: &schemaNullableSomeJSON{
			BaseSchemaField: kallax.NewSchemaField("some_json").(*kallax.BaseSchemaField),
			Foo:             kallax.NewJSONSchemaKey(kallax.JSONInt, "some_json", "Foo"),
		},
		Scanner: kallax.NewSchemaField("scanner"),
	},
	Person: &schemaPerson{
		BaseSchema: kallax.NewBaseSchema(
			"persons",
			"__person",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{
				"Pets": kallax.NewForeignKey("owner_id", false),
				"Car":  kallax.NewForeignKey("owner_id", false),
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
			kallax.ForeignKeys{
				"Owner": kallax.NewForeignKey("owner_id", true),
			},
			func() kallax.Record {
				return new(Pet)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("name"),
			kallax.NewSchemaField("kind"),
			kallax.NewSchemaField("owner_id"),
		),
		ID:      kallax.NewSchemaField("id"),
		Name:    kallax.NewSchemaField("name"),
		Kind:    kallax.NewSchemaField("kind"),
		OwnerFK: kallax.NewSchemaField("owner_id"),
	},
	QueryFixture: &schemaQueryFixture{
		BaseSchema: kallax.NewBaseSchema(
			"query",
			"__queryfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{
				"Relation":  kallax.NewForeignKey("owner_id", false),
				"Inverse":   kallax.NewForeignKey("inverse_id", true),
				"NRelation": kallax.NewForeignKey("owner_id", false),
			},
			func() kallax.Record {
				return new(QueryFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("inverse_id"),
			kallax.NewSchemaField("embedded"),
			kallax.NewSchemaField("inline"),
			kallax.NewSchemaField("map_of_string"),
			kallax.NewSchemaField("map_of_interface"),
			kallax.NewSchemaField("map_of_some_type"),
			kallax.NewSchemaField("foo"),
			kallax.NewSchemaField("string_property"),
			kallax.NewSchemaField("integer"),
			kallax.NewSchemaField("integer64"),
			kallax.NewSchemaField("float32"),
			kallax.NewSchemaField("boolean"),
			kallax.NewSchemaField("array_param"),
			kallax.NewSchemaField("slice_param"),
			kallax.NewSchemaField("alias_array_param"),
			kallax.NewSchemaField("alias_slice_param"),
			kallax.NewSchemaField("alias_string_param"),
			kallax.NewSchemaField("alias_int_param"),
			kallax.NewSchemaField("dummy_param"),
			kallax.NewSchemaField("alias_dummy_param"),
			kallax.NewSchemaField("slice_dummy_param"),
			kallax.NewSchemaField("idproperty_param"),
			kallax.NewSchemaField("interface_prop_param"),
			kallax.NewSchemaField("urlparam"),
			kallax.NewSchemaField("time_param"),
			kallax.NewSchemaField("alias_arr_alias_string_param"),
			kallax.NewSchemaField("alias_here_array_param"),
			kallax.NewSchemaField("array_alias_here_string_param"),
			kallax.NewSchemaField("scanner_valuer_param"),
		),
		ID:                        kallax.NewSchemaField("id"),
		InverseFK:                 kallax.NewSchemaField("inverse_id"),
		Embedded:                  kallax.NewSchemaField("embedded"),
		Inline:                    kallax.NewSchemaField("inline"),
		MapOfString:               kallax.NewSchemaField("map_of_string"),
		MapOfInterface:            kallax.NewSchemaField("map_of_interface"),
		MapOfSomeType:             kallax.NewSchemaField("map_of_some_type"),
		Foo:                       kallax.NewSchemaField("foo"),
		StringProperty:            kallax.NewSchemaField("string_property"),
		Integer:                   kallax.NewSchemaField("integer"),
		Integer64:                 kallax.NewSchemaField("integer64"),
		Float32:                   kallax.NewSchemaField("float32"),
		Boolean:                   kallax.NewSchemaField("boolean"),
		ArrayParam:                kallax.NewSchemaField("array_param"),
		SliceParam:                kallax.NewSchemaField("slice_param"),
		AliasArrayParam:           kallax.NewSchemaField("alias_array_param"),
		AliasSliceParam:           kallax.NewSchemaField("alias_slice_param"),
		AliasStringParam:          kallax.NewSchemaField("alias_string_param"),
		AliasIntParam:             kallax.NewSchemaField("alias_int_param"),
		DummyParam:                kallax.NewSchemaField("dummy_param"),
		AliasDummyParam:           kallax.NewSchemaField("alias_dummy_param"),
		SliceDummyParam:           kallax.NewSchemaField("slice_dummy_param"),
		IDPropertyParam:           kallax.NewSchemaField("idproperty_param"),
		InterfacePropParam:        kallax.NewSchemaField("interface_prop_param"),
		URLParam:                  kallax.NewSchemaField("urlparam"),
		TimeParam:                 kallax.NewSchemaField("time_param"),
		AliasArrAliasStringParam:  kallax.NewSchemaField("alias_arr_alias_string_param"),
		AliasHereArrayParam:       kallax.NewSchemaField("alias_here_array_param"),
		ArrayAliasHereStringParam: kallax.NewSchemaField("array_alias_here_string_param"),
		ScannerValuerParam:        kallax.NewSchemaField("scanner_valuer_param"),
	},
	QueryRelationFixture: &schemaQueryRelationFixture{
		BaseSchema: kallax.NewBaseSchema(
			"query_relation",
			"__queryrelationfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{
				"Owner": kallax.NewForeignKey("owner_id", true),
			},
			func() kallax.Record {
				return new(QueryRelationFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("name"),
			kallax.NewSchemaField("owner_id"),
		),
		ID:      kallax.NewSchemaField("id"),
		Name:    kallax.NewSchemaField("name"),
		OwnerFK: kallax.NewSchemaField("owner_id"),
	},
	ResultSetFixture: &schemaResultSetFixture{
		BaseSchema: kallax.NewBaseSchema(
			"resultset",
			"__resultsetfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(ResultSetFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("foo"),
		),
		ID:  kallax.NewSchemaField("id"),
		Foo: kallax.NewSchemaField("foo"),
	},
	SchemaFixture: &schemaSchemaFixture{
		BaseSchema: kallax.NewBaseSchema(
			"schema",
			"__schemafixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{
				"Nested":  kallax.NewForeignKey("schema_fixture_id", false),
				"Inverse": kallax.NewForeignKey("rel_id", true),
			},
			func() kallax.Record {
				return new(SchemaFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("string"),
			kallax.NewSchemaField("int"),
			kallax.NewSchemaField("inline"),
			kallax.NewSchemaField("map_of_string"),
			kallax.NewSchemaField("map_of_interface"),
			kallax.NewSchemaField("map_of_some_type"),
			kallax.NewSchemaField("rel_id"),
		),
		ID:             kallax.NewSchemaField("id"),
		String:         kallax.NewSchemaField("string"),
		Int:            kallax.NewSchemaField("int"),
		Inline:         kallax.NewSchemaField("inline"),
		MapOfString:    kallax.NewSchemaField("map_of_string"),
		MapOfInterface: kallax.NewSchemaField("map_of_interface"),
		MapOfSomeType:  kallax.NewSchemaField("map_of_some_type"),
		InverseFK:      kallax.NewSchemaField("rel_id"),
	},
	SchemaRelationshipFixture: &schemaSchemaRelationshipFixture{
		BaseSchema: kallax.NewBaseSchema(
			"relationship",
			"__schemarelationshipfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(SchemaRelationshipFixture)
			},
			false,
			kallax.NewSchemaField("id"),
		),
		ID: kallax.NewSchemaField("id"),
	},
	StoreFixture: &schemaStoreFixture{
		BaseSchema: kallax.NewBaseSchema(
			"store",
			"__storefixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(StoreFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("foo"),
			kallax.NewSchemaField("slice_prop"),
			kallax.NewSchemaField("alias_slice_prop"),
		),
		ID:             kallax.NewSchemaField("id"),
		Foo:            kallax.NewSchemaField("foo"),
		SliceProp:      kallax.NewSchemaField("slice_prop"),
		AliasSliceProp: kallax.NewSchemaField("alias_slice_prop"),
	},
	StoreWithConstructFixture: &schemaStoreWithConstructFixture{
		BaseSchema: kallax.NewBaseSchema(
			"store_construct",
			"__storewithconstructfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(StoreWithConstructFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("foo"),
		),
		ID:  kallax.NewSchemaField("id"),
		Foo: kallax.NewSchemaField("foo"),
	},
	StoreWithNewFixture: &schemaStoreWithNewFixture{
		BaseSchema: kallax.NewBaseSchema(
			"store_new",
			"__storewithnewfixture",
			kallax.NewSchemaField("id"),
			kallax.ForeignKeys{},
			func() kallax.Record {
				return new(StoreWithNewFixture)
			},
			false,
			kallax.NewSchemaField("id"),
			kallax.NewSchemaField("foo"),
			kallax.NewSchemaField("bar"),
		),
		ID:  kallax.NewSchemaField("id"),
		Foo: kallax.NewSchemaField("foo"),
		Bar: kallax.NewSchemaField("bar"),
	},
}
