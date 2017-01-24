package types

import (
	"database/sql"
	"fmt"
	"math"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestSlice(t *testing.T) {
	require := require.New(t)

	cases := []struct {
		v     interface{}
		input interface{}
		dest  interface{}
	}{
		{
			&([]url.URL{mustURL("https://foo.com"), mustURL("http://foo.foo")}),
			[]string{"https://foo.com", "http://foo.foo"},
			&([]url.URL{}),
		},
		{
			&([]*url.URL{mustPtrURL("https://foo.com"), mustPtrURL("http://foo.foo")}),
			[]string{"https://foo.com", "http://foo.foo"},
			&([]*url.URL{}),
		},
		{
			&([]string{"a", "b"}),
			[]string{"a", "b"},
			&([]string{}),
		},
		{
			&([]uint64{123, 321, 333}),
			[]uint64{123, 321, 333},
			&([]uint64{}),
		},
		{
			&([]int{123, 321, 333}),
			[]int{123, 321, 333},
			&([]int{}),
		},
		{
			&([]uint{123, 321, 333}),
			[]uint{123, 321, 333},
			&([]uint{}),
		},
		{
			&([]int32{123, 321, 333}),
			[]int32{123, 321, 333},
			&([]int32{}),
		},
		{
			&([]uint32{123, 321, 333}),
			[]uint32{123, 321, 333},
			&([]uint32{}),
		},
		{
			&([]int16{123, 321, 333}),
			[]int16{123, 321, 333},
			&([]int16{}),
		},
		{
			&([]uint16{123, 321, 333}),
			[]uint16{123, 321, 333},
			&([]uint16{}),
		},
		{
			&([]int8{1, 3, 4}),
			[]int8{1, 3, 4},
			&([]int8{}),
		},
		{
			&([]uint8{1, 3, 4}),
			[]uint8{1, 3, 4},
			&([]uint8{}),
		},
		{
			&([]byte{1, 3, 4}),
			[]byte{1, 3, 4},
			&([]byte{}),
		},
		{
			&([]float32{1., 3., .4}),
			[]float32{1., 3., .4},
			&([]float32{}),
		},
	}

	for _, c := range cases {
		arr := Slice(c.v)
		val, err := arr.Value()
		require.Nil(err)

		pqArr := pq.Array(c.input)
		pqVal, err := pqArr.Value()
		require.Nil(err)

		require.Equal(pqVal, val)
		require.Nil(Slice(c.dest).Scan(val))
		require.Equal(c.v, c.dest)
	}
}

func TestSlice_Integration(t *testing.T) {
	s := require.New(t)
	cases := []struct {
		name  string
		typ   string
		input interface{}
		dst   interface{}
	}{
		{
			"int8",
			"smallint",
			[]int8{math.MaxInt8, math.MinInt8},
			&([]int8{}),
		},
		{
			"unsigned int8",
			"smallint",
			[]uint8{math.MaxUint8, 0},
			&([]uint8{}),
		},
		{
			"int16",
			"smallint",
			[]int16{math.MaxInt16, math.MinInt16},
			&([]int16{}),
		},
		{
			"unsigned int16",
			"integer",
			[]uint16{math.MaxUint16, 0},
			&([]uint16{}),
		},
		{
			"int32",
			"integer",
			[]int32{math.MaxInt32, math.MinInt32},
			&([]int32{}),
		},
		{
			"unsigned int32",
			"bigint",
			[]uint32{math.MaxUint32, 0},
			&([]uint32{}),
		},
		{
			"int/int64",
			"bigint",
			[]int{math.MaxInt64, math.MinInt64},
			&([]int{}),
		},
		{
			"unsigned int/int64",
			"numeric(20)",
			[]uint{math.MaxUint64, 0},
			&([]uint{}),
		},
		{
			"float32",
			"decimal(10,3)",
			[]float32{.3, .6},
			&([]float32{.3, .6}),
		},
	}

	db, err := openTestDB()
	s.Nil(err)

	defer func() {
		_, err = db.Exec("DROP TABLE IF EXISTS foo")
		s.Nil(err)

		s.Nil(db.Close())
	}()

	for _, c := range cases {
		_, err := db.Exec(fmt.Sprintf(`CREATE TABLE foo (
			testcol %s[]
		)`, c.typ))
		s.Nil(err, c.name)

		_, err = db.Exec("INSERT INTO foo (testcol) VALUES ($1)", Slice(c.input))
		s.Nil(err, c.name)

		s.Nil(db.QueryRow("SELECT testcol FROM foo LIMIT 1").Scan(Slice(c.dst)), c.name)
		slice := reflect.ValueOf(c.dst).Elem().Interface()
		s.Equal(c.input, slice, c.name)

		_, err = db.Exec("DROP TABLE foo")
		s.Nil(err, c.name)
	}
}

func envOrDefault(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		v = def
	}
	return v
}

func openTestDB() (*sql.DB, error) {
	return sql.Open("postgres", fmt.Sprintf(
		"postgres://%s:%s@0.0.0.0:5432/%s?sslmode=disable",
		envOrDefault("DBUSER", "testing"),
		envOrDefault("DBPASS", "testing"),
		envOrDefault("DBNAME", "testing"),
	))
}
