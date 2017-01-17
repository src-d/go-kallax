package common

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

func (s *IdentificatorSuite) TestID_IsEmpty() {
	var id ID
	s.True(id.IsEmpty())

	id = NewID()
	s.False(id.IsEmpty())
}

func (s *IdentificatorSuite) TestID_ThreeNewIDsAreDifferent() {
	id1 := NewID()
	id2 := NewID()
	id3 := NewID()

	s.NotEqual(id1, id2)
	s.NotEqual(id1, id3)
	s.NotEqual(id2, id3)
}
