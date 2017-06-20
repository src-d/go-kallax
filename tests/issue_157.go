package tests

import kallax "gopkg.in/src-d/go-kallax.v1"

type Foo struct {
	kallax.Model `table:"foos"`
	ID           int64 `pk:""`
}
