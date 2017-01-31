package kallax

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
}

func (s *ModelSuite) TestID_IsEmpty() {
	var id ID
	s.True(id.IsEmpty())
	id = NewID()
	s.False(id.IsEmpty())
}

func (s *ModelSuite) TestID_ThreeNewIDsAreDifferent() {
	id1 := NewID()
	id2 := NewID()
	id3 := NewID()
	s.NotEqual(id1, id2)
	s.NotEqual(id1, id3)
	s.NotEqual(id2, id3)

	s.True(id1 == id1)
	s.False(id1 == id2)
}
