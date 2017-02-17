package tests

import kallax "github.com/src-d/go-kallax"

type Car struct {
	kallax.Model `table:"cars"`
	ID           kallax.ULID `pk:""`
	Owner        *Person     `fk:"owner_id,inverse"`
	ModelName    string
	events       map[string]int
}

func (c *Car) ensureMapInitialized() {
	if c.events == nil {
		c.events = make(map[string]int)
	}
}

func (c *Car) BeforeSave() error {
	c.ensureMapInitialized()
	c.events["BeforeSave"]++
	return nil
}

func (c *Car) AfterSave() error {
	c.ensureMapInitialized()
	c.events["AfterSave"]++
	return nil
}

func (c *Car) BeforeDelete() error {
	c.ensureMapInitialized()
	c.events["BeforeDelete"]++
	return nil
}

func (c *Car) AfterDelete() error {
	c.ensureMapInitialized()
	c.events["AfterDelete"]++
	return nil
}

type Person struct {
	kallax.Model `table:"persons"`
	ID           int64 `pk:"autoincr"`
	Name         string
	Pets         []*Pet `fk:"owner_id"`
	Car          *Car   `fk:"owner_id"`
	events       map[string]int
}

func (c *Person) ensureMapInitialized() {
	if c.events == nil {
		c.events = make(map[string]int)
	}
}

func (c *Person) BeforeSave() error {
	c.ensureMapInitialized()
	c.events["BeforeSave"]++
	return nil
}

func (c *Person) AfterSave() error {
	c.ensureMapInitialized()
	c.events["AfterSave"]++
	return nil
}

func (c *Person) BeforeDelete() error {
	c.ensureMapInitialized()
	c.events["BeforeDelete"]++
	return nil
}

func (c *Person) AfterDelete() error {
	c.ensureMapInitialized()
	c.events["AfterDelete"]++
	return nil
}

type Pet struct {
	kallax.Model `table:"pets"`
	ID           kallax.ULID `pk:""`
	Name         string
	Kind         string
	Owner        *Person `fk:"owner_id,inverse"`
	events       map[string]int
}

func (c *Pet) ensureMapInitialized() {
	if c.events == nil {
		c.events = make(map[string]int)
	}
}

func (c *Pet) BeforeSave() error {
	c.ensureMapInitialized()
	c.events["BeforeSave"]++
	return nil
}

func (c *Pet) AfterSave() error {
	c.ensureMapInitialized()
	c.events["AfterSave"]++
	return nil
}

func (c *Pet) BeforeDelete() error {
	c.ensureMapInitialized()
	c.events["BeforeDelete"]++
	return nil
}

func (c *Pet) AfterDelete() error {
	c.ensureMapInitialized()
	c.events["AfterDelete"]++
	return nil
}

func newPet(name, kind string, owner *Person) *Pet {
	pet := &Pet{ID: kallax.NewULID(), Name: name, Kind: kind, Owner: owner}
	owner.Pets = append(owner.Pets, pet)
	return pet
}

func newPerson(name string) *Person {
	return &Person{Name: name}
}

func newCar(model string, owner *Person) *Car {
	car := &Car{ID: kallax.NewULID(), ModelName: model, Owner: owner}
	owner.Car = car
	return car
}
