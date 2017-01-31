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
		`CREATE TABLE persons (
			id uuid primary key,
			name text
		)`,
		`CREATE TABLE cars (
			id uuid primary key,
			model_name text,
			owner_id uuid references persons(id)
		)`,
		`CREATE TABLE pets (
			id uuid primary key,
			name text,
			kind text,
			owner_id uuid references persons(id)
		)`,
	}
	suite.Run(t, &RelationshipsSuite{NewBaseSuite(schemas, "cars", "pets", "persons")})
}

func (s *RelationshipsSuite) TestInsertFind() {
	p := NewPerson("Dolan")
	car := NewCar("Tesla Model S", p)
	cat := NewPet("Garfield", "cat", p)
	dog := NewPet("Oddie", "dog", p)

	store := NewPersonStore(s.db)
	s.NoError(store.Insert(p))

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

func (s *RelationshipsSuite) assertPerson(name string, pers *Person, car *Car, pets ...*Pet) {
	s.Equal(name, pers.Name)
	s.Len(pers.Pets, len(pets))

	// Owner are set to nil to be able to deep equal in the tests.
	// Records coming from relationships don't have their relationships
	// case \"foo_id\":\nreturn kallax.VirtualColumn(\"foo_id\", r), nilpopulated, so it will always be nil.
	var petList = make([]*Pet, len(pets))
	for i, pet := range pets {
		p := *pet
		p.Owner = nil
		petList[i] = &p
	}

	var c Car
	if car == nil {
		s.Nil(pers.Car)
	} else {
		c = *car
		c.Owner = nil
		s.Equal(&c, pers.Car)
	}
	for i, p := range petList {
		s.Equal(p, pers.Pets[i])
	}
}

func (s *RelationshipsSuite) getPerson() *Person {
	q := NewPersonQuery().
		WithCar().
		WithPets(nil)
	pers, err := NewPersonStore(s.db).FindOne(q)
	s.NoError(err)
	s.NotNil(pers)

	return pers
}
