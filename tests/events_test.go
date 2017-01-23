package tests

import (
	"database/sql"
	"errors"

	"github.com/stretchr/testify/suite"
)

type EventsSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *EventsSuite) TestEventsInsert() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	s.Nil(err)
	s.Equal(map[string]bool{
		"BeforeInsert": true,
		"AfterInsert":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsUpdate() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	s.Nil(err)

	doc.Checks = make(map[string]bool, 0)
	updatedRows, err := store.Update(doc)
	s.Nil(err)
	s.True(updatedRows > 0)
	s.Equal(map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsUpdateError() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	doc.MustFailAfter = errors.New("after")
	updatedRows, err := store.Update(doc)
	s.True(updatedRows == 0)
	s.Equal(doc.MustFailAfter, err)

	doc.MustFailBefore = errors.New("before")
	updatedRows, err = store.Update(doc)
	s.True(updatedRows == 0)
	s.Equal(doc.MustFailBefore, err)
}

func (s *EventsSuite) TestEventsSaveOnInsert() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	updated, err := store.Save(doc)
	s.Nil(err)
	s.False(updated)
	s.Equal(map[string]bool{
		"BeforeInsert": true,
		"AfterInsert":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsSaveOnUpdate() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	updated, err := store.Save(doc)
	s.Nil(err)
	s.True(updated)
	s.Equal(map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsSaveInsert() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	s.Nil(err)
	s.Equal(map[string]bool{
		"BeforeSave": true,
		"AfterSave":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsSaveUpdate() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	s.Nil(err)

	doc.Checks = make(map[string]bool, 0)
	updatedRows, err := store.Update(doc)
	s.Nil(err)
	s.True(updatedRows > 0)
	s.Equal(map[string]bool{
		"BeforeSave": true,
		"AfterSave":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsSaveSave() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	doc.Checks = map[string]bool{"AfterInsert": true}

	updated, err := store.Save(doc)
	s.Nil(err)
	s.True(updated)
	s.Equal(map[string]bool{
		"AfterInsert": true,
		"BeforeSave":  true,
		"AfterSave":   true,
	}, doc.Checks)
}
