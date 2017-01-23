Kallax - PostgreSQL ORM for Go
=============================

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
