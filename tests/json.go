package tests

import kallax "github.com/src-d/go-kallax"

type JSONModel struct {
	kallax.Model `table:"jsons"`
	Foo          string
	Bar          *Bar
	Baz          map[string]interface{}
}

type Bar struct {
	Qux []Qux
	Mux string
}

type Qux struct {
	Schnooga string
	Balooga  int
	Boo      float64
}

func newJSONModel() *JSONModel {
	return &JSONModel{Baz: make(map[string]interface{})}
}
