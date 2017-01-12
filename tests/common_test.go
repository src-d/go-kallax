package tests

import (
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
)

const (
	testMongoHost = "127.0.0.1:27017"
	testDatabase  = "storable-test"
)

func Test(t *testing.T) { TestingT(t) }

type MongoSuite struct {
	db *mgo.Database
}

var _ = Suite(&MongoSuite{})

func (s *MongoSuite) SetUpTest(c *C) {
	conn, _ := mgo.Dial(testMongoHost)
	s.db = conn.DB(testDatabase)
}

func (s *MongoSuite) TearDownTest(c *C) {
	s.db.DropDatabase()
}
