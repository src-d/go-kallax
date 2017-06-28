package tests

import (
	"time"

	"gopkg.in/src-d/go-kallax.v1"
)

type AliasSliceString []string

type StoreFixture struct {
	kallax.Model   `table:"store" pk:"id"`
	ID             kallax.ULID
	Foo            string
	SliceProp      []string
	AliasSliceProp AliasSliceString
}

func newStoreFixture() *StoreFixture {
	return &StoreFixture{ID: kallax.NewULID()}
}

type StoreWithConstructFixture struct {
	kallax.Model `table:"store_construct"`
	ID           kallax.ULID `pk:""`
	Foo          string
}

func newStoreWithConstructFixture(f string) *StoreWithConstructFixture {
	if f == "" {
		return nil
	}
	return &StoreWithConstructFixture{ID: kallax.NewULID(), Foo: f}
}

type StoreWithNewFixture struct {
	kallax.Model `table:"store_new"`
	ID           kallax.ULID `pk:""`
	Foo          string
	Bar          string
}

func newStoreWithNewFixture() *StoreWithNewFixture {
	return &StoreWithNewFixture{ID: kallax.NewULID()}
}

type MultiKeySortFixture struct {
	kallax.Model `table:"query"`
	ID           kallax.ULID `pk:""`
	Name         string
	Start        time.Time
	End          time.Time
}

func newMultiKeySortFixture() *MultiKeySortFixture {
	return &MultiKeySortFixture{ID: kallax.NewULID()}
}

type SomeJSON struct {
	Foo int
}

type Nullable struct {
	kallax.Model `table:"nullable"`
	ID           int64 `pk:"autoincr"`
	T            *time.Time
	SomeJSON     *SomeJSON
	Scanner      *kallax.ULID
}

type Parent struct {
	kallax.Model `table:"parents" pk:"id,autoincr"`
	ID           int64
	Name         string
	Children     []*Child
}

type Child struct {
	kallax.Model `table:"children"`
	ID           int64 `pk:"autoincr"`
	Name         string
}

type ParentNoPtr struct {
	kallax.Model `table:"parents"`
	ID           int64 `pk:"autoincr"`
	Name         string
	Children     []Child `fk:"parent_id"`
}
