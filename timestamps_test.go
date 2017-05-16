package kallax

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimestampsBeforeSave(t *testing.T) {
	s := require.New(t)

	var ts Timestamps
	s.True(ts.CreatedAt.IsZero())
	s.True(ts.UpdatedAt.IsZero())

	s.NoError(ts.BeforeSave())
	s.False(ts.CreatedAt.IsZero())
	s.False(ts.UpdatedAt.IsZero())

	createdAt := ts.CreatedAt
	updatedAt := ts.UpdatedAt
	time.Sleep(50 * time.Millisecond)
	s.NoError(ts.BeforeSave())
	s.Equal(createdAt, ts.CreatedAt)
	s.NotEqual(updatedAt, ts.UpdatedAt)
}
