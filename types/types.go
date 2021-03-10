// Package types provides implementation of some wrapper SQL types.
package types // import "github.com/loyalguru/go-kallax/types"

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/lib/pq"
)

// SQLType is the common interface a type has to fulfill to be considered a
// SQL type.
type SQLType interface {
	sql.Scanner
	driver.Valuer
}

// Nullable converts the given type (which should be a pointer ideally) into
// a nullable type. For that, it must be either a pointer of a basic Go type or
// a type that implements sql.Scanner itself.
// time.Time and time.Duration are also supported, even though they are none of
// the above.
// If the given types does not fall into any of the above categories, it will
// actually return a valid sql.Scanner that will fail only when the Scan is
// performed.
func Nullable(typ interface{}) sql.Scanner {
	switch typ := typ.(type) {
	case *string:
		return &nullString{typ}
	case *bool:
		return &nullBool{typ}
	case *int8:
		return &nullInt8{typ}
	case *uint8:
		return &nullUint8{typ}
	case *int16:
		return &nullInt16{typ}
	case *uint16:
		return &nullUint16{typ}
	case *uint:
		return &nullUint{typ}
	case *int:
		return &nullInt{typ}
	case *uint32:
		return &nullUint32{typ}
	case *int32:
		return &nullInt32{typ}
	case *uint64:
		return &nullUint64{typ}
	case *int64:
		return &nullInt64{typ}
	case *float32:
		return &nullFloat32{typ}
	case *float64:
		return &nullFloat64{typ}
	case *time.Time:
		return &nullTime{typ}
	case *time.Duration:
		return &nullDuration{typ}
	case sql.Scanner:
		return &nullable{typ}
	case **string:
		return &nullPtrString{typ}
	case **bool:
		return &nullPtrBool{typ}
	case **int8:
		return &nullPtrInt8{typ}
	case **uint8:
		return &nullPtrUint8{typ}
	case **int16:
		return &nullPtrInt16{typ}
	case **uint16:
		return &nullPtrUint16{typ}
	case **uint:
		return &nullPtrUint{typ}
	case **int:
		return &nullPtrInt{typ}
	case **uint32:
		return &nullPtrUint32{typ}
	case **int32:
		return &nullPtrInt32{typ}
	case **uint64:
		return &nullPtrUint64{typ}
	case **int64:
		return &nullPtrInt64{typ}
	case **float32:
		return &nullPtrFloat32{typ}
	case **float64:
		return &nullPtrFloat64{typ}
	case **time.Time:
		return &nullPtrTime{typ}
	case **time.Duration:
		return &nullPtrDuration{typ}
	}

	return &nullableErr{typ}
}

type nullableErr struct {
	v interface{}
}

func (n *nullableErr) Scan(_ interface{}) error {
	return fmt.Errorf("kallax: type %T is not nullable and cannot be scanned", n.v)
}

type nullable struct {
	typ sql.Scanner
}

func (n *nullable) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	return n.typ.Scan(v)
}

type nullPtrString struct {
	v **string
}

func (n *nullPtrString) Scan(v interface{}) error {
	ns := new(sql.NullString)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = &ns.String
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrBool struct {
	v **bool
}

func (n *nullPtrBool) Scan(v interface{}) error {
	ns := new(sql.NullBool)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = &ns.Bool
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrInt8 struct {
	v **int8
}

func (n *nullPtrInt8) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := int8(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrUint8 struct {
	v **uint8
}

func (n *nullPtrUint8) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := uint8(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrInt16 struct {
	v **int16
}

func (n *nullPtrInt16) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := int16(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrUint16 struct {
	v **uint16
}

func (n *nullPtrUint16) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := uint16(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrInt32 struct {
	v **int32
}

func (n *nullPtrInt32) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := int32(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrUint32 struct {
	v **uint32
}

func (n *nullPtrUint32) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := uint32(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrInt struct {
	v **int
}

func (n *nullPtrInt) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := int(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrUint struct {
	v **uint
}

func (n *nullPtrUint) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}
	// TODO: better handling of this type
	if ns.Valid {
		v := uint(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrUint64 struct {
	v **uint64
}

func (n *nullPtrUint64) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}
	// TODO: better handling of this type
	if ns.Valid {
		v := uint64(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrInt64 struct {
	v **int64
}

func (n *nullPtrInt64) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = &ns.Int64
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrFloat32 struct {
	v **float32
}

func (n *nullPtrFloat32) Scan(v interface{}) error {
	ns := new(sql.NullFloat64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := float32(ns.Float64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrFloat64 struct {
	v **float64
}

func (n *nullPtrFloat64) Scan(v interface{}) error {
	ns := new(sql.NullFloat64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = &ns.Float64
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrTime struct {
	v **time.Time
}

func (n *nullPtrTime) Scan(v interface{}) error {
	ns := new(pq.NullTime)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = &ns.Time
	} else {
		*n.v = nil
	}

	return nil
}

type nullPtrDuration struct {
	v **time.Duration
}

func (n *nullPtrDuration) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		v := time.Duration(ns.Int64)
		*n.v = &v
	} else {
		*n.v = nil
	}

	return nil
}

type nullString struct {
	v *string
}

func (n *nullString) Scan(v interface{}) error {
	ns := new(sql.NullString)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = ns.String
	}

	return nil
}

type nullBool struct {
	v *bool
}

func (n *nullBool) Scan(v interface{}) error {
	ns := new(sql.NullBool)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = ns.Bool
	}

	return nil
}

type nullInt8 struct {
	v *int8
}

func (n *nullInt8) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = int8(ns.Int64)
	}

	return nil
}

type nullUint8 struct {
	v *uint8
}

func (n *nullUint8) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = uint8(ns.Int64)
	}

	return nil
}

type nullInt16 struct {
	v *int16
}

func (n *nullInt16) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = int16(ns.Int64)
	}

	return nil
}

type nullUint16 struct {
	v *uint16
}

func (n *nullUint16) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = uint16(ns.Int64)
	}

	return nil
}

type nullInt32 struct {
	v *int32
}

func (n *nullInt32) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = int32(ns.Int64)
	}

	return nil
}

type nullUint32 struct {
	v *uint32
}

func (n *nullUint32) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = uint32(ns.Int64)
	}

	return nil
}

type nullInt struct {
	v *int
}

func (n *nullInt) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = int(ns.Int64)
	}

	return nil
}

type nullUint struct {
	v *uint
}

func (n *nullUint) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}
	// TODO: better handling of this type
	if ns.Valid {
		*n.v = uint(ns.Int64)
	}

	return nil
}

type nullUint64 struct {
	v *uint64
}

func (n *nullUint64) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}
	// TODO: better handling of this type
	if ns.Valid {
		*n.v = uint64(ns.Int64)
	}

	return nil
}

type nullInt64 struct {
	v *int64
}

func (n *nullInt64) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = ns.Int64
	}

	return nil
}

type nullFloat32 struct {
	v *float32
}

func (n *nullFloat32) Scan(v interface{}) error {
	ns := new(sql.NullFloat64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = float32(ns.Float64)
	}

	return nil
}

type nullFloat64 struct {
	v *float64
}

func (n *nullFloat64) Scan(v interface{}) error {
	ns := new(sql.NullFloat64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = ns.Float64
	}

	return nil
}

type nullTime struct {
	v *time.Time
}

func (n *nullTime) Scan(v interface{}) error {
	ns := new(pq.NullTime)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = ns.Time
	}

	return nil
}

type nullDuration struct {
	v *time.Duration
}

func (n *nullDuration) Scan(v interface{}) error {
	ns := new(sql.NullInt64)
	if err := ns.Scan(v); err != nil {
		return err
	}

	if ns.Valid {
		*n.v = time.Duration(ns.Int64)
	}

	return nil
}

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

func (u URL) Value() (driver.Value, error) {
	url := url.URL(u)
	return (&url).String(), nil
}

type array struct {
	val  reflect.Value
	size int
}

// Array returns an SQLType for the given array type with a specific size.
// Note that the actual implementation of this relies on reflection, so be
// cautious with its usage.
// The array is scanned using a slice of the same type, so the same
// restrictions as the `Slice` function of this package are applied.
func Array(v interface{}, size int) SQLType {
	return &array{reflect.ValueOf(v), size}
}

func (a *array) Scan(v interface{}) error {
	sliceTyp := reflect.SliceOf(a.val.Type().Elem().Elem())
	newSlice := reflect.MakeSlice(sliceTyp, 0, 0)
	slicePtr := reflect.New(sliceTyp)
	slicePtr.Elem().Set(newSlice)
	if err := Slice(slicePtr.Interface()).Scan(v); err != nil {
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
	switch v := v.(type) {
	case []byte:
		return json.Unmarshal(v, j.val)
	case string:
		return j.Scan([]byte(v))
	case nil:
		return nil
	}

	return fmt.Errorf("kallax: cannot scan type %s into JSON type", reflect.TypeOf(v))
}

func (j *sqlJSON) Value() (driver.Value, error) {
	return json.Marshal(j.val)
}
