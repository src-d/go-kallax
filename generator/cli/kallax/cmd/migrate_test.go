package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathToFileURL(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`c:\foo\bar\baz`, "file:///c:/foo/bar/baz"},
		{"/foo/bar/baz", "file:///foo/bar/baz"},
	}

	for _, tt := range cases {
		require.Equal(t, tt.expected, pathToFileURL(tt.input), tt.input)
	}
}
