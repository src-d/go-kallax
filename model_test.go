package kallax

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIDIsEmpty(t *testing.T) {
	r := require.New(t)
	var id ID
	r.True(id.IsEmpty())

	id = NewID()
	r.False(id.IsEmpty())
}

func TestID_Value(t *testing.T) {
	id := NewID()
	v, _ := id.Value()
	require.Equal(t, id.String(), v)
}

func TestID_ThreeNewIDsAreDifferent(t *testing.T) {
	r := require.New(t)
	id1 := NewID()
	id2 := NewID()
	id3 := NewID()

	r.NotEqual(id1, id2)
	r.NotEqual(id1, id3)
	r.NotEqual(id2, id3)

	r.True(id1 == id1)
	r.False(id1 == id2)
}

func TestVirtualColumn(t *testing.T) {
	r := require.New(t)
	record := newModel("", "", 0)
	record.virtualColumns = nil
	r.Equal(nil, record.VirtualColumn("foo"))

	record.virtualColumns = nil
	s := VirtualColumn("foo", record)

	id := NewID()
	v, err := id.Value()
	r.NoError(err)
	r.NoError(s.Scan(v))
	r.Len(record.virtualColumns, 1)
	r.Equal(id, record.VirtualColumn("foo"))

	r.Error(s.Scan(nil))
}
