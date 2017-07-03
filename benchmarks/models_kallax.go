package benchmark

import kallax "gopkg.in/src-d/go-kallax.v1"

//go:generate kallax gen

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
