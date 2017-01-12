package generator

import (
	"reflect"

	. "gopkg.in/check.v1"
)

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s *TypesSuite) TestFieldInline(c *C) {
	tests := []struct {
		tag    string
		inline bool
	}{
		{"", false},
		{`bson:"foo"`, false},
		{`bson:"foo,inline"`, true},
		{`bson:"foo,inline,omitempty"`, true},
		{`bson:",inline,omitempty"`, true},
		{`bson:",inline"`, true},
	}

	for _, t := range tests {
		c.Assert(NewField("", "", reflect.StructTag(t.tag)).Inline(), Equals, t.inline)
	}
}
