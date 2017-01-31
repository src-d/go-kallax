package tests

import "github.com/src-d/go-kallax"

type EventsFixture struct {
	kallax.Model   `table:"event"`
	Checks         map[string]bool
	MustFailBefore error
	MustFailAfter  error
}

func newEventsFixture() *EventsFixture {
	return &EventsFixture{
		Checks: make(map[string]bool, 0),
	}
}

func (s *EventsFixture) BeforeInsert() error {
	if s.MustFailBefore != nil {
		return s.MustFailBefore
	}

	s.Checks["BeforeInsert"] = true
	return nil
}

func (s *EventsFixture) AfterInsert() error {
	if s.MustFailAfter != nil {
		return s.MustFailAfter
	}

	s.Checks["AfterInsert"] = true
	return nil
}

func (s *EventsFixture) BeforeUpdate() error {
	if s.MustFailBefore != nil {
		return s.MustFailBefore
	}

	s.Checks["BeforeUpdate"] = true
	return nil
}

func (s *EventsFixture) AfterUpdate() error {
	if s.MustFailAfter != nil {
		return s.MustFailAfter
	}

	s.Checks["AfterUpdate"] = true
	return nil
}

type EventsSaveFixture struct {
	kallax.Model   `table:"event"`
	Checks         map[string]bool
	MustFailBefore error
	MustFailAfter  error
}

func newEventsSaveFixture() *EventsSaveFixture {
	return &EventsSaveFixture{
		Checks: make(map[string]bool, 0),
	}
}

func (s *EventsSaveFixture) BeforeSave() error {
	if s.MustFailBefore != nil {
		return s.MustFailBefore
	}

	s.Checks["BeforeSave"] = true
	return nil
}

func (s *EventsSaveFixture) AfterSave() error {
	if s.MustFailAfter != nil {
		return s.MustFailAfter
	}

	s.Checks["AfterSave"] = true
	return nil
}
