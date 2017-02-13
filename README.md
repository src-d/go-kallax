Kallax - PostgreSQL ORM for Go
=============================

[![GoDoc](https://godoc.org/github.com/src-d/go-kallax?status.svg)](https://godoc.org/github.com/src-d/go-kallax) [![Build Status](https://travis-ci.org/src-d/go-kallax.svg?branch=master)](https://travis-ci.org/src-d/go-kallax) [![codecov](https://codecov.io/gh/src-d/go-kallax/branch/master/graph/badge.svg)](https://codecov.io/gh/src-d/go-kallax) [![Go Report Card](https://goreportcard.com/badge/github.com/src-d/go-kallax)](https://goreportcard.com/report/github.com/src-d/go-kallax) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


Kallax is a PostgreSQL typesafe ORM for the Go language.

It aims to provide a way of programmatically write queries and interact with a PostgreSQL database without having to write a single line of SQL, use strings to refer to columns and use values of any type in queries.

For that reason, the first priority of kallax is to provide type safety to the data access layer.
Another of the goals of kallax is make sure all models are, first and foremost, Go structs without having to use database-specific types such as, for example, `sql.NullInt64`.
Support for arrays of all basic Go types and all JSON and arrays operators is provided as well.

## Contents

* [Installation](#installation)
* [Usage](#usage)
* [Define models](#define-models)
  * [Struct tags](#struct-tags)
  * [Model constructors](#model-constructors)
  * [Model events](#model-events)
* [Model schema](#model-schema)
  * [Automatic schema generation and migrations](#automatic-schema-generation-and-migration)
  * [Use schema](#use-schema)
* [Manipulate models](#manipulate-models)
  * [Insert models](#insert-models)
  * [Update models](#update-models)
  * [Save models](#save-models)
  * [Delete models](#delete-models)
* [Query models](#query-models)
  * [Simple queries](#simple-queries)
  * [Query with relationships](#query-with-relationships)
  * [Querying JSON](#querying-json)
* [Transactions](#transactions)
* [Contributing](#contributing)

## Installation

The recommended way to install `kallax` is:

```
go get -u github.com/src-d/kallax/...
```

> *kallax* includes a binary tool used by [go generate](http://blog.golang.org/generate),
please be sure that `$GOPATH/bin` is on your `$PATH`

## Usage
 
Imagine you have the following file in the package where your models are.
```go
package models

type User struct {
        kallax.Model `table:"users"`
        Username string
        Email string
        Password string
}
```

Then put the following on any file of that package:

```go
//go:generate kallax gen
```

Now all you have to do is run `go generate ./...` and a `kallax.go` file will be generated with all the generated code for your model.

If you don't want to use `go generate`, even though is the preferred use, you can just go to your package and run `kallax gen` yourself.

## Define models

A model is just a Go struct that embeds the `kallax.Model` type. All the fields of this struct will be columns in the database table.

By embedding `kallax.Model`, you are already embedding the `ID` field. The `ID` is always an `UUID`. Right now, it is not possible to specify an ID that is not an UUID.

First, let's review the rules and conventions for model fields:
* All the fields with basic types or types that implement [sql.Scanner](https://golang.org/pkg/database/sql/#Scanner) and [driver.Valuer](https://golang.org/pkg/database/sql/driver/#Valuer) will be considered a column in the table of their matching type.
* Arrays or slices of types mentioned above will be treated as PostgreSQL arrays of their matching type.
* Fields that are structs (or pointers to structs) or interfaces not implementing [sql.Scanner](https://golang.org/pkg/database/sql/#Scanner) and [driver.Valuer](https://golang.org/pkg/database/sql/driver/#Valuer) will be considered as JSON. Same with arrays or slices of types that follow these rules.
* Fields that are structs (or pointers to structs) with the struct tag `kallax:",inline"` or are embedded will be considered inline, and their fields would be considered as if they were at the root of the model.
* By default, the name of a column will be the name of the struct field converted to lower snake case (e.g. `UserName` => `user_name`, `UserID` => `user_id`). You can override it with the struct tag `kallax:"my_custom_name"`.
* Slices or arrays of structs (or pointers to structs) that are models themselves will be considered a 1:N relationship.
* A struct or pointer to struct field that is a model itself will be considered a 1:1 relationship.
* For relationships, the foreign key is assumed to be the name of the model converted to lower snake case plus `_id` (e.g. `User` => `user_id`). You can override this with the struct tag `fk:"my_custom_fk"`.
* For inverse relationship, you need to use the struct tag `fk:",inverse"`. You can combine the `inverse` with overriding the foreign key with `fk:"my_custom_fk,inverse"`. In the case of inverses, the foreign key name does not specify the name of the column in the relationship table, but the name of the column in the own table. The name of the column in the other table is always supposed to be `id` and cannot be changed.
* Foreign keys *do not have to be in the model*, they are automagically managed underneath by kallax.

Kallax also provides a `kallax.Timestamps` struct that contains `CreatedAt` and `UpdatedAt` that will be managed automatically.

Let's see an example of models with all these cases:

```go
type User struct {
        kallax.Model `table:"users"`
        kallax.Timestamps
        Username string
        Password string
        Emails []string
        // This is for demo purposes, please don't do this
        // 1:N relationships load all N rows by default, so
        // only do it when N is small.
        // If N is big, you should probably be querying the posts
        // table instead.
        Posts []*Post `fk:"poster_id"`
}

type Post struct {
        kallax.Model `table:"posts"`
        kallax.Timestamps
        Content string `kallax:"post_content"`
        Poster *User `fk:"poster_id,inverse"`
        Metadata Metadata `kallax:",inline"`
}

type Metadata struct {
        MetadataType MetadataType
        Metadata map[string]interface{} // this will be json
}
```

### Struct tags

| Tag | Description | Can be used in |
| --- | --- | --- | --- |
| `table"table_name"` | Specifies the name of the table for a model | embedded `kallax.Model` |
| `kallax:"column_name"` | Specifies the name of the column | Any model field that is not a relationship |
| `kallax:"-"` | Ignores the field and does not store it | Any model field |
| `kallax:",inline"` | Adds the fields of the struct field to the model. Column name can also be given before the comma | Any struct field |
| `fk:"foreign_key_name"` | Name of the foreign key column | Any relationship field |
| `fk:",inverse"` | Specifies the relationship is an inverse relationship. Foreign key name can also be given before the comma | Any relationship field |

### Model constructors

Kallax generates a constructor for your type named `New{TypeName}`. But you can customize it by implementing a private constructor named `new{TypeName}`. The constructor generated by kallax will use the same signature your private constructor has. You can use this to provide default values or construct the model with some values.

If you implement this constructor:

```go
func newUser(username, password string, emails ...string) (*User, error) {
        if username != "" || len(emails) == 0 || password != "" {
                return errors.New("all fields are required")
        }

        return &User{Username: username, Password: password, Emails: emails}
}
```

Kallax will generate one with the following signature:

```go
func NewUser(username string, password string, emails ...string) (*User, error)
```

Then, why is it needed that kallax generates the public constructor? To make sure all the model internal fields are initialized correctly, set and ID for the model, etc.

### Model events

Events can be defined for models and they will be invoked at certain times of the model lifecycle.

* `BeforeInsert`: will be called before inserting the model.
* `BeforeUpdate`: will be called before updating the model.
* `BeforeSave`: will be called before updating or inserting the model. It's always called before `BeforeInsert` and `BeforeUpdate`.
* `BeforeDelete`: will be called before deleting the model.
* `AfterInsert`: will be called after inserting the model. The presence of this event will cause the insertion of the model to run in a transaction. If the event returns an error, it will be rolled back.
* `AfterUpdate`: will be called after updating the model. The presence of this event will cause the update of the model to run in a transaction. If the event returns an error, it will be rolled back.
* `AfterSave`: will be called after updating or inserting the model. It's always called after `AfterInsert` and `AfterUpdate`. The presence of this event will cause the operation with the model to run in a transaction. If the event returns an error, it will be rolled back.
* `AfterDelete`: will be called after deleting the model. The presence of this event will cause the deletion to run in a transaction. If the event returns an error, it will be rolled back.

To implement these events, just implement the following interfaces. You can implement as many as you want:

* [BeforeInserter](https://godoc.org/github.com/src-d/go-kallax#BeforeInserter)
* [BeforeUpdater](https://godoc.org/github.com/src-d/go-kallax#BeforeUpdater)
* [BeforeSaver](https://godoc.org/github.com/src-d/go-kallax#BeforeSaver)
* [BeforeDeleter](https://godoc.org/github.com/src-d/go-kallax#BeforeDeleter)
* [AfterInserter](https://godoc.org/github.com/src-d/go-kallax#AfterInserter)
* [AfterUpdater](https://godoc.org/github.com/src-d/go-kallax#AfterUpdater)
* [AfterSaver](https://godoc.org/github.com/src-d/go-kallax#AfterSaver)
* [AfterDeleter](https://godoc.org/github.com/src-d/go-kallax#AfterDeleter)

Example:

```go
func (u *User) BeforeSave() error {
        if u.Password == "" {
                return errors.New("cannot save user without password")
        }

        if !isCrypted(u.Password) {
                u.Password = crypt(u.Password)
        }
        return nil
}
```

## Kallax generated code

Kallax generates a bunch of code for every single model you have and saves it to a file named `kallax.go` in the same package.

For every model you have, kallax will generate the following for you:

* Internal methods for your model to make it work with kallax and satisfy the [Record](https://godoc.org/github.com/src-d/go-kallax#Record) interface.
* A store named `{TypeName}Store`: the store is the way to access the data. A store of a given type is the way to access and manipulate data of that type.
* A query named `{TypeName}Query`: the query is the way you will be able to build programmatically the queries to perform on the store. A store only will accept queries of its own type.
The query will contain methods for adding criteria to your query for every field of your struct, called `FindBy`s.
* A resultset named `{TypeName}ResultSet`: a resultset is the way to iterate over and obtain all elements in a resultset returned by the store. A store of a given type will always return a result set of the matching type, which will only return records of that type.
* Schema of all the models containing all the fields. That way, you can access the name of a specific field without having to use a string, that is, a typesafe way.

## Model schema

### Automatic schema generation and migrations

Automatic `CREATE TABLE` for models and migrations is not yet supported, even though it will probably come in future releases.

### Use schema

A global variable `Schema` will be created in your `kallax.go`, that contains a field with the name of every of your models. Those are the schemas of your models. Each model schema contains all the fields of that model.

So, to access the username field of the user model, it can be accessed as:

```go
Schema.User.Username
```

## Manipulate models

For all of the following sections, we will assume we have a store `store` for our model's type.

### Insert models

To insert a model we just need to use the `Insert` method of the store and pass it a model. If the model does not have an ID, one will be assigned to it.

```go
user := NewUser("fancy_username", "super_secret_password", "foo@email.me")
err := store.Insert(user)
if err != nil {
        // handle error
}
```

If our model has relationships, they will be saved (**note:** saved as in insert or update) as well. The relationships of the relationships will not, though. Relationships are only saved with one level of depth.

```go
user := NewUser("foo")
user.Posts = append(user.Posts, NewPost(user, "new post"))

err := store.Insert(user)
if err != nil {
        // handle error
}
```

If there are any relationships in the model, both the model and the relationships will be saved in a transaction and only succeed if all of them are saved correctly.

### Update models

To insert a model we just need to use the `Update` method of the store and pass it a model. It will return an error if the model was not already persisted or has not an ID.

```go
user := FindLast()
rowsUpdated, err := store.Update(user)
if err != nil {
        // handle error
}
```

By default, when a model is updated, all its fields are updated. You can also specify which fields to update passing them to update.

```go
rowsUpdated, err := store.Update(user, Schema.User.Username, Schema.User.Password)
if err != nil {
        // handle error
}
```

If our model has relationships, they will be saved (**note:** saved as in insert or update) as well. The relationships of the relationships will not, though. Relationships are only saved with one level of depth.

```go
user := FindLastPoster()
rowsUpdated, err := store.Update(user)
if err != nil {
        // handle error
}
```

If there are any relationships in the model, both the model and the relationships will be saved in a transaction and only succeed if all of them are saved correctly.

### Save models

To save a model we just need to use the `Save` method of the store and pass it a model. It will update the model if it was already persisted or insert it otherwise.

```go
updated, err := store.Save(user)
if err != nil {
        // handle error
}

if updated {
        // it was updated, not inserted
}
```

If our model has relationships, they will be saved as well. The relationships of the relationships will not, though. Relationships are only saved with one level of depth.

```go
user := NewUser("foo")
user.Posts = append(user.Posts, NewPost(user, "new post"))

updated, err := store.Save(user)
if err != nil {
        // handle error
}
```

If there are any relationships in the model, both the model and the relationships will be saved in a transaction and only succeed if all of them are saved correctly.

### Delete models

To delete a model we just have to use the `Delete` method of the store. It will return an error if the model was not already persisted.

```go
err := store.Delete(user)
if err != nil {
        // handle error
}
```

Relationships of the model are **not** automatically removed using `Delete`.

For that, specific methods are generated in the store of the model.

For one to many relationships:

```go
// remove specific posts
err := store.RemovePosts(user, post1, post2, post3)
if err != nil {
        // handle error
}

// remove all posts
err := store.RemovePosts(user)
```

For one to one relationships:

```go
// remove the thing
err := store.RemoveThing(user)
```

## Query models

### Simple queries

To perform a query you have to do the following things: 
* Create a query
* Pass the query to `Find`, `FindOne`, `MustFind` or `MustFindOne` of the store
* Gather the results from the result set

```go
// Create the query
q := NewUserQuery().
        Where(kallax.Like(Schema.User.Username, "joe%")).
        Order(kallax.Asc(Schema.User.Username)).
        Limit(20).
        Offset(2)

rs, err := store.Find(q)
if err != nil {
        // handle error
}

for rs.Next() {
        user, err := rs.Get()
        if err != nil {
                // handle error
        }
}
```

You can query just a single row with `FindOne`.

```go
q := NewUserQuery().
        Where(kallax.Eq(Schema.User.Username, "Joe"))

user, err := store.FindOne(q)
```

By default, all columns in a row are retrieved. To not retrieve all of them, you can specify the columns to include/exclude.

```go
// Select only Username and password
NewUserQuery().Select(Schema.User.Username, Schema.User.Password)

// Select all but password
NewUserQuery().SelectNot(Schema.User.Password)
```

### Count results

Instead of passing the query to `Find` or `FindOne`, you can pass it to `Count` to get the number of rows in the resultset.

```go
n, err := store.Count(q)
```

### Query with relationships

By default, no relationships are retrieved unless the query specifies so.

For each of your relationships, a method in your query is created to be able to include these relationships in your query.

One to one relationships:

```go
// Select all posts including the user that posted them
q := NewPostQuery().WithPoster()
rs, err := store.Find(q)
```

One to one relationships are always included in the same query. So, if you have 4 one to one relationships and you want them all, only 1 query will be done, but everything will be retrieved.

One to many relationships:

```go
// Select all users including their posts
// NOTE: this is a really bad idea, because all posts will be loaded
// if the N side of your 1:N relationship is big, consider querying the N store
// instead of doing this
// A condition can be passed to the `With{Name}` method to filter the results.
q := NewUserQuery().WithPosts(nil)
rs, err := store.Find(q)
```

To avoid the N+1 problem with 1:N relationships, kallax performs batching in this case. 
So, a batch of users are retrieved from the database in a single query, then all the posts for those users and finally, they are merged. 
This process is repeated until there are no more rows in the result.
Because of this, retrieving 1:N relationships is really fast.

The default batch size is 50, you can change this using the `BatchSize` method all queries have.

### Querying JSON

You can query arbitrary JSON using the JSON operators defined in the [kallax](https://godoc.org/github.com/src-d/go-kallax) package. The schema of the JSON (if it's a struct, obviously for maps it is not) is also generated.

```go
q := NewPostQuery().Where(kallax.JSONContainsAnyKey(
        Schema.Post.Metadata,
        "foo", "bar",
))
```

### Transactions

To execute things in a transaction the `Transaction` method of the model store can be used. All the operations done using the store provided to the callback will be run in a transaction.
If the callback returns an error, the transaction will be rolled back.

```go
store.Transaction(func(s *UserStore) error {
        if err := s.Insert(user1); err != nil {
                return err
        }

        return s.Insert(user2)
})
```

The fact that a transaction receives a store with the type of the model can be a problem if you want to store several models of different types. You can, indeed, create new stores of the other types, but do so with care. Do not use the internal `*kallax.Store`, as it does not perform any type checks or some of the operations the concrete type stores do.

```go
store.Transaction(func(s *UserStore) error {
        postStore := &PostStore{s.Store}

        for _, p := range posts {
                if err := postStore.Insert(p); err != nil {
                        return err
                }
        }

        return s.Insert(user)
})
```

`Transaction` can be used inside a transaction, but it does not open a new one, reuses the existing one.

## Contributing 

### Suggesting features

Kallax is a very opinionated ORM that works for us, so changes that make things not work for us or add complexity via configuration will not be considered for adding.
If we decide not to implement the feature you're suggesting, just keep in mind that it might not be because it is not a good ide, but because it does not work for us.

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
