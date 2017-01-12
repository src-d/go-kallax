package operators

import (
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type OperatorsSuite struct{}

var _ = check.Suite(&OperatorsSuite{})

var (
	Foo = FieldExample("foo")
)

type FieldExample string

func (f FieldExample) String() string {
	return string(f)
}
