package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EventsSuite struct {
	BaseTestSuite
}

func TestEventsSuite(t *testing.T) {
	schema := []string{
		`CREATE TABLE event (
			id uuid primary key,
			checks JSON,
			must_fail_before JSON,
			must_fail_after JSON
		)`,
	}
	suite.Run(t, &EventsSuite{NewBaseSuite(schema, "event")})
}

type eventsCheck map[string]bool

func assertEventsPassed(s *EventsSuite, expected eventsCheck, received eventsCheck) {
	for expectedEvent, expectedSign := range expected {
		receivedSign, ok := received[expectedEvent]
		if s.True(ok, fmt.Sprintf("Event '%s' was not received", expectedEvent)) {
			s.Equal(expectedSign, receivedSign, expectedEvent)
		}
	}

	s.Equal(len(expected), len(received))
}

func (s *EventsSuite) TestEventsInsert() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	s.Nil(err)
	assertEventsPassed(s, map[string]bool{
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
	assertEventsPassed(s, map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsUpdateError() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	doc.MustFailAfter = errors.New("kallax: after")
	updatedRows, err := store.Update(doc)
	s.Equal(int64(0), updatedRows)
	s.Equal(doc.MustFailAfter, err)

	doc.MustFailBefore = errors.New("kallax: before")
	updatedRows, err = store.Update(doc)
	s.Equal(int64(0), updatedRows)
	s.Equal(doc.MustFailBefore, err)
}

func (s *EventsSuite) TestEventsSaveOnInsert() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	updated, err := store.Save(doc)
	s.Nil(err)
	s.False(updated)
	assertEventsPassed(s, map[string]bool{
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
	assertEventsPassed(s, map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	}, doc.Checks)
}

func (s *EventsSuite) TestEventsSaveInsert() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	s.Nil(err)
	assertEventsPassed(s, map[string]bool{
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
	assertEventsPassed(s, map[string]bool{
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
	assertEventsPassed(s, map[string]bool{
		"AfterInsert": true,
		"BeforeSave":  true,
		"AfterSave":   true,
	}, doc.Checks)
}
