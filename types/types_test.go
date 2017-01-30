package types

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestURL(t *testing.T) {
	require := require.New(t)
	expectedURL := "https://foo.com"

	var u URL
	require.Nil(u.Scan(expectedURL))
	require.Equal(expectedURL, urlStr(url.URL(u)))

	u = URL{}
	require.Nil(u.Scan([]byte("https://foo.com")))
	require.Equal(expectedURL, urlStr(url.URL(u)))

	val, err := u.Value()
	require.Nil(err)
	require.Equal(expectedURL, val)
}

func urlStr(u url.URL) string {
	url := &u
	return url.String()
}

func mustURL(u string) url.URL {
	url, _ := url.Parse(u)
	return *url
}

func mustPtrURL(u string) *url.URL {
	url, _ := url.Parse(u)
	return url
}

type jsonType struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestJSON(t *testing.T) {
	input := `{"foo":"a","bar":1}`

	t.Run("into object", func(t *testing.T) {
		var dst jsonType
		expected := jsonType{"a", 1}

		json := JSON(&dst)
		require.Nil(t, json.Scan([]byte(input)))
		require.Equal(t, expected, dst)

		val, err := json.Value()
		require.Nil(t, err)
		require.Equal(t, input, string(val.([]byte)))
	})

	t.Run("into map", func(t *testing.T) {
		var dst = make(map[string]interface{})

		json := JSON(&dst)
		require.Nil(t, json.Scan([]byte(input)))
		val, ok := dst["foo"]
		require.True(t, ok)
		require.Equal(t, "a", val.(string))

		val, ok = dst["bar"]
		require.True(t, ok)
		require.Equal(t, float64(1), val.(float64))
	})
}

func TestArray(t *testing.T) {
	require := require.New(t)
	input, err := pq.Array([]int64{1, 2}).Value()
	require.Nil(err)

	var dst [2]int64

	arr := Array(&dst, 2)
	require.Nil(arr.Scan(input))
	require.Equal(int64(1), dst[0])
	require.Equal(int64(2), dst[1])

	v, err := arr.Value()
	require.Nil(err)
	require.Equal(input, v)
}

func TestNullable(t *testing.T) {
	var (
		Str      string
		Int64    int64
		Float64  float64
		Bool     bool
		Time     time.Time
		Duration time.Duration
	)
	tim := time.Now().UTC()
	tim = time.Date(tim.Year(), tim.Month(), tim.Day(), tim.Hour(), tim.Minute(), tim.Second(), 0, tim.Location())
	s := require.New(t)
	cases := []struct {
		name         string
		typ          string
		nonNullInput interface{}
		dst          interface{}
	}{
		{
			"string",
			"text",
			"foo",
			&Str,
		},
		{
			"int64",
			"bigint",
			int64(1),
			&Int64,
		},
		{
			"float64",
			"decimal",
			float64(.5),
			&Float64,
		},
		{
			"bool",
			"bool",
			true,
			&Bool,
		},
		{
			"time.Duration",
			"bigint",
			3 * time.Second,
			&Duration,
		},
		{
			"time.Time",
			"timestamptz",
			tim,
			&Time,
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
		s.Nil(db.QueryRow("SELECT null").Scan(Nullable(c.dst)), c.name)
		elem := reflect.ValueOf(c.dst).Elem()
		zero := reflect.Zero(elem.Type())
		s.Equal(zero.Interface(), elem.Interface(), c.name)

		var input = c.nonNullInput
		if v, ok := c.nonNullInput.(time.Duration); ok {
			input = int64(v)
		}

		_, err := db.Exec(fmt.Sprintf(`CREATE TABLE foo (
			testcol %s
		)`, c.typ))
		s.Nil(err, c.name)

		_, err = db.Exec("INSERT INTO foo (testcol) VALUES ($1)", input)
		s.Nil(err, c.name)

		s.Nil(db.QueryRow("SELECT testcol FROM foo").Scan(Nullable(c.dst)), c.name)
		elem = reflect.ValueOf(c.dst).Elem()
		s.Equal(c.nonNullInput, elem.Interface(), c.name)

		_, err = db.Exec("DROP TABLE foo")
		s.Nil(err, c.name)
	}
}
