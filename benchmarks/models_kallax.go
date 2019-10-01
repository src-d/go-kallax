package benchmark

import kallax "github.com/zbyte/go-kallax"

type Person struct {
	kallax.Model `table:"people"`
	ID           int64 `pk:"autoincr"`
	Name         string
	Pets         []*Pet
}

type Pet struct {
	kallax.Model `table:"pets"`
	ID           int64 `pk:"autoincr"`
	Name         string
	Kind         PetKind
}

type PetKind string

const (
	Cat  PetKind = "cat"
	Dog  PetKind = "dog"
	Fish PetKind = "fish"
)
