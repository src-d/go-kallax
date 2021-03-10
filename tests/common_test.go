package tests

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"

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
	db      *sql.DB
	schemas []string
	tables  []string

	// the count of opened connections.
	// this will be set in the SetupTest function.
	openConnectionsBeforeTest int
}

func NewBaseSuite(schemas []string, tables ...string) BaseTestSuite {
	return BaseTestSuite{
		schemas: schemas,
		tables:  tables,
	}
}

func (s *BaseTestSuite) SetupSuite() {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(connectionString, user, password, host, database),
	)

	if err != nil {
		panic(fmt.Sprintf("It was unable to connect to the DB.\n%s\n", err))
	}

	// set all connections will be closed immediately.
	// this is required to check connections are leaked or not.
	// because database/sql keep connection in the pool by default.
	db.SetMaxIdleConns(0)

	s.db = db
}

func (s *BaseTestSuite) TearDownSuite() {
	s.db.Close()
}

func (s *BaseTestSuite) SetupTest() {
	// save current open connection count for detecting that connection was leaked while a test.
	s.openConnectionsBeforeTest = s.db.Stats().OpenConnections

	if len(s.tables) == 0 {
		return
	}

	s.QuerySucceed(s.schemas...)
}

func (s *BaseTestSuite) TearDownTest() {
	openConnections := s.db.Stats().OpenConnections
	leakedConnections := openConnections - s.openConnectionsBeforeTest
	if leakedConnections > 0 {
		s.Fail(fmt.Sprintf("%d database connections were leaked", leakedConnections))
	}

	if len(s.tables) == 0 {
		return
	}
	var queries []string
	for _, t := range s.tables {
		queries = append(queries, fmt.Sprintf("DROP TABLE IF EXISTS %s", t))
	}
	s.QuerySucceed(queries...)
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

func (s *BaseTestSuite) resultOrError(res interface{}, err error) bool {
	if !reflect.ValueOf(res).Elem().IsValid() {
		res = nil
	}

	if err == nil && res == nil {
		s.Fail("FindOne should return an error or a document, but nothing was returned")
		return false
	}

	if err != nil && res != nil {
		s.Fail("FindOne should return only an error or a document, but it was returned both")
		return false
	}

	return true
}

func (s *BaseTestSuite) resultsOrError(res interface{}, err error) bool {
	if reflect.ValueOf(res).Kind() != reflect.Slice {
		panic("resultsOrError expects res is a slice")
	}

	if err == nil && res == nil {
		s.Fail("FindAll should return an error or a documents, but nothing was returned")
		return false
	}

	if err != nil && res != nil {
		s.Fail("FindAll should return only an error or a documents, but it was returned both")
		return false
	}

	return true
}

func envOrDefault(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}
