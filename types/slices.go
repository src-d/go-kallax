package types

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type slice struct {
	val interface{}
}

// Slice wraps a slice value so it can be scanned from and converted to
// PostgreSQL arrays. The following values can be used with this function:
//  - slices of all Go basic types
//  - slices of *url.URL and url.URL
//  - slices of types that implement sql.Scanner and driver.Valuer (take into
//    account that these make use of reflection for scan/value)
//
// NOTE: Keep in mind to always use the following types in the database schema
// to keep it in sync with the values allowed in Go.
//  - int64: bigint
//  - uint64: numeric(20)
//  - int: bigint
//  - uint: numeric(20)
//  - int32: integer
//  - uint32: bigint
//  - int16: smallint
//  - uint16: integer
//  - int8: smallint
//  - uint8: smallint
// Use them with care.
func Slice(v interface{}) SQLType {
	switch v := v.(type) {
	case []url.URL:
		return Slice(&v)
	case []uint64:
		return (*Uint64Array)(&v)
	case *[]uint64:
		return (*Uint64Array)(v)
	case []int:
		return (*IntArray)(&v)
	case *[]int:
		return (*IntArray)(v)
	case []uint:
		return (*UintArray)(&v)
	case *[]uint:
		return (*UintArray)(v)
	case []int32:
		return (*Int32Array)(&v)
	case *[]int32:
		return (*Int32Array)(v)
	case []uint32:
		return (*Uint32Array)(&v)
	case *[]uint32:
		return (*Uint32Array)(v)
	case []int16:
		return (*Int16Array)(&v)
	case *[]int16:
		return (*Int16Array)(v)
	case []uint16:
		return (*Uint16Array)(&v)
	case *[]uint16:
		return (*Uint16Array)(v)
	case []int8:
		return (*Int8Array)(&v)
	case *[]int8:
		return (*Int8Array)(v)
	case []byte:
		return (*ByteArray)(&v)
	case *[]byte:
		return (*ByteArray)(v)
	case *[]float32:
		return (*Float32Array)(v)
	case []float32:
		return (*Float32Array)(&v)
	}
	return &slice{v}
}

func (a *slice) Scan(v interface{}) error {
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

func (a slice) Value() (driver.Value, error) {
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

// Uint64Array represents a one-dimensional array of the PostgreSQL unsigned bigint type.
type Uint64Array []uint64

// Scan implements the sql.Scanner interface.
func (a *Uint64Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Uint64Array", src)
}

func (a *Uint64Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Uint64Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Uint64Array, len(elems))
		for i, v := range elems {
			if b[i], err = strconv.ParseUint(string(v), 10, 64); err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Uint64Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendUint(b, a[0], 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendUint(b, a[i], 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// IntArray represents a one-dimensional array of the PostgreSQL integer type.
type IntArray []int

// Scan implements the sql.Scanner interface.
func (a *IntArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to IntArray", src)
}

func (a *IntArray) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "IntArray")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(IntArray, len(elems))
		for i, v := range elems {
			if b[i], err = strconv.Atoi(string(v)); err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, int64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, int64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// UintArray represents a one-dimensional array of the PostgreSQL unsigned integer type.
type UintArray []uint

// Scan implements the sql.Scanner interface.
func (a *UintArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to UintArray", src)
}

func (a *UintArray) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "UintArray")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(UintArray, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseUint(string(v), 10, 0)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = uint(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a UintArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendUint(b, uint64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendUint(b, uint64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Int32Array represents a one-dimensional array of the PostgreSQL integer type.
type Int32Array []int32

// Scan implements the sql.Scanner interface.
func (a *Int32Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Int32Array", src)
}

func (a *Int32Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Int32Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Int32Array, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseInt(string(v), 10, 32)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = int32(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Int32Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, int64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, int64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Uint32Array represents a one-dimensional array of the PostgreSQL unsigned integer type.
type Uint32Array []uint32

// Scan implements the sql.Scanner interface.
func (a *Uint32Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Uint32Array", src)
}

func (a *Uint32Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Uint32Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Uint32Array, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseUint(string(v), 10, 32)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = uint32(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Uint32Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendUint(b, uint64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendUint(b, uint64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Int16Array represents a one-dimensional array of the PostgreSQL integer type.
type Int16Array []int16

// Scan implements the sql.Scanner interface.
func (a *Int16Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Int16Array", src)
}

func (a *Int16Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Int16Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Int16Array, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseInt(string(v), 10, 16)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = int16(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Int16Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, int64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, int64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Uint16Array represents a one-dimensional array of the PostgreSQL unsigned integer type.
type Uint16Array []uint16

// Scan implements the sql.Scanner interface.
func (a *Uint16Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Uint16Array", src)
}

func (a *Uint16Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Uint16Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Uint16Array, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseUint(string(v), 10, 16)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = uint16(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Uint16Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendUint(b, uint64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendUint(b, uint64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Int8Array represents a one-dimensional array of the PostgreSQL integer type.
type Int8Array []int8

// Scan implements the sql.Scanner interface.
func (a *Int8Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Int8Array", src)
}

func (a *Int8Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Int8Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Int8Array, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseInt(string(v), 10, 8)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = int8(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Int8Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, int64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, int64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// ByteArray represents a byte array `bytea`.
type ByteArray []uint8

// Scan implements the sql.Scanner interface.
func (a *ByteArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		dst := *(*[]byte)(a)

		if a == nil || cap(dst) < len(src) {
			// nil, or shorter destination than source. Create a new slice.
			dst = make([]byte, len(src))
		} else {
			// Resize slice, but retain capacity
			dst = dst[0:len(src)]
		}

		// Copy, do not retain reference to reference types.
		copy(dst, src)

		*(*[]byte)(a) = dst
		return nil
	case string:
		*(*[]byte)(a) = []byte(src)
		return nil
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to ByteArray", src)
}

// Value implements the driver.Valuer interface.
func (a ByteArray) Value() (driver.Value, error) {
	return ([]byte)(a), nil
}

// Float32Array represents a one-dimensional array of the PostgreSQL real type.
type Float32Array []float32

// Scan implements the sql.Scanner interface.
func (a *Float32Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("kallax: cannot convert %T to Float32Array", src)
}

func (a *Float32Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Float32Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Float32Array, len(elems))
		for i, v := range elems {
			val, err := strconv.ParseFloat(string(v), 32)
			if err != nil {
				return fmt.Errorf("kallax: parsing array element index %d: %v", i, err)
			}
			b[i] = float32(val)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Float32Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendFloat(b, float64(a[0]), 'f', -1, 64)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendFloat(b, float64(a[i]), 'f', -1, 64)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// parseArray extracts the dimensions and elements of an array represented in
// text format. Only representations emitted by the backend are supported.
// Notably, whitespace around brackets and delimiters is significant, and NULL
// is case-sensitive.
//
// See http://www.postgresql.org/docs/current/static/arrays.html#ARRAYS-IO
//
// Copied from lib/pq (https://github.com/lib/pq/blob/master/array.go#L636)
// changing pq for kallax in the errors to distinguish between errors from pq
// and the ones originated in kallax.
func parseArray(src, del []byte) (dims []int, elems [][]byte, err error) {
	var depth, i int

	if len(src) < 1 || src[0] != '{' {
		return nil, nil, fmt.Errorf("kallax: unable to parse array; expected %q at offset %d", '{', 0)
	}

Open:
	for i < len(src) {
		switch src[i] {
		case '{':
			depth++
			i++
		case '}':
			elems = make([][]byte, 0)
			goto Close
		default:
			break Open
		}
	}
	dims = make([]int, i)

Element:
	for i < len(src) {
		switch src[i] {
		case '{':
			if depth == len(dims) {
				break Element
			}
			depth++
			dims[depth-1] = 0
			i++
		case '"':
			var elem = []byte{}
			var escape bool
			for i++; i < len(src); i++ {
				if escape {
					elem = append(elem, src[i])
					escape = false
				} else {
					switch src[i] {
					default:
						elem = append(elem, src[i])
					case '\\':
						escape = true
					case '"':
						elems = append(elems, elem)
						i++
						break Element
					}
				}
			}
		default:
			for start := i; i < len(src); i++ {
				if bytes.HasPrefix(src[i:], del) || src[i] == '}' {
					elem := src[start:i]
					if len(elem) == 0 {
						return nil, nil, fmt.Errorf("kallax: unable to parse array; unexpected %q at offset %d", src[i], i)
					}
					if bytes.Equal(elem, []byte("NULL")) {
						elem = nil
					}
					elems = append(elems, elem)
					break Element
				}
			}
		}
	}

	for i < len(src) {
		if bytes.HasPrefix(src[i:], del) && depth > 0 {
			dims[depth-1]++
			i += len(del)
			goto Element
		} else if src[i] == '}' && depth > 0 {
			dims[depth-1]++
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("kallax: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}

Close:
	for i < len(src) {
		if src[i] == '}' && depth > 0 {
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("kallax: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}
	if depth > 0 {
		err = fmt.Errorf("kallax: unable to parse array; expected %q at offset %d", '}', i)
	}
	if err == nil {
		for _, d := range dims {
			if (len(elems) % d) != 0 {
				err = fmt.Errorf("kallax: multidimensional arrays must have elements with matching dimensions")
			}
		}
	}
	return
}

func scanLinearArray(src, del []byte, typ string) (elems [][]byte, err error) {
	dims, elems, err := parseArray(src, del)
	if err != nil {
		return nil, err
	}
	if len(dims) > 1 {
		return nil, fmt.Errorf("kallax: cannot convert ARRAY%s to %s", strings.Replace(fmt.Sprint(dims), " ", "][", -1), typ)
	}
	return elems, err
}
