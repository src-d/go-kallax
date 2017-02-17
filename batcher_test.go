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
