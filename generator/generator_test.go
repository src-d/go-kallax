package generator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMigrationGeneratorLoadLock(t *testing.T) {
	dir, err := ioutil.TempDir("", "kallax-migration-generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	g := NewMigrationGenerator("migration", dir)
	schema, err := g.LoadLock()
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.Len(t, schema.Tables, 0)

	content, err := mkSchema(mkTable("foo")).MarshalText()
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(dir, string(migrationLock)), content, 0755)
	require.NoError(t, err)

	schema, err = g.LoadLock()
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.Len(t, schema.Tables, 1)
}

func TestMigrationGeneratorBuild(t *testing.T) {
	dir, err := ioutil.TempDir("", "kallax-migration-generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	g := NewMigrationGenerator("migration", dir)
	content, err := mkSchema(mkTable("foo")).MarshalText()
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(dir, string(migrationLock)), content, 0755)
	require.NoError(t, err)

	migration, err := g.Build()
	require.NoError(t, err)
	require.NotNil(t, migration)
}

func TestMigrationGeneratorGenerate(t *testing.T) {
	old := mkSchema(table1)
	new := mkSchema(table1, table2)
	migration, err := NewMigration(old, new)
	require.NoError(t, err)

	dir, err := ioutil.TempDir("", "kallax-migration-generator")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	g := NewMigrationGenerator("migration", dir)
	g.now = func() time.Time {
		var t time.Time
		return t
	}

	require.NoError(t, g.Generate(migration))

	content, err := ioutil.ReadFile(g.migrationFile(migrationUp, g.now()))
	require.NoError(t, err)
	require.Equal(t, expectedTable2+"\n\n", string(content))

	content, err = ioutil.ReadFile(g.migrationFile(migrationDown, g.now()))
	require.NoError(t, err)
	require.Equal(t, "DROP TABLE table2;\n\n", string(content))

	expected, err := migration.Lock.MarshalText()
	require.NoError(t, err)

	content, err = ioutil.ReadFile(filepath.Join(dir, string(migrationLock)))
	require.NoError(t, err)
	require.Equal(t, string(expected), string(content))
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"the fancy slug", "the_fancy_slug"},
		{"ThE-FaNcYnEss", "the_fancyness"},
		{"this is: a migration", "this_is_a_migration"},
		{"add caché", "add_cach"},
	}

	for _, c := range cases {
		require.Equal(t, c.expected, slugify(c.input))
	}
}
