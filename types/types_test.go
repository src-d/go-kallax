package types

import (
	"net/url"
	"testing"

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
