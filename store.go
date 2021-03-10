package kallax

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

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

func defaultLogger(message string, args ...interface{}) {
	log.Printf("%s, args: %v", message, args)
}

// runnerLogger is a database runner that logs all SQL statements executed.
type proxyLogger struct {
	squirrel.DBProxyContext
	logger LoggerFunc
}

func (p *proxyLogger) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := p.DBProxyContext.Exec(query, args...)
	p.logger(fmt.Sprintf("kallax: Exec: (%v) %s", time.Since(start), query), args...)
	return result, err
}

func (p *proxyLogger) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := p.DBProxyContext.Query(query, args...)
	p.logger(fmt.Sprintf("kallax: Query: (%v) %s", time.Since(start), query), args...)
	return rows, err
}

func (p *proxyLogger) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	start := time.Now()
	rowScanner := p.DBProxyContext.QueryRow(query, args...)
	p.logger(fmt.Sprintf("kallax: QueryRow: (%v) %s", time.Since(start), query), args...)
	return rowScanner
}

func (p *proxyLogger) Prepare(query string) (*sql.Stmt, error) {
	//If chained runner is a proxy, run Prepare(). Otherwise, noop.
	start := time.Now()
	statement, err := p.DBProxyContext.Prepare(query)
	p.logger(fmt.Sprintf("kallax: Prepare: (%v) %s", time.Since(start), query))
	return statement, err
}

// PrepareContext will not be logged

// dbRunner is a copypaste from squirrel.dbRunner, used to make sql.DB implement squirrel.QueryRower.
// squirrel will silently fail and return nil if BaseRunner(s) supplied to RunWith don't implement QueryRower, so
// it has been copied there to avoid that.
// TODO: Delete this when squirrel dependency is dropped.
type dbRunner struct {
	*sql.DB
}

func (r *dbRunner) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return r.DB.QueryRow(query, args...)
}

// txRunner does the analogous for sql.Tx
type txRunner struct {
	*sql.Tx
}

func (r *txRunner) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return r.Tx.QueryRow(query, args...)
}

// Store is a structure capable of retrieving records from a concrete table in
// the database.
type Store struct {
	db        squirrel.DBProxyContext
	runner    squirrel.DBProxyContext
	logger    LoggerFunc
}

// NewStore returns a new Store instance.
func NewStore(db *sql.DB) *Store {
	return (&Store{
		db:        &dbRunner{db},
	}).init()
}

// init initializes the store runner with debugging and returns itself for chainability
func (s *Store) init() *Store {
	s.runner = s.db

	if s.logger != nil {
		s.runner = &proxyLogger{logger: s.logger, DBProxyContext: s.runner}
	}

	return s
}

// Debug returns a new store that will print all SQL statements to stdout using
// the log.Printf function.
func (s *Store) Debug() *Store {
	return s.DebugWith(defaultLogger)
}

// DebugWith returns a new store that will print all SQL statements using the
// given logger function.
func (s *Store) DebugWith(logger LoggerFunc) *Store {
	return (&Store{
		db:        s.db,
		logger:    logger,
	}).init()
}

// Runner gives access to the current runner (*sql.DB or *sql.TX when in a transaction)
func (s *Store) Runner() squirrel.DBProxyContext {
	return s.runner
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

	values, cols, err := RecordValues(record, cols...)
	if err != nil {
		return err
	}

	virtualCols, virtualColValues := virtualColumns(record, cols)
	cols = append(cols, virtualCols...)
	values = append(values, virtualColValues...)

	var colBuf bytes.Buffer
	var valBuf bytes.Buffer

	for i, col := range cols {
		if i != 0 {
			colBuf.WriteRune(',')
			valBuf.WriteRune(',')
		}
		colBuf.WriteString(col)
		valBuf.WriteString(fmt.Sprintf("$%d", i+1))
	}

	var query bytes.Buffer
	query.WriteString("INSERT INTO ")
	query.WriteString(schema.Table())
	query.WriteString(" (")
	query.WriteString(colBuf.String())
	query.WriteString(") VALUES (")
	query.WriteString(valBuf.String())
	query.WriteString(")")

	if schema.isPrimaryKeyAutoIncrementable() {
		var pk interface{}
		pk, err = record.ColumnAddress(schema.ID().String())
		if err != nil {
			return err
		}

		query.WriteString(fmt.Sprintf(" RETURNING %s", schema.ID().String()))
		//err = s.runner.QueryRow(query.String(), values...).Scan(pk)
		rows, err := s.runner.Query(query.String(), values...)
		if err != nil {
			return err
		}
		if rows.Next() {
			err = rows.Scan(pk)
			rows.Close()
			if err != nil {
				return err
			}
		}
	} else {
		_, err = s.runner.Exec(query.String(), values...)
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

	// remove the ID from there
	columnNames := ColumnNames(cols)
	values, columnNames, err := RecordValues(record, columnNames...)
	if err != nil {
		return 0, err
	}

	virtualCols, virtualColValues := virtualColumns(record, columnNames)
	columnNames = append(columnNames, virtualCols...)
	values = append(values, virtualColValues...)

	var query bytes.Buffer
	query.WriteString("UPDATE ")
	query.WriteString(schema.Table())
	query.WriteString(" SET ")
	for i, col := range columnNames {
		if i != 0 {
			query.WriteRune(',')
		}
		query.WriteString(col)
		query.WriteRune('=')
		query.WriteString(fmt.Sprintf("$%d", i+1))
	}
	query.WriteString(" WHERE ")
	query.WriteString(schema.ID().String())
	query.WriteRune('=')
	query.WriteString(fmt.Sprintf("$%d", len(columnNames)+1))

	result, err := s.runner.Exec(query.String(), append(values, record.GetID())...)
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

	var query bytes.Buffer
	query.WriteString("DELETE FROM ")
	query.WriteString(schema.Table())
	query.WriteString(" WHERE ")
	query.WriteString(schema.ID().String())
	query.WriteString("=$1")

	_, err := s.runner.Exec(query.String(), record.GetID())
	return err
}

// RawQuery performs a raw SQL query with the given parameters and returns a
// result set with the results.
// WARNING: A result set created from a raw query can only be scanned using the
// RawScan method of ResultSet, instead of Scan.
func (s *Store) RawQuery(sql string, params ...interface{}) (ResultSet, error) {
	rows, err := s.runner.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	return NewResultSet(rows, true, nil), nil
}

// RawExec executes a raw SQL query with the given parameters and returns
// the number of affected rows.
func (s *Store) RawExec(sql string, params ...interface{}) (int64, error) {
	result, err := s.runner.Exec(sql, params...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find performs a query and returns a result set with the results.
func (s *Store) Find(q Query) (ResultSet, error) {
	rels := q.getRelationships()
	if containsRelationshipOfType(rels, OneToMany) {
		return NewBatchingResultSet(newBatchQueryRunner(q.Schema(), s.runner, q)), nil
	}

	columns, builder := q.compile()
	if offset := q.GetOffset(); offset > 0 {
		builder = builder.Offset(offset)
	}

	if limit := q.GetLimit(); limit > 0 {
		builder = builder.Limit(limit)
	}

	rows, err := builder.RunWith(s.runner).Query()
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

	rows, err := builder.RunWith(s.runner).Query()
	if err != nil {
		return err
	}
	defer rows.Close()

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
	err = builder.Column("COUNT(*)").
		RunWith(s.runner).
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
	var tx *sql.Tx
	var err error
	if db, ok := s.db.(*dbRunner); ok {
		// db is *sql.DB, not *sql.Tx
		tx, err = db.Begin()
		if err != nil {
			return fmt.Errorf("kallax: can't open transaction: %s", err)
		}
	} else {
		// store is already holding a transaction
		return callback(s)
	}

	txStore := (&Store{
		db:        &txRunner{tx},
		logger:    s.logger,
	}).init()

	if err := callback(txStore); err != nil {
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
