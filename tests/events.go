package tests

import "github.com/src-d/go-kallax"

type EventsFixture struct {
	kallax.Document `bson:",inline" collection:"event"`
	Checks          map[string]bool
	MustFailBefore  error
	MustFailAfter   error
}

func newEventsFixture() *EventsFixture {
	return &EventsFixture{
		Checks: make(map[string]bool, 0),
	}
}

func (s *EventsFixtureStore) BeforeInsert(doc *EventsFixture) error {
	if doc.MustFailBefore != nil {
		return doc.MustFailBefore
	}

	doc.Checks["BeforeInsert"] = true
	return nil
}

func (s *EventsFixtureStore) AfterInsert(doc *EventsFixture) error {
	if doc.MustFailAfter != nil {
		return doc.MustFailAfter
	}

	doc.Checks["AfterInsert"] = true
	return nil
}

func (s *EventsFixtureStore) BeforeUpdate(doc *EventsFixture) error {
	if doc.MustFailBefore != nil {
		return doc.MustFailBefore
	}

	doc.Checks["BeforeUpdate"] = true
	return nil
}

func (s *EventsFixtureStore) AfterUpdate(doc *EventsFixture) error {
	if doc.MustFailAfter != nil {
		return doc.MustFailAfter
	}

	doc.Checks["AfterUpdate"] = true
	return nil
}

type EventsSaveFixture struct {
	kallax.Document `bson:",inline" collection:"event"`
	Checks          map[string]bool
	MustFailBefore  error
	MustFailAfter   error
}

func newEventsSaveFixture() *EventsSaveFixture {
	return &EventsSaveFixture{
		Checks: make(map[string]bool, 0),
	}
}

func (s *EventsSaveFixtureStore) BeforeSave(doc *EventsSaveFixture) error {
	if doc.MustFailBefore != nil {
		return doc.MustFailBefore
	}

	doc.Checks["BeforeSave"] = true
	return nil
}

func (s *EventsSaveFixtureStore) AfterSave(doc *EventsSaveFixture) error {
	if doc.MustFailAfter != nil {
		return doc.MustFailAfter
	}

	doc.Checks["AfterSave"] = true
	return nil
}
