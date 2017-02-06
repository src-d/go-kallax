Kallax - PostgreSQL ORM for Go
=============================

[![GoDoc](https://godoc.org/github.com/src-d/go-kallax?status.svg)](https://godoc.org/github.com/src-d/go-kallax) [![Build Status](https://travis-ci.org/src-d/go-kallax.svg?branch=master)](https://travis-ci.org/src-d/go-kallax) [![codecov](https://codecov.io/gh/src-d/go-kallax/branch/master/graph/badge.svg)](https://codecov.io/gh/src-d/go-kallax) [![Go Report Card](https://goreportcard.com/badge/github.com/src-d/go-kallax)](https://goreportcard.com/report/github.com/src-d/go-kallax) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Installation
------------

The recommended way to install `kallax` is:

```
go get -u github.com/src-d/kallax/...
```

> *kallax* includes a binary tool used by [go generate](http://blog.golang.org/generate),
please be sure that `$GOPATH/bin` is on your `$PATH`

### Running tests

For obvious reasons, an instance of PostgreSQL is required to run the tests of this package.

By default, it assumes that an instance exists at `0.0.0.0:5432` with an user, password and database name all equal to `testing`.

If that is not the case you can set the following environment variables:

- `DBNAME`: name of the database
- `DBUSER`: database user
- `DBPASS`: database user password

License
-------

MIT, see [LICENSE](LICENSE)
