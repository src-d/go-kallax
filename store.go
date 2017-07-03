package kallax

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

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
	// ErrInvalidTxCallback is returned when a nil callback is passed.
	ErrInvalidTxCallback = errors.New("kallax: invalid transaction callback given")
	// ErrNotFound is returned when a certain entity is not found.
	ErrNotFound = errors.New("kallax: entity not found")
	// ErrCantSetID is returned when a model is inserted and it does not have
	// neither an autoincrement primary key nor implements the IDSetter
	// interface.
	ErrCantSetID = errors.New("kallax: model does not have an auto incrementable primary key, it needs to implement IDSetter interface")
	// ErrNoColumns is an error returned when the user tries to insert a model
	// with no other columns than the autoincrementable primary key.
	ErrNoColumns = errors.New("kallax: your model does not have any column besides its autoincrementable primary key and cannot be inserted")
)

// GenericStorer is a type that contains a generic store and has methods to
// retrieve it and set it.
type GenericStorer interface {
	// GenericStore returns the generic store in this type.
	GenericStore() *Store
	// SetGenericStore sets the generic store for this type.
	SetGenericStore(*Store)
}

// StoreFrom sets the generic store of `from` in `to`.
func StoreFrom(to, from GenericStorer) {
	if to == nil || from == nil {
		return
	}

	to.SetGenericStore(from.GenericStore())
}

// LoggerFunc is a function that takes a log message with some arguments and
// logs it.
type LoggerFunc func(string, ...interface{})

// debugProxy is a database proxy that logs all SQL statements executed.
type debugProxy struct {
	logger LoggerFunc
	proxy  squirrel.DBProxy
}

func defaultLogger(message string, args ...interface{}) {
	log.Printf("%s, args: %v", message, args)
}

func (p *debugProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	p.logger(fmt.Sprintf("kallax: Exec: %s", query), args...)
	return p.proxy.Exec(query, args...)
}

func (p *debugProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	p.logger(fmt.Sprintf("kallax: Query: %s", query), args...)
	return p.proxy.Query(query, args...)
}

func (p *debugProxy) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	p.logger(fmt.Sprintf("kallax: QueryRow: %s", query), args...)
	return p.proxy.QueryRow(query, args...)
}

func (p *debugProxy) Prepare(query string) (*sql.Stmt, error) {
	p.logger(fmt.Sprintf("kallax: Prepare: %s", query))
	return p.proxy.Prepare(query)
}

// Store is a structure capable of retrieving records from a concrete table in
// the database.
type Store struct {
	builder squirrel.StatementBuilderType
	db      *sql.DB
	proxy   squirrel.DBProxy
}

// NewStore returns a new Store instance.
func NewStore(db *sql.DB) *Store {
	proxy := squirrel.NewStmtCacher(db)
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(proxy)
	return &Store{
		db:      db,
		proxy:   proxy,
		builder: builder,
	}
}

func newStoreWithTransaction(tx *sql.Tx) *Store {
	proxy := squirrel.NewStmtCacher(tx)
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(proxy)
	return &Store{
		proxy:   proxy,
		builder: builder,
	}
}

// Debug returns a new store that will print all SQL statements to stdout using
// the log.Printf function.
func (s *Store) Debug() *Store {
	return s.DebugWith(defaultLogger)
}

// DebugWith returns a new store that will print all SQL statements using the
// given logger function.
func (s *Store) DebugWith(logger LoggerFunc) *Store {
	proxy := &debugProxy{logger, s.proxy}
	return &Store{
		builder: s.builder.RunWith(proxy),
		db:      s.db,
		proxy:   proxy,
	}
}

// Insert insert the given record in the table, returns error if no-new
// record is given. The record id is set if it's empty.
func (s *Store) Insert(schema Schema, record Record) error {
	if record.IsPersisted() {
		return ErrNonNewDocument
	}

	cols := ColumnNames(schema.Columns())
	if schema.isPrimaryKeyAutoIncrementable() {
		// we have to remove the pk from the list, in case the
		// pk is auto incremented if it's 0
		// ID is always the first field, so it's safe to slice here
		cols = cols[1:]
	}

	if len(cols) == 0 {
		return ErrNoColumns
	}

	values, err := RecordValues(record, cols...)
	if err != nil {
		return err
	}

	virtualCols, virtualColValues := virtualColumns(record, cols)
	cols = append(cols, virtualCols...)
	values = append(values, virtualColValues...)

	builder := s.builder.
		Insert(schema.Table()).
		Columns(cols...).
		Values(values...)
	if schema.isPrimaryKeyAutoIncrementable() {
		var pk interface{}
		pk, err = record.ColumnAddress(schema.ID().String())
		if err != nil {
			return err
		}

		err = builder.
			Suffix(fmt.Sprintf("RETURNING %q", schema.ID())).
			QueryRow().
			Scan(pk)
	} else {
		_, err = builder.Exec()
	}

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
func (s *Store) Update(schema Schema, record Record, cols ...SchemaField) (int64, error) {
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
		cols = schema.Columns()
	}

	columnNames := ColumnNames(cols)
	values, err := RecordValues(record, columnNames...)
	if err != nil {
		return 0, err
	}

	virtualCols, virtualColValues := virtualColumns(record, columnNames)
	columnNames = append(columnNames, virtualCols...)
	values = append(values, virtualColValues...)

	var clauses = make(map[string]interface{}, len(cols))
	for i, col := range columnNames {
		clauses[col] = values[i]
	}

	result, err := s.builder.
		Update(schema.Table()).
		SetMap(clauses).
		Where(squirrel.Eq{
			schema.ID().String(): record.GetID(),
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

// Save inserts or updates the given record in the table.
func (s *Store) Save(schema Schema, record Record) (updated bool, err error) {
	if !record.IsPersisted() {
		return false, s.Insert(schema, record)
	}

	rowsUpdated, err := s.Update(schema, record)
	if err != nil {
		return false, err
	}

	return rowsUpdated > 0, nil
}

// Delete removes the record from the table. A non-new record with non-empty
// ID is required.
func (s *Store) Delete(schema Schema, record Record) error {
	if record.GetID().IsEmpty() {
		return ErrEmptyID
	}

	_, err := s.builder.
		Delete(schema.Table()).
		Where(squirrel.Eq{
			schema.ID().String(): record.GetID(),
		}).
		Exec()
	return err
}

// RawQuery performs a raw SQL query with the given parameters and returns a
// result set with the results.
// WARNING: A result set created from a raw query can only be scanned using the
// RawScan method of ResultSet, instead of Scan.
func (s *Store) RawQuery(sql string, params ...interface{}) (ResultSet, error) {
	rows, err := s.proxy.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	return NewResultSet(rows, true, nil), nil
}

// RawExec executes a raw SQL query with the given parameters and returns
// the number of affected rows.
func (s *Store) RawExec(sql string, params ...interface{}) (int64, error) {
	result, err := s.proxy.Exec(sql, params...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find performs a query and returns a result set with the results.
func (s *Store) Find(q Query) (ResultSet, error) {
	rels := q.getRelationships()
	if containsRelationshipOfType(rels, OneToMany) ||
		containsRelationshipOfType(rels, Through) {
		return NewBatchingResultSet(newBatchQueryRunner(q.Schema(), s.proxy, q)), nil
	}

	columns, builder := q.compile()
	if offset := q.GetOffset(); offset > 0 {
		builder = builder.Offset(offset)
	}

	if limit := q.GetLimit(); limit > 0 {
		builder = builder.Limit(limit)
	}

	rows, err := builder.RunWith(s.proxy).Query()
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
func (s *Store) Reload(schema Schema, record Record) error {
	if record.GetID().IsEmpty() {
		return ErrEmptyID
	}

	q := NewBaseQuery(schema)
	q.Where(Eq(schema.ID(), record.GetID()))
	q.Limit(1)
	columns, builder := q.compile()

	rows, err := builder.RunWith(s.proxy).Query()
	if err != nil {
		return err
	}

	rs := NewResultSet(rows, false, nil, columns...)
	if !rs.Next() {
		return ErrNotFound
	}

	return rs.Scan(record)
}

// Count returns the number of rows selected by the given query.
func (s *Store) Count(q Query) (count int64, err error) {
	_, queryBuilder := q.compile()
	builder := builder.Set(queryBuilder, "Columns", nil).(squirrel.SelectBuilder)
	err = builder.Column(fmt.Sprintf("COUNT(%s)", q.Schema().ID())).
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
// If a transaction is already opened in this store, instead of opening a new
// one, the other will be reused.
func (s *Store) Transaction(callback func(*Store) error) error {
	if s.db == nil {
		return callback(s)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("kallax: can't open transaction: %s", err)
	}

	if err := callback(newStoreWithTransaction(tx)); err != nil {
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

// RecordWithSchema is a structure that contains both a record and its schema.
// Only for internal purposes.
type RecordWithSchema struct {
	Schema Schema
	Record Record
}

func virtualColumns(r Record, columns []string) (cols []string, vals []interface{}) {
	c, ok := r.(VirtualColumnContainer)
	if !ok {
		return
	}

	vcols := c.getVirtualColumns()
	for col, val := range vcols {
		if !containsString(columns, col) {
			cols = append(cols, col)
			vals = append(vals, val)
		}
	}

	return
}

func containsString(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}
