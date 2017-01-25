package kallax

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/lann/builder"
)

var (
	// ErrNonNewDocument non-new documents cannot be inserted
	ErrNonNewDocument = errors.New("kallax: cannot insert a non new document")
	// ErrNewDocument a new documents cannot be updated
	ErrNewDocument = errors.New("kallax: cannot updated a new document")
	// ErrEmptyID a document without ID cannot be used with Save method
	ErrEmptyID = errors.New("kallax: a record without id is not allowed")
	// ErrNoRowUpdate is returned when an update operation does not affect any
	// rows, meaning the model being updated does not exist.
	ErrNoRowUpdate = errors.New("kallax: update affected no rows")
	// ErrNotWritable is returned when a record is not writable.
	ErrNotWritable = errors.New("kallax: record is not writable")
	// ErrStop can be returned inside a ForEach callback to stop iteration.
	ErrStop = errors.New("kallax: stopped ForEach execution")
	// ErrTransactionInsideTransaction is returned when a transaction is run
	// inside a transaction.
	ErrTransactionInsideTransaction = errors.New("kallax: can't start a transaction inside a transaction")
	ErrInvalidTxCallback            = errors.New("kallax: invalid transaction callback given")
)

// Store is a structure capable of retrieving records from a concrete table in
// the database.
type Store struct {
	db       *sql.DB
	proxy    squirrel.DBProxy
	schema   Schema
	inserter squirrel.InsertBuilder
	updater  squirrel.UpdateBuilder
	deleter  squirrel.DeleteBuilder
}

// NewStore returns a new Store instance.
func NewStore(db *sql.DB, schema Schema) *Store {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	proxy := squirrel.NewStmtCacher(db)
	return &Store{
		db:       db,
		proxy:    proxy,
		schema:   schema,
		inserter: builder.Insert(schema.Table()).RunWith(proxy),
		updater:  builder.Update(schema.Table()).RunWith(proxy),
		deleter:  builder.Delete(schema.Table()).RunWith(proxy),
	}
}

func newStoreWithTransaction(tx *sql.Tx, schema Schema) *Store {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	proxy := squirrel.NewStmtCacher(tx)
	return &Store{
		proxy:    proxy,
		schema:   schema,
		inserter: builder.Insert(schema.Table()).RunWith(proxy),
		updater:  builder.Update(schema.Table()).RunWith(proxy),
		deleter:  builder.Delete(schema.Table()).RunWith(proxy),
	}
}

// Insert insert the given record in the table, returns error if no-new
// record is given. The record id is set if it's empty.
func (s *Store) Insert(record Record) error {
	if record.IsPersisted() {
		return ErrNonNewDocument
	}

	if record.GetID().IsEmpty() {
		record.SetID(NewID())
	}

	cols := ColumnNames(s.schema.Columns())
	values, err := RecordValues(record, cols...)
	if err != nil {
		return err
	}

	_, err = s.inserter.
		Columns(cols...).
		Values(values...).
		Exec()
	if err != nil {
		return err
	}

	record.setWritable(true)
	record.setPersisted()
	return nil
}

// Update updates the given fields of a record in the table. All fields are
// updated if no fields are provided. For an update to take place, the record is
// required to have a non-empty ID and not to be a new record.
// Returns the number of updated rows and an error, if any.
func (s *Store) Update(record Record, cols ...SchemaField) (int64, error) {
	if !record.IsWritable() {
		return 0, ErrNotWritable
	}

	if !record.IsPersisted() {
		return 0, ErrNewDocument
	}

	if record.GetID().IsEmpty() {
		return 0, ErrEmptyID
	}

	if len(cols) == 0 {
		cols = s.schema.Columns()
	}

	columnNames := ColumnNames(cols)
	values, err := RecordValues(record, columnNames...)
	if err != nil {
		return 0, err
	}

	var clauses = make(map[string]interface{}, len(cols))
	for i, col := range columnNames {
		clauses[col] = values[i]
	}

	result, err := s.updater.
		SetMap(clauses).
		Where(squirrel.Eq{
			s.schema.ID().String(): record.GetID(),
		}).
		Exec()
	if err != nil {
		return 0, err
	}

	cnt, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if cnt == 0 {
		return 0, ErrNoRowUpdate
	}

	return cnt, nil
}

// Save inserts or updates the given record in the table. It requires a record
// with non empty ID.
func (s *Store) Save(record Record) (updated bool, err error) {
	if record.GetID().IsEmpty() {
		return false, ErrEmptyID
	}

	if !record.IsPersisted() {
		return false, s.Insert(record)
	}

	rowsUpdated, err := s.Update(record)
	if err != nil {
		return false, err
	}

	return rowsUpdated > 0, nil
}

// Delete removes the record from the table. A non-new record with non-empty
// ID is required.
func (s *Store) Delete(record Record) error {
	if record.GetID().IsEmpty() {
		return ErrEmptyID
	}

	_, err := s.deleter.
		Where(squirrel.Eq{
			s.schema.ID().String(): record.GetID(),
		}).
		Exec()
	return err
}

// RawQuery performs a raw SQL query with the given parameters and returns a
// result set with the results.
// WARNING: A result set created from a raw query can only be scanned using the
// RawScan method of ResultSet, instead of Scan.
func (s *Store) RawQuery(sql string, params ...interface{}) (ResultSet, error) {
	rows, err := s.db.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	return NewResultSet(rows, true, nil), nil
}

// RawExec executes a raw SQL query with the given parameters and returns
// the number of affected rows.
func (s *Store) RawExec(sql string, params ...interface{}) (int64, error) {
	result, err := s.db.Exec(sql, params...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find performs a query and returns a result set with the results.
func (s *Store) Find(q Query) (ResultSet, error) {
	rels := q.getRelationships()
	if containsRelationshipOfType(rels, ManyToMany) {
		return nil, fmt.Errorf("kallax: many to many relationships are not supported")
	}

	if containsRelationshipOfType(rels, OneToMany) {
		return NewBatchingResultSet(s.schema, s.proxy, q), nil
	}

	columns, builder := q.compile()
	if offset := q.GetOffset(); offset > 0 {
		builder = builder.Offset(offset)
	}

	if limit := q.GetLimit(); limit > 0 {
		builder = builder.Limit(limit)
	}

	rows, err := builder.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}

	return NewResultSet(
		rows,
		q.isReadOnly(),
		q.getRelationships(),
		columns...,
	), nil
}

// MustFind performs a query and returns a result set with the results.
// It panics if the query fails.
func (s *Store) MustFind(q Query) ResultSet {
	rs, err := s.Find(q)
	if err != nil {
		panic(err)
	}
	return rs
}

// Reload refreshes the record with the data in the database and makes the
// record writable.
func (s *Store) Reload(record Record) error {
	if record.GetID().IsEmpty() {
		return ErrEmptyID
	}

	q := NewBaseQuery(s.schema)
	q.Where(Eq(s.schema.ID(), record.GetID()))
	q.Limit(1)
	columns, builder := q.compile()

	rows, err := builder.RunWith(s.proxy).Query()
	if err != nil {
		return err
	}

	rs := NewResultSet(rows, false, nil, columns...)
	if !rs.Next() {
		return sql.ErrNoRows
	}

	return rs.Scan(record)
}

// Count returns the number of rows selected by the given query.
func (s *Store) Count(q Query) (count int64, err error) {
	_, queryBuilder := q.compile()
	builder := builder.Set(queryBuilder, "Columns", nil).(squirrel.SelectBuilder)
	err = builder.Column(fmt.Sprintf("COUNT(%s)", s.schema.ID())).
		RunWith(s.proxy).
		QueryRow().
		Scan(&count)
	return
}

// MustCount returns the number of rows selected by the given query. It panics
// if the query fails.
func (s *Store) MustCount(q Query) int64 {
	cnt, err := s.Count(q)
	if err != nil {
		panic(err)
	}

	return cnt
}

// Transaction executes the given callback in a transaction and rollbacks if
// an error is returned.
// The transaction is only open in the store passed as a parameter to the
// callback.
func (s *Store) Transaction(callback func(*Store) error) error {
	if s.db == nil {
		return ErrTransactionInsideTransaction
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("kallax: can't open transaction: %s", err)
	}

	if err := callback(newStoreWithTransaction(tx, s.schema)); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("kallax: unable to rollback transaction: %s", err)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("kallax: unable to commit transaction: %s", err)
	}

	return nil
}
