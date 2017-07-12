package kallax

import (
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
)

func TestOneToManyWithFilterNotWritable(t *testing.T) {
	r := require.New(t)
	db, err := openTestDB()
	r.NoError(err)
	setupTables(t, db)
	defer db.Close()
	defer teardownTables(t, db)

	store := NewStore(db)
	m := newModel("foo", "bar", 1)
	r.NoError(store.Insert(ModelSchema, m))

	for i := 0; i < 4; i++ {
		r.NoError(store.Insert(RelSchema, newRel(m.GetID(), fmt.Sprint(i))))
	}

	q := NewBaseQuery(ModelSchema)
	r.NoError(q.AddRelation(RelSchema, "rels", OneToMany, Eq(f("foo"), "1")))
	runner := newBatchQueryRunner(ModelSchema, squirrel.NewStmtCacher(db), q)
	record, err := runner.next()
	r.NoError(err)
	r.False(record.IsWritable())
}

func TestBatcherLimit(t *testing.T) {
	r := require.New(t)
	db, err := openTestDB()
	r.NoError(err)
	setupTables(t, db)
	defer db.Close()
	defer teardownTables(t, db)

	store := NewStore(db)
	for i := 0; i < 10; i++ {
		m := newModel("foo", "bar", 1)
		r.NoError(store.Insert(ModelSchema, m))

		for i := 0; i < 4; i++ {
			r.NoError(store.Insert(RelSchema, newRel(m.GetID(), fmt.Sprint(i))))
		}
	}

	q := NewBaseQuery(ModelSchema)
	q.BatchSize(2)
	q.Limit(5)
	r.NoError(q.AddRelation(RelSchema, "rels", OneToMany, Eq(f("foo"), "1")))
	runner := newBatchQueryRunner(ModelSchema, store.proxy, q)
	rs := NewBatchingResultSet(runner)

	var count int
	for rs.Next() {
		_, err := rs.Get(nil)
		r.NoError(err)
		count++
	}
	r.NoError(err)
	r.Equal(5, count)
}

func TestBatcherNoExtraQueryIfLessThanLimit(t *testing.T) {
	r := require.New(t)
	db, err := openTestDB()
	r.NoError(err)
	setupTables(t, db)
	defer db.Close()
	defer teardownTables(t, db)

	store := NewStore(db)
	for i := 0; i < 4; i++ {
		m := newModel("foo", "bar", 1)
		r.NoError(store.Insert(ModelSchema, m))

		for i := 0; i < 4; i++ {
			r.NoError(store.Insert(RelSchema, newRel(m.GetID(), fmt.Sprint(i))))
		}
	}

	q := NewBaseQuery(ModelSchema)
	q.Limit(6)
	r.NoError(q.AddRelation(RelSchema, "rels", OneToMany, Eq(f("foo"), "1")))
	var queries int
	proxy := store.DebugWith(func(_ string, _ ...interface{}) {
		queries++
	}).proxy
	runner := newBatchQueryRunner(ModelSchema, proxy, q)
	rs := NewBatchingResultSet(runner)

	var count int
	for rs.Next() {
		_, err := rs.Get(nil)
		r.NoError(err)
		count++
	}
	r.NoError(err)
	r.Equal(4, count)
	r.Equal(2, queries)
}
