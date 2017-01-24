package tests

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	host     = "0.0.0.0:5432"
	database = "testing"
	user     = "testing"
	password = "testing"
)

func TestHijackSuite(t *testing.T) {
	suite.Run(t, new(CommonSuite))
}

type CommonSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *CommonSuite) SetupSuite() {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, database),
	)
	s.Nil(err)
	s.NotNil(db)
	s.db = db

	res, err := s.db.Exec(`DROP TABLE IF EXISTS testing`)
	s.NotNil(res)
	s.Nil(err)

	res, err = s.db.Exec(`CREATE TABLE testing (id uuid primary key)`)
	s.NotNil(res)
	s.Nil(err)
}

func (s *CommonSuite) TearDownSuite() {
	res, err := s.db.Exec("DROP TABLE testing")
	s.NotNil(res)
	s.Nil(err)

	res, err = s.db.Exec("DROP TABLE _THIS_TABLE_DOES_NOT_EXIST")
	s.Nil(res)
	s.NotNil(err)
}
