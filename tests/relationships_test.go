package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RelationshipsSuite struct {
	BaseTestSuite
}

func TestRelationships(t *testing.T) {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS persons (
			id serial primary key,
			name text
		)`,
		`CREATE TABLE IF NOT EXISTS brands (
			id uuid primary key,
			name text
		)`,
		`CREATE TABLE IF NOT EXISTS cars (
			id uuid primary key,
			model_name text,
			owner_id integer references persons(id),
			brand_id uuid references brands(id)
		)`,
		`CREATE TABLE IF NOT EXISTS pets (
			id uuid primary key,
			name text,
			kind text,
			owner_id integer references persons(id)
		)`,
	}
	suite.Run(t, &RelationshipsSuite{NewBaseSuite(schemas, "cars", "pets", "persons")})
}

func (s *RelationshipsSuite) TestInsertFind() {
	require := s.Require()
	p := NewPerson("Dolan")
	car := NewCar("Tesla Model S", p)
	cat := NewPet("Garfield", "cat", p)
	dog := NewPet("Oddie", "dog", p)

	store := NewPersonStore(s.db)
	require.NoError(store.Insert(p))

	pers := s.getPerson()
	s.assertPerson(p.Name, pers, car, cat, dog)
}

func (s *RelationshipsSuite) TestInsertFindRemove() {
	p := NewPerson("Dolan")
	car := NewCar("Tesla Model S", p)
	cat := NewPet("Garfield", "cat", p)
	dog := NewPet("Oddie", "dog", p)
	reptar := NewPet("Reptar", "dinosaur", p)

	store := NewPersonStore(s.db)
	s.NoError(store.Insert(p))

	pers := s.getPerson()
	s.assertPerson(p.Name, pers, car, cat, dog, reptar)

	s.NoError(store.RemovePets(pers, dog))
	pers = s.getPerson()
	s.assertPerson(p.Name, pers, car, cat, reptar)

	s.NoError(store.RemovePets(pers))
	s.NoError(store.RemoveCar(pers))
	pers = s.getPerson()
	s.assertPerson(p.Name, pers, nil)
}

func (s *RelationshipsSuite) TestInsertFindUpdate() {
	p := NewPerson("Dolan")
	car := NewCar("Tesla Model S", p)
	cat := NewPet("Garfield", "cat", p)
	dog := NewPet("Oddie", "dog", p)

	store := NewPersonStore(s.db)
	s.NoError(store.Insert(p))

	pers := s.getPerson()
	s.assertPerson(p.Name, pers, car, cat, dog)

	pony := NewPet("Sparkling Twilight", "pony", pers)
	_, err := store.Save(pers)
	s.NoError(err)

	pers = s.getPerson()
	s.assertPerson(p.Name, pers, car, cat, dog, pony)
}

func (s *RelationshipsSuite) TestEvents() {
	p := NewPerson("Dolan")
	car := NewCar("Tesla Model S", p)
	cat := NewPet("Garfield", "cat", p)
	dog := NewPet("Oddie", "dog", p)
	reptar := NewPet("Reptar", "dinosaur", p)

	store := NewPersonStore(s.db)
	s.NoError(store.Insert(p))

	s.assertEvents(p.events, "BeforeSave", "AfterSave")
	s.assertEvents(car.events, "BeforeSave", "AfterSave")
	s.assertEvents(cat.events, "BeforeSave", "AfterSave")
	s.assertEvents(dog.events, "BeforeSave", "AfterSave")
	s.assertEvents(reptar.events, "BeforeSave", "AfterSave")

	s.NoError(store.RemovePets(p, dog))
	s.assertNoEvents(cat.events, "BeforeDelete", "AfterDelete")
	s.assertNoEvents(reptar.events, "BeforeDelete", "AfterDelete")
	s.assertEvents(dog.events, "BeforeDelete", "AfterDelete")

	s.NoError(store.RemovePets(p))
	s.assertEvents(reptar.events, "BeforeDelete", "AfterDelete")
	s.assertEvents(cat.events, "BeforeDelete", "AfterDelete")

	s.NoError(store.RemoveCar(p))
	s.assertEvents(car.events, "BeforeDelete", "AfterDelete")
}

func (s *RelationshipsSuite) TestSaveWithInverse() {
	p := NewPerson("Foo")
	car := NewCar("Bar", p)

	store := NewCarStore(s.db)
	s.NoError(store.Insert(car))

	s.NotNil(s.getPerson())
}

func (s *RelationshipsSuite) TestSaveRelations() {
	p := NewPerson("Musk")
	brand := makeBrand("Tesla")
	car := newBrandedCar("Model S", p, brand)

	store := NewCarStore(s.db).Debug()
	_, err := store.Save(car)
	s.NoError(err)

	car, err = store.FindOne(NewCarQuery().FindByID(car.ID).WithBrand())
	s.NoError(err)
	s.NotNil(car)
	s.NotNil(car.Brand)

	pStore := NewPersonStore(s.db).Debug()

	p.Name = "Elon"
	_, err = pStore.Save(p)
	s.NoError(err)
	s.NotNil(p.Car)
	s.NotNil(p.Car.Brand)

	car, err = store.FindOne(NewCarQuery().FindByID(car.ID).WithBrand())
	s.NoError(err)
	s.NotNil(car)
	s.NotNil(car.Brand)
}

func (s *RelationshipsSuite) assertEvents(evs map[string]int, events ...string) {
	for _, e := range events {
		s.Equal(1, evs[e])
	}
}

func (s *RelationshipsSuite) assertNoEvents(evs map[string]int, events ...string) {
	for _, e := range events {
		s.Equal(0, evs[e])
	}
}

func (s *RelationshipsSuite) assertPerson(name string, pers *Person, car *Car, pets ...*Pet) {
	require := s.Require()
	require.False(pers.GetID().IsEmpty(), "ID should not be empty")
	require.Equal(name, pers.Name)
	pers.events = nil
	require.Len(pers.Pets, len(pets))

	// Owner are set to nil to be able to deep equal in the tests.
	// Records coming from relationships don't have their relationships
	// populated, so it will always be nil.
	// Same with events.
	var petList = make([]*Pet, len(pets))
	for i, pet := range pets {
		p := *pet
		require.False(p.GetID().IsEmpty(), "ID should not be empty")
		p.Owner = nil
		p.events = nil
		petList[i] = &p
	}

	var c Car
	if car == nil {
		require.Nil(pers.Car)
	} else {
		c = *car
		require.False(c.GetID().IsEmpty(), "ID should not be empty")
		c.Owner = nil
		c.events = nil
		require.Equal(&c, pers.Car)
	}
	for i, p := range petList {
		require.Equal(p, pers.Pets[i])
	}
}

func (s *RelationshipsSuite) getPerson() *Person {
	require := s.Require()
	q := NewPersonQuery().
		WithCar().
		WithPets(nil)
	pers, err := NewPersonStore(s.db).FindOne(q)
	require.NoError(err)
	require.NotNil(pers)

	return pers
}
