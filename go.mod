module github.com/loyalguru/go-kallax

go 1.14

require (
	github.com/Masterminds/squirrel v1.5.0
	github.com/fatih/color v1.10.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang-migrate/migrate v3.5.4+incompatible // indirect
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0
	github.com/lib/pq v1.10.0
	github.com/oklog/ulid v1.3.1
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.5
	golang.org/x/tools v0.1.0
)

//replace github.com/golang-migrate/migrate/v4 => github.com/networkteam/migrate/v4 v4.15.0
