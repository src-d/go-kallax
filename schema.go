package kallax

type Schema interface {
	Alias() string
	Table() string
	Identifier() string
	Columns() []string
}
