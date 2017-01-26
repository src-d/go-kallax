package kallax

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
)

// ResultSet is the common interface all result sets need to implement.
type ResultSet interface {
	// RawScan allows for raw scanning of fields in a result set.
	RawScan(...interface{}) error
	// Next moves the pointer to the next item in the result set and returns
	// if there was any.
	Next() bool
	// Get returns the next record of the given schema.
	Get(Schema) (Record, error)
	io.Closer
}

// ErrRawScan is an error returned when a the `Scan` method of `ResultSet`
// is called with a `ResultSet` created as a result of a `RawQuery`, which is
// not allowed.
var ErrRawScan = errors.New("kallax: result set comes from raw query, use RawScan instead")

// BaseResultSet is a generic collection of rows.
type BaseResultSet struct {
	relationships []Relationship
	columns       []string
	readOnly      bool
	*sql.Rows
}

// NewResultSet creates a new result set with the given rows and columns.
// It is mandatory that all column names are in the same order and are exactly
// equal to the ones in the query that produced the rows.
func NewResultSet(rows *sql.Rows, readOnly bool, relationships []Relationship, columns ...string) *BaseResultSet {
	return &BaseResultSet{
		relationships,
		columns,
		readOnly,
		rows,
	}
}

// Get returns the next record in the schema.
func (rs *BaseResultSet) Get(schema Schema) (Record, error) {
	record := schema.New()
	if err := rs.Scan(record); err != nil {
		return nil, err
	}
	return record, nil
}

// Scan fills the column fields of the given value with the current row.
func (rs *BaseResultSet) Scan(record Record) error {
	if len(rs.columns) == 0 {
		return ErrRawScan
	}

	var (
		relationships = make([]Record, len(rs.relationships))
		pointers      = make([]interface{}, len(rs.columns))
	)

	for i, col := range rs.columns {
		ptr, err := record.ColumnAddress(col)
		if err != nil {
			return err
		}
		pointers[i] = ptr
	}

	for i, r := range rs.relationships {
		rec, err := record.NewRelationshipRecord(r.Field)
		if err != nil {
			return err
		}

		for _, col := range r.Schema.Columns() {
			ptr, err := rec.ColumnAddress(col.String())
			if err != nil {
				return err
			}
			pointers = append(pointers, ptr)
		}

		relationships[i] = rec
	}

	if err := rs.Rows.Scan(pointers...); err != nil {
		return err
	}

	for i, r := range rs.relationships {
		err := record.SetRelationship(r.Field, relationships[i])
		if err != nil {
			return err
		}
	}

	record.setWritable(!rs.readOnly)
	record.setPersisted()
	return nil
}

// RowScan copies the columns in the current row into the values pointed at by
// dest. The number of values in dest must be the same as the number of columns
// selected in the query.
func (rs *BaseResultSet) RawScan(dest ...interface{}) error {
	return rs.Rows.Scan(dest...)
}

// NewBatchingResultSet returns a new result set that performs batching
// underneath.
func NewBatchingResultSet(runner *batchQueryRunner) *BatchingResultSet {
	return &BatchingResultSet{runner: runner}
}

// BatchingResultSet is a result set that retrieves all the items up to the
// batch size set in the query.
// If there are 1:N relationships, it collects all the identifiers of
// those records, retrieves all the rows matching them in the table of the
// the N end, and assigns them to their correspondent to the record they belong
// to.
// It will continue doing this process until no more rows are returned by the
// query.
// This minimizes the number of queries and operations to perform in order to
// retrieve a set of results and their relationships.
type BatchingResultSet struct {
	runner  *batchQueryRunner
	last    Record
	lastErr error
}

// Next advances the internal index of the fetched records in one.
// If there are no fetched records, will fetch the next batch.
// It will return false when there are no more rows.
func (rs *BatchingResultSet) Next() bool {
	rs.last, rs.lastErr = rs.runner.next()
	if rs.lastErr == ErrNoMoreRows {
		return false
	}

	return true
}

// Get returns the next processed record and the last error occurred.
// Even though it accepts a schema, it is ignored, as the result set is
// already aware of it. This is here just to be able to imeplement the
// ResultSet interface.
func (rs *BatchingResultSet) Get(_ Schema) (Record, error) {
	return rs.last, rs.lastErr
}

// Close will do nothing, as the internal result sets used by this are closed
// when the rows at fetched. It will never throw an error.
func (rs *BatchingResultSet) Close() error {
	return nil
}

// RawScan will always throw an error, as this is not a supported operation of
// a batching result set.
func (rs *BatchingResultSet) RawScan(_ ...interface{}) error {
	return fmt.Errorf("kallax: cannot perform a raw scan on a batching result set")
}
