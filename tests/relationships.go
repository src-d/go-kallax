package tests

import kallax "github.com/src-d/go-kallax"

type Car struct {
	kallax.Model `table:"cars"`
	Owner        *Person `fk:"owner_id,inverse"`
	ModelName    string
}

type Person struct {
	kallax.Model `table:"persons"`
	Name         string
	Pets         []*Pet `fk:"owner_id"`
	Car          *Car   `fk:"owner_id"`
}

type Pet struct {
	kallax.Model `table:"pets"`
	Name         string
	Kind         string
	Owner        *Person `fk:"owner_id,inverse"`
}

func newPet(name, kind string, owner *Person) *Pet {
	pet := &Pet{Name: name, Kind: kind, Owner: owner}
	owner.Pets = append(owner.Pets, pet)
	return pet
}

func newPerson(name string) *Person {
	return &Person{Name: name}
}

func newCar(model string, owner *Person) *Car {
	car := &Car{ModelName: model, Owner: owner}
	owner.Car = car
	return car
}
