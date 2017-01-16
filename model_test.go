package kallax

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID_IsEmpty(t *testing.T) {
	require := require.New(t)

	var id ID
	require.True(id.IsEmpty())

	id = NewID()
	require.False(id.IsEmpty())
}
