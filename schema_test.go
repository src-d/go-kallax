package kallax

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var emptySchema = NewBaseSchema("", "", nil, nil, nil, false)

func TestBaseSchemaFieldQualifiedName(t *testing.T) {
	var cases = []struct {
		name     string
		field    SchemaField
		schema   Schema
		expected string
	}{
		{"non empty schema alias", f("foo"), ModelSchema, "__model.foo"},
		{"empty schema alias", f("foo"), emptySchema, "foo"},
	}

	r := require.New(t)
	for _, c := range cases {
		r.Equal(c.expected, c.field.QualifiedName(c.schema), c.name)
	}
}

func TestJSONSchemaKeyQualifiedName(t *testing.T) {
	var cases = []struct {
		name     string
		key      *JSONSchemaKey
		schema   Schema
		expected string
	}{
		{
			"json text key",
			NewJSONSchemaKey(JSONText, "foo", "bar", "baz"),
			ModelSchema,
			"__model.foo #>>'{bar,baz}'",
		},
		{
			"json int key",
			NewJSONSchemaKey(JSONInt, "foo", "bar", "baz"),
			ModelSchema,
			"CAST(__model.foo #>>'{bar,baz}' as bigint)",
		},
		{
			"json any key",
			NewJSONSchemaKey(JSONAny, "foo", "bar", "baz"),
			ModelSchema,
			"__model.foo #>'{bar,baz}'",
		},
		{
			"json key with empty schema",
			NewJSONSchemaKey(JSONBool, "foo", "bar", "baz"),
			nil,
			"CAST(foo #>>'{bar,baz}' as bool)",
		},
	}

	r := require.New(t)
	for _, c := range cases {
		r.Equal(c.expected, c.key.QualifiedName(c.schema), c.name)
	}
}
