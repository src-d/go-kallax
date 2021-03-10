module github.com/networkteam/go-kallax

go 1.13

require (
	github.com/Masterminds/squirrel v1.4.0
	github.com/fatih/color v1.7.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0
	github.com/lib/pq v1.8.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/oklog/ulid v1.3.1
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.22.1
	golang.org/x/tools v0.0.0-20200625173320-e31c80b82c03
)

replace github.com/golang-migrate/migrate/v4 => github.com/networkteam/migrate/v4 v4.15.0
