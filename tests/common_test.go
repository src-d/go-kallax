package tests

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/stretchr/testify/suite"
)

var (
	connectionString = "postgres://%s:%s@%s/%s?sslmode=disable"
	host             = envOrDefault("DBHOST", "0.0.0.0:5432")
	database         = envOrDefault("DBNAME", "testing")
	user             = envOrDefault("DBUSER", "testing")
	password         = envOrDefault("DBPASS", "testing")
)

type BaseTestSuite struct {
	suite.Suite
	db          *sql.DB
	initQueries []string
}

func (s *BaseTestSuite) SetupSuite() {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(connectionString, user, password, host, database),
	)

	if err != nil {
		panic(fmt.Sprintf("It was unable to connect to the DB.\n%s\n", err))
	}

	s.db = db

	if !s.resetSchema() {
		s.Require().FailNow("Tests can not be run because database Schema can not be accessed")
	}
}

func (s *BaseTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *BaseTestSuite) SetupTest() {
	if len(s.initQueries) > 0 {
		s.QuerySucceed(s.initQueries...)
	}
}

func (s *BaseTestSuite) TearDownTest() {
	s.resetSchema()
}

func (s *BaseTestSuite) QuerySucceed(queries ...string) bool {
	success := true
	for _, query := range queries {
		res, err := s.db.Exec(query)
		assert1 := s.NotNil(res, "Resulset should not be empty")
		assert2 := s.Nil(err, fmt.Sprintf("%s\nshould succeed but it failed.\n%s\n", query, err))
		if !assert1 || !assert2 {
			success = false
		}
	}

	return success
}

func (s *BaseTestSuite) QueryFails(queries ...string) bool {
	success := true
	for _, query := range queries {
		res, err := s.db.Exec(query)
		assert1 := s.Nil(res, "Resulset should be empty but it was not")
		assert2 := s.NotNil(err, fmt.Sprintf("%s\nshould fail but it succeed", query))
		if !assert1 || !assert2 {
			success = false
		}
	}

	return success
}

func (s *BaseTestSuite) resetSchema() bool {
	return s.QuerySucceed(
		fmt.Sprintf(`DROP SCHEMA %s CASCADE;`, database),
		fmt.Sprintf(`CREATE SCHEMA %s;`, database),
	)
}

func envOrDefault(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}
