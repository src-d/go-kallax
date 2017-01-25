package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionSuite))
}

type ConnectionSuite struct {
	BaseTestSuite
}

func (s *ConnectionSuite) TestConnection() {
	s.QuerySucceed(
		`CREATE TABLE testing (id uuid primary key)`,
		`DROP TABLE testing`,
		`DROP TABLE IF EXISTS testing`,
	)
	s.QueryFails(`DROP TABLE _THIS_TABLE_DOES_NOT_EXIST`)
}
