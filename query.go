package kallax

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// Query returns information about some query settings and compiles the query.
type Query interface {
	compile() ([]string, squirrel.SelectBuilder)
	getRelationships() []Relationship
	isReadOnly() bool
	// GetOffset returns the number of skipped rows in the query.
	GetOffset() uint64
	// GetLimit returns the max number of rows retrieved by the query.
	GetLimit() uint64
	// GetBatchSize returns the number of rows retrieved by the store per
	// batch. This is only used and has effect on queries with 1:N
	// relationships.
	GetBatchSize() uint64
}

type columnSet []SchemaField

func (cs columnSet) contains(col SchemaField) bool {
	for _, c := range cs {
		if c.String() == col.String() {
			return true
		}
	}
	return false
}

func (cs *columnSet) add(cols ...SchemaField) {
	for _, col := range cols {
		cs.addCol(col)
	}
}

func (cs *columnSet) addCol(col SchemaField) {
	if !cs.contains(col) {
		*cs = append(*cs, col)
	}
}

func (cs *columnSet) remove(cols ...SchemaField) {
	var newSet = make(columnSet, 0, len(*cs))
	toRemove := columnSet(cols)
	for _, col := range *cs {
		if !toRemove.contains(col) {
			newSet = append(newSet, col)
		}
	}
	*cs = newSet
}

func (cs columnSet) copy() []SchemaField {
	var result = make(columnSet, len(cs))
	for i, col := range cs {
		result[i] = col
	}
	return result
}

// BaseQuery is a generic query builder.
type BaseQuery struct {
	schema          Schema
	columns         columnSet
	excludedColumns columnSet
	// relationColumns contains the qualified names of the columns selected by the 1:1
	// relationships
	relationColumns []string
	relationships   []Relationship
	builder         squirrel.SelectBuilder

	selectChanged bool
	batchSize     uint64
	offset        uint64
	limit         uint64
}

// NewBaseQuery creates a new BaseQuery for querying the given table
// and the given selected columns.
func NewBaseQuery(schema Schema) *BaseQuery {
	return &BaseQuery{
		builder: squirrel.StatementBuilder.
			PlaceholderFormat(squirrel.Dollar).
			Select().
			From(schema.Table() + " " + schema.Alias()),
		columns:   columnSet(schema.Columns()),
		batchSize: 50,
		schema:    schema,
	}
}

func (q *BaseQuery) isReadOnly() bool {
	return q.selectChanged
}

// Select adds the given columns to the list of selected columns in the query.
func (q *BaseQuery) Select(columns ...SchemaField) {
	if !q.selectChanged {
		q.columns = columnSet{}
		q.selectChanged = true
	}

	q.excludedColumns.remove(columns...)
	q.columns.add(columns...)
}

// SelectNot adds the given columns to the list of excluded columns in the query.
func (q *BaseQuery) SelectNot(columns ...SchemaField) {
	q.excludedColumns.add(columns...)
}

// Copy returns an identical copy of the query. BaseQuery is mutable, that is
// why this method is provided.
func (q *BaseQuery) Copy() *BaseQuery {
	return &BaseQuery{
		builder:         q.builder,
		columns:         q.columns.copy(),
		excludedColumns: q.excludedColumns.copy(),
		relationColumns: q.relationColumns[:],
		relationships:   q.relationships[:],
		selectChanged:   q.selectChanged,
		batchSize:       q.GetBatchSize(),
		limit:           q.GetLimit(),
		offset:          q.GetOffset(),
		schema:          q.schema,
	}
}

func (q *BaseQuery) getRelationships() []Relationship {
	return q.relationships
}

func (q *BaseQuery) selectedColumns() []SchemaField {
	var result = make([]SchemaField, 0, len(q.columns))
	for _, col := range q.columns {
		if !q.excludedColumns.contains(col) {
			result = append(result, col)
		}
	}
	return result
}

// AddRelation adds a relationship if the given to the query, which is present
// in the given field of the query base schema. A condition to filter can also
// be passed in the case of one to many relationships.
func (q *BaseQuery) AddRelation(schema Schema, field string, typ RelationshipType, filter Condition) error {
	if typ == ManyToMany {
		return fmt.Errorf("kallax: many to many relationship are not supported, field: %s", field)
	}

	fk, ok := q.schema.ForeignKey(field)
	if !ok {
		return fmt.Errorf(
			"kallax: cannot find foreign key to join tables %s and %s",
			q.schema.Table(), schema.Table(),
		)
	}
	schema = schema.WithAlias(field)

	if typ == OneToOne {
		q.join(schema, fk)
	}

	q.relationships = append(q.relationships, Relationship{typ, field, schema, filter})
	return nil
}

func (q *BaseQuery) join(schema Schema, fk SchemaField) {
	q.builder = q.builder.LeftJoin(fmt.Sprintf(
		"%s %s ON (%s = %s)",
		schema.Table(),
		schema.Alias(),
		fk.QualifiedName(schema),
		q.schema.ID().QualifiedName(q.schema),
	))

	for _, col := range schema.Columns() {
		q.relationColumns = append(
			q.relationColumns,
			col.QualifiedName(schema),
		)
	}
}

// Order adds the given order clauses to the list of columns to order the
// results by.
func (q *BaseQuery) Order(cols ...ColumnOrder) {
	var c = make([]string, len(cols))
	for i, v := range cols {
		c[i] = v.ToSql(q.schema)
	}
	q.builder = q.builder.OrderBy(c...)
}

// BatchSize sets the batch size.
func (q *BaseQuery) BatchSize(size uint64) {
	q.batchSize = size
}

// GetBatchSize returns the number of rows retrieved per batch while retrieving
// 1:N relationships.
func (q *BaseQuery) GetBatchSize() uint64 {
	return q.batchSize
}

// Limit sets the max number of rows to retrieve.
func (q *BaseQuery) Limit(n uint64) {
	q.limit = n
}

// GetLimit returns the max number of rows to retrieve.
func (q *BaseQuery) GetLimit() uint64 {
	return q.limit
}

// Offset sets the number of rows to skip.
func (q *BaseQuery) Offset(n uint64) {
	q.offset = n
}

// GetOffset returns the number of rows to skip.
func (q *BaseQuery) GetOffset() uint64 {
	return q.offset
}

// Where adds a new condition to filter the query. All conditions added are
// concatenated with "and".
//   q.Where(Eq(NameColumn, "foo"))
//   q.Where(Gt(AgeColumn, 18))
//   // ... WHERE name = "foo" AND age > 18
func (q *BaseQuery) Where(cond Condition) {
	q.builder = q.builder.Where(cond(q.schema))
}

// compile returns the selected column names and the select builder.
func (q *BaseQuery) compile() ([]string, squirrel.SelectBuilder) {
	columns := q.selectedColumns()
	var (
		qualifiedColumns = make([]string, len(columns))
		columnNames      = make([]string, len(columns))
	)

	for i := range columns {
		qualifiedColumns[i] = columns[i].QualifiedName(q.schema)
		columnNames[i] = columns[i].String()
	}
	return columnNames, q.builder.Columns(
		append(qualifiedColumns, q.relationColumns...)...,
	)
}

// String returns the SQL generated by the
func (q *BaseQuery) String() string {
	_, builder := q.compile()
	sql, _, _ := builder.ToSql()
	return sql
}

// ColumnOrder is a column name with its order.
type ColumnOrder interface {
	// ToSql returns the SQL representation of the column with its order.
	ToSql(Schema) string
	isColumnOrder()
}

type colOrder struct {
	order string
	col   SchemaField
}

func (o *colOrder) ToSql(schema Schema) string {
	return fmt.Sprintf("%s %s", o.col.QualifiedName(schema), o.order)
}
func (colOrder) isColumnOrder() {}

const (
	asc  = "ASC"
	desc = "DESC"
)

// Asc returns a column ordered by ascending order.
func Asc(col SchemaField) ColumnOrder {
	return &colOrder{asc, col}
}

// Desc returns a column ordered by descending order.
func Desc(col SchemaField) ColumnOrder {
	return &colOrder{desc, col}
}
