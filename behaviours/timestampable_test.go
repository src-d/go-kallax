package behaviours

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestTimestampable(t *testing.T) {
	suite.Run(t, new(TimestampableSuite))
}

type TimestampableSuite struct {
	suite.Suite
}

func (s *TimestampableSuite) TestTimestamp() {
	item := TimestampDates{}
	s.True(item.CreatedAt.IsZero())
	s.True(item.UpdatedAt.IsZero())
	item.Timestamp()
	createdAt := item.CreatedAt
	updatedAt := item.CreatedAt
	s.False(createdAt.IsZero())
	s.False(updatedAt.IsZero())
	item.Timestamp()
	s.Equal(createdAt, item.CreatedAt)
	s.NotEqual(updatedAt, item.UpdatedAt)
}

func (s *TimestampableSuite) TestTimestampBeforePersist() {
	item := TimestampDates{}
	error := item.BeforePersist()
	s.Nil(error)
	s.False(item.CreatedAt.IsZero())
	s.False(item.UpdatedAt.IsZero()) 
}
