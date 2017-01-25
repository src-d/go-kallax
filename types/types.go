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
			return fmt.Errorf("kallax: error scanning url: %s", err)
		}

		*u = URL(*url)
		return nil
	}
	return fmt.Errorf("kallax: cannot scan type %s into URL type", reflect.TypeOf(v))
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

	return fmt.Errorf("kallax: cannot scan type %s into JSON type", reflect.TypeOf(v))
}

// JSONValue converts something into json.
// WARNING: This is here temporarily, might be removed in the future, use
// `JSON` instead.
func JSONValue(v interface{}) (driver.Value, error) {
	return json.Marshal(v)
}

// SQLType is the common interface a type has to fulfill to be considered a
// SQL type.
type SQLType interface {
	sql.Scanner
	driver.Valuer
}

type array struct {
	val  reflect.Value
	size int
}

func Array(v interface{}, size int) SQLType {
	return &array{reflect.ValueOf(v), size}
}

func (a *array) Scan(v interface{}) error {
	sliceTyp := reflect.SliceOf(a.val.Type().Elem().Elem())
	newSlice := reflect.MakeSlice(sliceTyp, 0, 0)
	slicePtr := reflect.New(sliceTyp)
	slicePtr.Elem().Set(newSlice)
	if err := pq.Array(slicePtr.Interface()).Scan(v); err != nil {
		return err
	}

	if slicePtr.Elem().Len() != a.size {
		return fmt.Errorf(
			"kallax: cannot scan array of size %d into array of size %d",
			newSlice.Len(),
			a.size,
		)
	}

	for i := 0; i < a.size; i++ {
		a.val.Elem().Index(i).Set(slicePtr.Elem().Index(i))
	}

	return nil
}

func (a *array) Value() (driver.Value, error) {
	sliceTyp := reflect.SliceOf(a.val.Type().Elem().Elem())
	newSlice := reflect.MakeSlice(sliceTyp, a.size, a.size)
	for i := 0; i < a.size; i++ {
		newSlice.Index(i).Set(a.val.Elem().Index(i))
	}

	slicePtr := reflect.New(sliceTyp)
	slicePtr.Elem().Set(newSlice)
	return pq.Array(slicePtr.Interface()).Value()
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
