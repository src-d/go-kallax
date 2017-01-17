package behaviours

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestIdentificator(t *testing.T) {
	suite.Run(t, new(IdentificatorSuite))
}

type IdentificatorSuite struct {
	suite.Suite
}

func (s *IdentificatorSuite) TestUniqueness() {
	item1 := Identificator{}
	item2 := Identificator{}
	s.True(item1.ID.IsEmpty())
	s.True(item2.ID.IsEmpty())
	ID1 := item1.Identify()
	ID2 := item2.Identify()
	s.NotEqual(ID1, ID2)
}

func (s *IdentificatorSuite) TestIdentifyWorksOnlyOnce() {
	item1 := Identificator{}
	s.True(item1.ID.IsEmpty())
	ID1a := item1.Identify()
	ID1b := item1.Identify()
	s.Equal(ID1a, ID1b)
}

func (s *IdentificatorSuite) TestIDCreationBeforeInsertion() {
	item1 := Identificator{}
	s.True(item1.ID.IsEmpty())
	error := item1.BeforeInsert()
	s.Nil(error)
	s.False(item1.ID.IsEmpty())
}