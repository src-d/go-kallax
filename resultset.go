package kallax

import (
	"database/sql"
	"fmt"
)

// ErrRawScan is an error returned when a the `Scan` method of `ResultSet`
// is called with a `ResultSet` created as a result of a `RawQuery`, which is
// not allowed.
var ErrRawScan = fmt.Errorf("result set comes from raw query, use RawScan instead")

// ErrStop is an error returned by a callback to stop the iteration in a
// Resulset.ForEach iteration
var ErrStop = fmt.Errorf("STOP")

// ResultSet is a generic collection of rows.
type ResultSet struct {
	columns  []string
	readOnly bool
	*sql.Rows
}

// NewResultSet creates a new result set with the given rows and columns.
// It is mandatory that all column names are in the same order and are exactly
// equal to the ones in the query that produced the rows.
func NewResultSet(rows *sql.Rows, readOnly bool, columns ...string) *ResultSet {
	return &ResultSet{columns, readOnly, rows}
}

// Scan fills the column fields of the given value with the current row.
func (rs *ResultSet) Scan(record Record) error {
	if len(rs.columns) == 0 {
		return ErrRawScan
	}

	var pointers = make([]interface{}, len(rs.columns))
	for i, col := range rs.columns {
		ptr, err := record.ColumnAddress(col)
		if err != nil {
			return err
		}
		pointers[i] = ptr
	}

	if err := rs.Rows.Scan(pointers...); err != nil {
		return err
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
