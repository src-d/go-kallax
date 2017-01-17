package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/lib/pq"
)

// URL is a wrapper of url.URL that implements SQLType interface.
type URL url.URL

func (u *URL) Scan(v interface{}) error {
	switch t := v.(type) {
	case []byte:
		return u.Scan(string(t))
	case string:
		url, err := url.Parse(t)
		if err != nil {
			return fmt.Errorf("error scanning url: %s", err)
		}

		*u = URL(*url)
		return nil
	}
	return fmt.Errorf("cannot scan type %s into URL type", reflect.TypeOf(v))
}

func (u URL) Value() (interface{}, error) {
	url := url.URL(u)
	return (&url).String(), nil
}

// ScanJSON scans json v into dst.
// WARNING: This is here temporarily, might be removed in the future, use
// `JSON` instead.
func ScanJSON(v interface{}, dst interface{}) error {
	switch v := v.(type) {
	case []byte:
		return json.Unmarshal(v, dst)
	case string:
		return ScanJSON([]byte(v), dst)
	}

	return fmt.Errorf("cannot scan type %s into JSON type", reflect.TypeOf(v))
}

// JSONValue converts something into json.
// WARNING: This is here temporarily, might be removed in the future, use
// `JSON` instead.
func JSONValue(v interface{}) (driver.Value, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// SQLType is the common interface a type has to fulfill to be considered a
// SQL type.
type SQLType interface {
	sql.Scanner
	driver.Valuer
}

type array struct {
	val interface{}
}

// Array wraps a slice value so it can be scanned from and converted to
// PostgreSQL arrays. The following values can be used with this function:
//  - slices of basic types
//  - slices of *url.URL and url.URL
//  - slices of types that implement sql.Scanner and driver.Valuer (take into
//    account that these make use of reflection for scan/value)
func Array(v interface{}) SQLType {
	switch v := v.(type) {
	case []url.URL:
		return Array(&v)
	}
	return &array{v}
}

func (a *array) Scan(v interface{}) error {
	switch o := a.val.(type) {
	case *[]url.URL:
		var s []string
		if err := pq.Array(&s).Scan(v); err != nil {
			return err
		}
		var res = make([]url.URL, len(s))
		for i, v := range s {
			var u = new(URL)
			if err := u.Scan(v); err != nil {
				return err
			}
			res[i] = url.URL(*u)
		}
		*o = res
		return nil
	case *[]*url.URL:
		var s []string
		if err := pq.Array(&s).Scan(v); err != nil {
			return err
		}
		var res = make([]*url.URL, len(s))
		for i, v := range s {
			var u = new(URL)
			if err := u.Scan(v); err != nil {
				return err
			}
			res[i] = (*url.URL)(u)
		}
		*o = res
		return nil
	}
	return pq.Array(a.val).Scan(v)
}

func (a array) Value() (driver.Value, error) {
	switch v := a.val.(type) {
	case *[]url.URL:
		var s = make([]string, len(*v))
		for i, u := range *v {
			url, err := URL(u).Value()
			if err != nil {
				return nil, err
			}
			s[i] = url.(string)
		}
		return pq.Array(s).Value()
	case *[]*url.URL:
		var s = make([]string, len(*v))
		for i, u := range *v {
			url, err := (*URL)(u).Value()
			if err != nil {
				return nil, err
			}
			s[i] = url.(string)
		}
		return pq.Array(s).Value()
	default:
		return pq.Array(v).Value()
	}
}

type sqlJSON struct {
	val interface{}
}

// JSON makes sure the given value is converted to and scanned from SQL as
// a JSON. Note that this uses the standard json.Unmarshal and json.Marshal
// and it relies on reflection. To speed up the encoding/decoding you can
// implement interfaces json.Marshaller and json.Unmarshaller for your type
// with, for example, ffjson.
func JSON(v interface{}) SQLType {
	return &sqlJSON{v}
}

func (j *sqlJSON) Scan(v interface{}) error {
	return ScanJSON(v, j.val)
}

func (j *sqlJSON) Value() (driver.Value, error) {
	return JSONValue(j.val)
}
