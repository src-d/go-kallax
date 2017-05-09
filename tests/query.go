package tests

import (
	"net/url"
	"time"

	"gopkg.in/src-d/go-kallax.v1"
	"gopkg.in/src-d/go-kallax.v1/tests/fixtures"
)

type QueryFixture struct {
	kallax.Model `table:"query"`
	ID           kallax.ULID `pk:""`

	Relation  *QueryRelationFixture   `fk:"owner_id"`
	Inverse   *QueryRelationFixture   `fk:"inverse_id,inverse"`
	NRelation []*QueryRelationFixture `fk:"owner_id"`
	Embedded  fixtures.QueryDummy
	Ignored   fixtures.QueryDummy `kallax:"-"`
	Inline    struct {
		Inline string
	} `kallax:",inline"`
	MapOfString               map[string]string
	MapOfInterface            map[string]interface{}
	MapOfSomeType             map[string]fixtures.QueryDummy
	Foo                       string
	StringProperty            string
	Integer                   int
	Integer64                 int64
	Float32                   float32
	Boolean                   bool
	ArrayParam                [3]string
	SliceParam                []string
	AliasArrayParam           fixtures.AliasArray
	AliasSliceParam           fixtures.AliasSlice
	AliasStringParam          fixtures.AliasString
	AliasIntParam             fixtures.AliasInt
	DummyParam                fixtures.QueryDummy
	AliasDummyParam           fixtures.AliasDummyParam
	SliceDummyParam           []fixtures.QueryDummy
	IDPropertyParam           kallax.ULID
	InterfacePropParam        fixtures.InterfaceImplementation `sqltype:"jsonb"`
	URLParam                  url.URL
	TimeParam                 time.Time
	AliasArrAliasStringParam  fixtures.AliasArrAliasString
	AliasHereArrayParam       AliasHereArray
	ArrayAliasHereStringParam []AliasHereString
	ScannerValuerParam        ScannerValuer `sqltype:"jsonb"`
}

type AliasHereString string
type AliasHereArray [3]string
type ScannerValuer struct {
	fixtures.ScannerValuer
}

type AliasID kallax.ULID

func newQueryFixture(f string) *QueryFixture {
	return &QueryFixture{ID: kallax.NewULID(), Foo: f}
}

func (q *QueryFixture) Eq(v *QueryFixture) bool {
	return q.ID == v.ID
}

type QueryRelationFixture struct {
	kallax.Model `table:"query_relation"`
	ID           kallax.ULID `pk:""`
	Name         string
	Owner        *QueryFixture `fk:"owner_id,inverse"`
}

var queryFixtures = []*QueryFixture{
	&QueryFixture{
		ID:               kallax.NewULID(),
		Foo:              "Foo0",
		StringProperty:   "StringProperty0",
		Integer:          0,
		Integer64:        0,
		Float32:          0,
		Boolean:          true,
		ArrayParam:       [3]string{"ArrayParam0One", "ArrayParam0Two", "ArrayParam0Three"},
		SliceParam:       []string{"SliceParam0One", "SliceParam0Two", "SliceParam0Three"},
		AliasArrayParam:  [3]string{"AliasArray0One", "AliasArray0Two", "AliasArray0Three"},
		AliasSliceParam:  []string{"AliasSlice0One", "AliasSlice0Two", "AliasSlice0Three"},
		AliasStringParam: "AliasString0",
		AliasIntParam:    0,
	},
	&QueryFixture{
		ID:               kallax.NewULID(),
		Foo:              "Foo1",
		StringProperty:   "StringProperty1",
		Integer:          1,
		Integer64:        1,
		Float32:          1,
		Boolean:          false,
		ArrayParam:       [3]string{"ArrayParm1One", "ArrayParm1Two", "ArrayParm1Three"},
		SliceParam:       []string{"SliceParam1One", "SliceParam1Two", "SliceParam1Three"},
		AliasArrayParam:  [3]string{"AliasArray1One", "AliasArray1Two", "AliasArray1Three"},
		AliasSliceParam:  []string{"AliasSlice1One", "AliasSlice1Two", "AliasSlice1Three"},
		AliasStringParam: "AliasString1",
		AliasIntParam:    1,
	},
	&QueryFixture{
		ID:               kallax.NewULID(),
		Foo:              "Foo2",
		StringProperty:   "StringProperty2",
		Integer:          2,
		Integer64:        2,
		Float32:          2,
		Boolean:          true,
		ArrayParam:       [3]string{"ArrayParm2One", "ArrayParm2Two", "ArrayParm2Three"},
		SliceParam:       []string{"SliceParam2One", "SliceParam2Two", "SliceParam2Three"},
		AliasArrayParam:  [3]string{"AliasArray2One", "AliasArray2Two", "AliasArray2Three"},
		AliasSliceParam:  []string{"AliasSlice2One", "AliasSlice2Two", "AliasSlice2Three"},
		AliasStringParam: "AliasString2",
		AliasIntParam:    2,
	},
}

func resetQueryFixtures() {
	for i, fixture := range queryFixtures {
		queryFixtures[i] = &QueryFixture{
			ID:               fixture.ID,
			Foo:              fixture.Foo,
			StringProperty:   fixture.StringProperty,
			Integer:          fixture.Integer,
			Integer64:        fixture.Integer64,
			Float32:          fixture.Float32,
			Boolean:          fixture.Boolean,
			ArrayParam:       fixture.ArrayParam,
			SliceParam:       fixture.SliceParam,
			AliasArrayParam:  fixture.AliasArrayParam,
			AliasSliceParam:  fixture.AliasSliceParam,
			AliasStringParam: fixture.AliasStringParam,
			AliasIntParam:    fixture.AliasIntParam,
		}
	}
}
