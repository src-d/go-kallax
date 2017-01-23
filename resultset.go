package kallax

import (
	"database/sql"
	"fmt"
)

// ErrRawScan is an error returned when a the `Scan` method of `ResultSet`
// is called with a `ResultSet` created as a result of a `RawQuery`, which is
// not allowed.
var ErrRawScan = fmt.Errorf("result set comes from raw query, use RawScan instead")

// ResultSet is a generic collection of rows.
type ResultSet struct {
	relationships []Relationship
	columns       []string
	readOnly      bool
	*sql.Rows
}

// NewResultSet creates a new result set with the given rows and columns.
// It is mandatory that all column names are in the same order and are exactly
// equal to the ones in the query that produced the rows.
func NewResultSet(rows *sql.Rows, readOnly bool, relationships []Relationship, columns ...string) *ResultSet {
	return &ResultSet{
		relationships,
		columns,
		readOnly,
		rows,
	}
}

// Scan fills the column fields of the given value with the current row.
func (rs *ResultSet) Scan(record Record) error {
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
func (rs *ResultSet) RawScan(dest ...interface{}) error {
	return rs.Rows.Scan(dest...)
}
