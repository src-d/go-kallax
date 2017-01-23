Kallax - PostgreSQL ORM for Go
=============================

Installation
------------

The recommended way to install `kallax` is:

```
go get -u github.com/src-d/kallax/...
go install github.com/src-d/go-kallax/generator/cli/kallax
```

> *kallax* includes a binary tool used by [go generate](http://blog.golang.org/generate),
please be sure that `$GOPATH/bin` is on your `$PATH`

use
---
```
kallax gen -i=PATH_TO_THE_PACKAGE
```
Where `PATH_TO_THE_PACKAGE` is the package directory containing the models to parse

License
-------

MIT, see [LICENSE](LICENSE)
