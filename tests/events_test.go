package tests

import "errors"

func (s *CommonSuite) TestEventsInsert() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	s.Nil(err)
	s.Equal(doc.Checks, map[string]bool{
		"BeforeInsert": true,
		"AfterInsert":  true,
	})
}

func (s *CommonSuite) TestEventsUpdate() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	s.Nil(err)

	doc.Checks = make(map[string]bool, 0)
	updatedRows, err := store.Update(doc)
	s.Nil(err)
	s.True(updatedRows > 0)
	s.Equal(doc.Checks, map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	})
}

func (s *CommonSuite) TestEventsUpdateError() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	doc.MustFailAfter = errors.New("after")
	updatedRows, err := store.Update(doc)
	s.True(updatedRows == 0)
	s.Equal(err, doc.MustFailAfter)

	doc.MustFailBefore = errors.New("before")
	updatedRows, err = store.Update(doc)
	s.True(updatedRows == 0)
	s.Equal(err, doc.MustFailBefore)
}

func (s *CommonSuite) TestEventsSaveOnInsert() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	updated, err := store.Save(doc)
	s.Nil(err)
	s.Equal(updated, false)
	s.Equal(doc.Checks, map[string]bool{
		"BeforeInsert": true,
		"AfterInsert":  true,
	})
}

func (s *CommonSuite) TestEventsSaveOnUpdate() {
	store := NewEventsFixtureStore(s.db)

	doc := NewEventsFixture()
	err := store.Insert(doc)
	doc.Checks = make(map[string]bool, 0)

	updated, err := store.Save(doc)
	s.Nil(err)
	s.Equal(updated, true)
	s.Equal(doc.Checks, map[string]bool{
		"BeforeUpdate": true,
		"AfterUpdate":  true,
	})
}

func (s *CommonSuite) TestEventsSaveInsert() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	s.Nil(err)
	s.Equal(doc.Checks, map[string]bool{
		"BeforeSave": true,
		"AfterSave":  true,
	})
}

func (s *CommonSuite) TestEventsSaveUpdate() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	s.Nil(err)

	doc.Checks = make(map[string]bool, 0)
	updatedRows, err := store.Update(doc)
	s.Nil(err)
	s.True(updatedRows > 0)
	s.Equal(doc.Checks, map[string]bool{
		"BeforeSave": true,
		"AfterSave":  true,
	})
}

func (s *CommonSuite) TestEventsSaveSave() {
	store := NewEventsSaveFixtureStore(s.db)

	doc := NewEventsSaveFixture()
	err := store.Insert(doc)
	doc.Checks = map[string]bool{"AfterInsert": true}

	updated, err := store.Save(doc)
	s.Nil(err)
	s.Equal(updated, true)
	s.Equal(doc.Checks, map[string]bool{
		"AfterInsert": true,
		"BeforeSave":  true,
		"AfterSave":   true,
	})
}
