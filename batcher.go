package kallax

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type batchQueryRunner struct {
	schema        Schema
	cols          []string
	q             Query
	oneToOneRels  []Relationship
	oneToManyRels []Relationship
	throughRels   []Relationship
	db            squirrel.DBProxy
	builder       squirrel.SelectBuilder
	total         int
	eof           bool
	// records is the cache of the records in the last batch.
	records []Record
}

var errNoMoreRows = errors.New("kallax: there are no more rows in the result set")

func newBatchQueryRunner(schema Schema, db squirrel.DBProxy, q Query) *batchQueryRunner {
	cols, builder := q.compile()
	var (
		oneToOneRels  []Relationship
		oneToManyRels []Relationship
		throughRels   []Relationship
	)

	for _, rel := range q.getRelationships() {
		switch rel.Type {
		case OneToOne:
			oneToOneRels = append(oneToOneRels, rel)
		case OneToMany:
			oneToManyRels = append(oneToManyRels, rel)
		case Through:
			throughRels = append(throughRels, rel)
		}
	}

	return &batchQueryRunner{
		schema:        schema,
		cols:          cols,
		q:             q,
		oneToOneRels:  oneToOneRels,
		oneToManyRels: oneToManyRels,
		throughRels:   throughRels,
		db:            db,
		builder:       builder,
	}
}

func (r *batchQueryRunner) next() (Record, error) {
	if r.eof {
		return nil, errNoMoreRows
	}

	if len(r.records) == 0 {
		var (
			records []Record
			err     error
		)

		limit := r.q.GetLimit()
		if limit <= 0 || limit > uint64(r.total) {
			records, err = r.loadNextBatch()
			if err != nil {
				return nil, err
			}
		}

		if len(records) == 0 {
			r.eof = true
			return nil, errNoMoreRows
		}

		r.total += len(records)
		r.records = records[1:]
		return records[0], nil
	}

	record := r.records[0]
	r.records = r.records[1:]
	return record, nil
}

func (r *batchQueryRunner) loadNextBatch() ([]Record, error) {
	limit := r.q.GetLimit() - uint64(r.total)
	if r.q.GetBatchSize() < limit || limit <= 0 {
		limit = r.q.GetBatchSize()
	}

	rows, err := r.builder.
		Offset(r.q.GetOffset() + uint64(r.total)).
		Limit(limit).
		RunWith(r.db).
		Query()

	if err != nil {
		return nil, err
	}

	return r.processBatch(rows)
}

func (r *batchQueryRunner) processBatch(rows *sql.Rows) ([]Record, error) {
	batchRs := NewResultSet(
		rows,
		r.q.isReadOnly(),
		r.oneToOneRels,
		r.cols...,
	)

	var records []Record
	for batchRs.Next() {
		var rec = r.schema.New()
		if err := batchRs.Scan(rec); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}

	if err := batchRs.Close(); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	var ids = make([]interface{}, len(records))
	var identType Identifier
	for i, r := range records {
		identType = r.GetID()
		ids[i] = r.GetID().Raw()
	}

	for _, rel := range r.oneToManyRels {
		indexedResults, err := r.getRecordRelationships(ids, rel)
		if err != nil {
			return nil, err
		}

		err = setIndexedResults(records, rel, indexedResults)
		if err != nil {
			return nil, err
		}
	}

	for _, rel := range r.throughRels {
		indexedResults, err := r.getRecordThroughRelationships(ids, rel, identType)
		if err != nil {
			return nil, err
		}

		err = setIndexedResults(records, rel, indexedResults)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}

func setIndexedResults(records []Record, rel Relationship, indexedResults indexedRecords) error {
	for _, r := range records {
		err := r.SetRelationship(rel.Field, indexedResults[r.GetID().Raw()])
		if err != nil {
			return err
		}

		// If the relationship is partial, we can not ensure the results
		// in the field reflect the truth of the database.
		// In this case, the parent is marked as non-writable.
		if rel.Filter != nil {
			r.setWritable(false)
		}
	}

	return nil
}

type indexedRecords map[interface{}][]Record

func (r *batchQueryRunner) getRecordRelationships(ids []interface{}, rel Relationship) (indexedRecords, error) {
	fk, ok := r.schema.ForeignKey(rel.Field)
	if !ok {
		return nil, fmt.Errorf("kallax: cannot find foreign key on field %s of table %s", rel.Field, r.schema.Table())
	}

	filter := In(fk, ids...)
	if rel.Filter != nil {
		filter = And(rel.Filter, filter)
	}

	q := NewBaseQuery(rel.Schema)
	q.Where(filter)
	cols, builder := q.compile()
	rows, err := builder.RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}

	return indexedResultsFromRows(rows, cols, rel.Schema, fk, nil)
}

func (r *batchQueryRunner) getRecordThroughRelationships(ids []interface{}, rel Relationship, identType Identifier) (indexedRecords, error) {
	lfk, rfk, ok := r.schema.ForeignKeys(rel.Field)
	if !ok {
		return nil, fmt.Errorf("kallax: cannot find foreign keys for through relationship on field %s of table %s", rel.Field, r.schema.Table())
	}

	filter := In(r.schema.ID(), ids...)
	if rel.Filter != nil {
		filter = And(rel.Filter, filter)
	}

	if rel.IntermediateFilter != nil {
		filter = And(rel.IntermediateFilter, filter)
	}

	q := NewBaseQuery(rel.Schema)
	lschema := r.schema.WithAlias(rel.Schema.Alias())
	intSchema := rel.IntermediateSchema.WithAlias(rel.Schema.Alias())
	q.joinThrough(lschema, intSchema, rel.Schema, lfk, rfk)
	q.Where(filter)
	cols, builder := q.compile()
	// manually add the extra column to also select the parent id
	builder = builder.Column(lschema.ID().QualifiedName(lschema))
	rows, err := builder.RunWith(r.db).Query()
	if err != nil {
		return nil, err
	}

	// we need to pass a new pointer of the parent identifier type so the
	// resultset can fill it and we can know to which record it belongs when
	// indexing by parent id.
	return indexedResultsFromRows(rows, cols, rel.Schema, rfk, identType.newPtr())
}

// indexedResultsFromRows returns the results in the given rows indexed by the
// parent id. In the case of many to many relationships, the record odes not
// have a specific field with the ID of the parent to index by it,
// that's why parentIDPtr is passed for these cases. parentIDPtr is a pointer
// to an ID of the type required by the parent to be filled by the result set.
func indexedResultsFromRows(rows *sql.Rows, cols []string, schema Schema, fk SchemaField, parentIDPtr interface{}) (indexedRecords, error) {
	relRs := NewResultSet(rows, false, nil, cols...)
	var indexedResults = make(indexedRecords)
	for relRs.Next() {
		var (
			rec Record
			err error
		)

		if parentIDPtr != nil {
			rec, err = relRs.customGet(schema, parentIDPtr)
		} else {
			rec, err = relRs.Get(schema)
		}

		if err != nil {
			return nil, err
		}

		rec.setPersisted()
		rec.setWritable(true)

		var id interface{}
		if parentIDPtr != nil {
			id = parentIDPtr.(Identifier).Raw()
		} else {
			val, err := rec.Value(fk.String())
			if err != nil {
				return nil, err
			}

			id = val.(Identifier).Raw()
		}

		indexedResults[id] = append(indexedResults[id], rec)
	}

	if err := relRs.Close(); err != nil {
		return nil, err
	}

	return indexedResults, nil
}
