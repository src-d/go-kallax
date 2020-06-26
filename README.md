<img src="https://cdn.rawgit.com/networkteam/go-kallax/master/kallax.svg" width="400" />

[![GoDoc](https://godoc.org/github.com/networkteam/go-kallax?status.svg)](https://godoc.org/github.com/networkteam/go-kallax) [![Build Status](https://travis-ci.org/networkteam/go-kallax.svg?branch=master)](https://travis-ci.org/networkteam/go-kallax) [![codecov](https://codecov.io/gh/networkteam/go-kallax/branch/master/graph/badge.svg)](https://codecov.io/gh/networkteam/go-kallax) [![Go Report Card](https://goreportcard.com/badge/github.com/networkteam/go-kallax)](https://goreportcard.com/report/github.com/networkteam/go-kallax) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Kallax is a PostgreSQL typesafe ORM for the Go language.

**This repository is a fork of the original src-d/go-kallax to fix some important issues since the original package is not actively maintained anymore.**

It aims to provide a way of programmatically write queries and interact with a PostgreSQL database without having to write a single line of SQL, use strings to refer to columns and use values of any type in queries.

For that reason, the first priority of kallax is to provide type safety to the data access layer.
Another of the goals of kallax is make sure all models are, first and foremost, Go structs without having to use database-specific types such as, for example, `sql.NullInt64`.
Support for arrays of all basic Go types and all JSON and arrays operators is provided as well.

## Changes to the original version

- Full Go module support
- Removed statement cacher for prepared statements to fix memory leaks (when creating lots of new Store instances)

## Contents

- [Installation](#installation)
- [Usage](#usage)
- [Define models](#define-models)
  - [Struct tags](#struct-tags)
  - [Primary keys](#primary-keys)
  - [Model constructors](#model-constructors)
  - [Model events](#model-events)
- [Model schema](#model-schema)
  - [Use schema](#use-schema)
- [Manipulate models](#manipulate-models)
  - [Insert models](#insert-models)
  - [Update models](#update-models)
  - [Save models](#save-models)
  - [Delete models](#delete-models)
- [Query models](#query-models)
  - [Simple queries](#simple-queries)
  - [Generated findbys](#generated-findbys)
  - [Query with relationships](#query-with-relationships)
  - [Querying JSON](#querying-json)
- [Transactions](#transactions)
- [Caveats](#caveats)
- [Migrations](#migrations)
- [Custom operators](#custom-operators)
- [Debug SQL queries](#debug-sql-queries)
- [Benchmarks](#benchmarks)
- [Acknowledgements](#acknowledgements)
- [Contributing](#contributing)

## Installation

The recommended way to install `kallax` is:

```
go get -u github.com/networkteam/go-kallax/...
```

> _kallax_ includes a binary tool used by [go generate](http://blog.golang.org/generate),
> please be sure that `$GOPATH/bin` is on your `$PATH`

## Usage

Imagine you have the following file in the package where your models are.

```go
package models

type User struct {
        kallax.Model         `table:"users" pk:"id"`
        ID       kallax.ULID
        Username string
        Email    string
        Password string
}
```

Then put the following on any file of that package:

```go
//go:generate kallax gen
```

Now all you have to do is run `go generate ./...` and a `kallax.go` file will be generated with all the generated code for your model.

If you don't want to use `go generate`, even though is the preferred use, you can just go to your package and run `kallax gen` yourself.

### Excluding files from generation

Sometimes you might want to use the generated code in the same package it is defined and cause problems during the generation when you regenerate your models. You can exclude files in the package by changing the `go:generate` comment to the following:

```go
//go:generate kallax gen -e file1.go -e file2.go
```

## Define models

A model is just a Go struct that embeds the `kallax.Model` type. All the fields of this struct will be columns in the database table.

A model also needs to have one (and just one) primary key. The primary key is defined using the `pk` struct tag on the `kallax.Model` embedding. You can also set the primary key in a field of the struct with the struct tag `pk`, which can be `pk:""` for a non auto-incrementable primary key or `pk:"autoincr"` for one that is auto-incrementable.
More about primary keys is discussed at the [primary keys](#primary-keys) section.

First, let's review the rules and conventions for model fields:

- All the fields with basic types or types that implement [sql.Scanner](https://golang.org/pkg/database/sql/#Scanner) and [driver.Valuer](https://golang.org/pkg/database/sql/driver/#Valuer) will be considered a column in the table of their matching type.
- Arrays or slices of types mentioned above will be treated as PostgreSQL arrays of their matching type.
- Fields that are structs (or pointers to structs) or interfaces not implementing [sql.Scanner](https://golang.org/pkg/database/sql/#Scanner) and [driver.Valuer](https://golang.org/pkg/database/sql/driver/#Valuer) will be considered as JSON. Same with arrays or slices of types that follow these rules.
- Fields that are structs (or pointers to structs) with the struct tag `kallax:",inline"` or are embedded will be considered inline, and their fields would be considered as if they were at the root of the model.
- All pointer fields are nullable by default. That means you do not need to use `sql.NullInt64`, `sql.NullBool` and the likes because kallax automatically takes care of that for you. **WARNING:** all JSON and `sql.Scanner` implementors will be initialized with `new(T)` in case they are `nil` before they are scanned.
- By default, the name of a column will be the name of the struct field converted to lower snake case (e.g. `UserName` => `user_name`, `UserID` => `user_id`). You can override it with the struct tag `kallax:"my_custom_name"`.
- Slices of structs (or pointers to structs) that are models themselves will be considered a 1:N relationship. Arrays of models are **not supported** by design.
- A struct or pointer to struct field that is a model itself will be considered a 1:1 relationship.
- For relationships, the foreign key is assumed to be the name of the model converted to lower snake case plus `_id` (e.g. `User` => `user_id`). You can override this with the struct tag `fk:"my_custom_fk"`.
- For inverse relationship, you need to use the struct tag `fk:",inverse"`. You can combine the `inverse` with overriding the foreign key with `fk:"my_custom_fk,inverse"`. In the case of inverses, the foreign key name does not specify the name of the column in the relationship table, but the name of the column in the own table. The name of the column in the other table is always the primary key of the other model and cannot be changed for the time being.
- Foreign keys _do not have to be in the model_, they are automagically managed underneath by kallax.

Kallax also provides a `kallax.Timestamps` struct that contains `CreatedAt` and `UpdatedAt` that will be managed automatically.

Let's see an example of models with all these cases:

```go
type User struct {
        kallax.Model       `table:"users" pk:"id,autoincr"`
        kallax.Timestamps
        ID        int64
        Username  string
        Password  string
        Emails    []string
        // This is for demo purposes, please don't do this
        // 1:N relationships load all N rows by default, so
        // only do it when N is small.
        // If N is big, you should probably be querying the posts
        // table instead.
        Posts []*Post `fk:"poster_id"`
}

type Post struct {
        kallax.Model      `table:"posts"`
        kallax.Timestamps
        ID       int64    `pk:"autoincr"`
        Content  string   `kallax:"post_content"`
        Poster   *User    `fk:"poster_id,inverse"`
        Metadata Metadata `kallax:",inline"`
}

type Metadata struct {
        MetadataType MetadataType
        Metadata map[string]interface{} // this will be json
}
```

### Struct tags

| Tag                                     | Description                                                                                                                                                                         | Can be used in                             |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| `table:"table_name"`                    | Specifies the name of the table for a model. If not provided, the name of the table will be the name of the struct in lower snake case (e.g. `UserPreference` => `user_preference`) | embedded `kallax.Model`                    |
| `pk:"primary_key_column_name"`          | Specifies the column name of the primary key.                                                                                                                                       | embedded `kallax.Model`                    |
| `pk:"primary_key_column_name,autoincr"` | Specifies the column name of the autoincrementable primary key.                                                                                                                     | embedded `kallax.Model`                    |
| `pk:""`                                 | Specifies the field is a primary key                                                                                                                                                | any field with a valid identifier type     |
| `pk:"autoincr"`                         | Specifies the field is an auto-incrementable primary key                                                                                                                            | any field with a valid identifier type     |
| `kallax:"column_name"`                  | Specifies the name of the column                                                                                                                                                    | Any model field that is not a relationship |
| `kallax:"-"`                            | Ignores the field and does not store it                                                                                                                                             | Any model field                            |
| `kallax:",inline"`                      | Adds the fields of the struct field to the model. Column name can also be given before the comma, but it is ignored, since the field is not a column anymore                        | Any struct field                           |
| `fk:"foreign_key_name"`                 | Name of the foreign key column                                                                                                                                                      | Any relationship field                     |
| `fk:",inverse"`                         | Specifies the relationship is an inverse relationship. Foreign key name can also be given before the comma                                                                          | Any relationship field                     |
| `unique:"true"`                         | Specifies the column has an unique constraint.                                                                                                                                      | Any non-primary key field                  |

### Primary keys

Primary key types need to satisfy the [Identifier](https://godoc.org/github.com/networkteam/go-kallax/#Identifier) interface. Even though they have to do that, the generator is smart enough to know when to wrap some types to make it easier on the user.

The following types can be used as primary key:

- `int64`
- [`uuid.UUID`](https://godoc.org/github.com/gofrs/uuid#UUID)
- [`kallax.ULID`](https://godoc.org/github.com/networkteam/go-kallax/#ULID): this is a type kallax provides that implements a lexically sortable UUID. You can store it as `uuid` like any other UUID, but internally it's an ULID and you will be able to sort lexically by it.

Due to how sql mapping works, pointers to `uuid.UUID` and `kallax.ULID` are not set to `nil` if they appear as `NULL` in the database, but to [`uuid.Nil`](https://godoc.org/github.com/satori/go.uuid#pkg-variables). Using pointers to UUIDs is discouraged for this reason.

If you need another type as primary key, feel free to open a pull request implementing that.

**Known limitations**

- Only one primary key can be specified and it can't be a composite key.

### Model constructors

Kallax generates a constructor for your type named `New{TypeName}`. But you can customize it by implementing a private constructor named `new{TypeName}`. The constructor generated by kallax will use the same signature your private constructor has. You can use this to provide default values or construct the model with some values.

If you implement this constructor:

```go
func newUser(username, password string, emails ...string) (*User, error) {
        if username == "" || len(emails) == 0 || password == "" {
                return nil, errors.New("all fields are required")
        }

        return &User{Username: username, Password: password, Emails: emails}, nil
}
```

Kallax will generate one with the following signature:

```go
func NewUser(username string, password string, emails ...string) (*User, error)
```

**IMPORTANT:** if your primary key is not auto-incrementable, you should set an ID for every model you create in your constructor. Or, at least, set it before saving it. Inserting, updating, deleting or reloading an object with no primary key set will return an error.

If you don't implement your own constructor it's ok, kallax will generate one for you just instantiating your object like this:

```go
func NewT() *T {
        return new(T)
}
```

### Model events

Events can be defined for models and they will be invoked at certain times of the model lifecycle.

- `BeforeInsert`: will be called before inserting the model.
- `BeforeUpdate`: will be called before updating the model.
- `BeforeSave`: will be called before updating or inserting the model. It's always called before `BeforeInsert` and `BeforeUpdate`.
- `BeforeDelete`: will be called before deleting the model.
- `AfterInsert`: will be called after inserting the model. The presence of this event will cause the insertion of the model to run in a transaction. If the event returns an error, it will be rolled back.
- `AfterUpdate`: will be called after updating the model. The presence of this event will cause the update of the model to run in a transaction. If the event returns an error, it will be rolled back.
- `AfterSave`: will be called after updating or inserting the model. It's always called after `AfterInsert` and `AfterUpdate`. The presence of this event will cause the operation with the model to run in a transaction. If the event returns an error, it will be rolled back.
- `AfterDelete`: will be called after deleting the model. The presence of this event will cause the deletion to run in a transaction. If the event returns an error, it will be rolled back.

To implement these events, just implement the following interfaces. You can implement as many as you want:

- [BeforeInserter](https://godoc.org/github.com/networkteam/go-kallax#BeforeInserter)
- [BeforeUpdater](https://godoc.org/github.com/networkteam/go-kallax#BeforeUpdater)
- [BeforeSaver](https://godoc.org/github.com/networkteam/go-kallax#BeforeSaver)
- [BeforeDeleter](https://godoc.org/github.com/networkteam/go-kallax#BeforeDeleter)
- [AfterInserter](https://godoc.org/github.com/networkteam/go-kallax#AfterInserter)
- [AfterUpdater](https://godoc.org/github.com/networkteam/go-kallax#AfterUpdater)
- [AfterSaver](https://godoc.org/github.com/networkteam/go-kallax#AfterSaver)
- [AfterDeleter](https://godoc.org/github.com/networkteam/go-kallax#AfterDeleter)

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

- Internal methods for your model to make it work with kallax and satisfy the [Record](https://godoc.org/github.com/networkteam/go-kallax#Record) interface.
- A store named `{TypeName}Store`: the store is the way to access the data. A store of a given type is the way to access and manipulate data of that type. You can get an instance of the type store with `New{TypeName}Store(*sql.DB)`.
- A query named `{TypeName}Query`: the query is the way you will be able to build programmatically the queries to perform on the store. A store only will accept queries of its own type. You can create a new query with `New{TypeName}Query()`.
  The query will contain methods for adding criteria to your query for every field of your struct, called `FindBy`s. The query object is not immutable, that is, every condition added to it, changes the query. If you want to reuse part of a query, you can call the `Copy()` method of a query, which will return a query identical to the one used to call the method.
- A resultset named `{TypeName}ResultSet`: a resultset is the way to iterate over and obtain all elements in a resultset returned by the store. A store of a given type will always return a result set of the matching type, which will only return records of that type.
- Schema of all the models containing all the fields. That way, you can access the name of a specific field without having to use a string, that is, a typesafe way.

## Model schema

### Use schema

A global variable `Schema` will be created in your `kallax.go`, that contains a field with the name of every of your models. Those are the schemas of your models. Each model schema contains all the fields of that model.

So, to access the username field of the user model, it can be accessed as:

```go
Schema.User.Username
```

## Manipulate models

For all of the following sections, we will assume we have a store `store` for our model's type.

### Insert models

To insert a model we just need to use the `Insert` method of the store and pass it a model. If the primary key is not auto-incrementable and the object does not have one set, the insertion will fail.

```go
user := NewUser("fancy_username", "super_secret_password", "foo@email.me")
err := store.Insert(user)
if err != nil {
        // handle error
}
```

If our model has relationships, they will be saved, and so will the relationships of the relationships and so on. TL;DR: inserts are recursive.
**Note:** the relationships will be saved using `Save`, not `Insert`.

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

If our model has relationships, they will be saved, and so will the relationships of the relationships and so on. TL;DR: updates are recursive.
**Note:** the relationships will be saved using `Save`, not `Update`.

```go
user := FindLastPoster()
rowsUpdated, err := store.Update(user)
if err != nil {
        // handle error
}
```

If there are any relationships in the model, both the model and the relationships will be saved in a transaction and only succeed if all of them are saved correctly.

### Save models

To save a model we just need to use the `Save` method of the store and pass it a model. `Save` is just a shorthand that will call `Insert` if the model is not yet persisted and `Update` if it is.

```go
updated, err := store.Save(user)
if err != nil {
        // handle error
}

if updated {
        // it was updated, not inserted
}
```

If our model has relationships, they will be saved, and so will the relationships of the relationships and so on. TL;DR: saves are recursive.

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

Note that for that to work, the thing you're deleting must **not** be empty. That is, you need to eagerly load (or set afterwards) the relationships.

```go
user, err := store.FindOne(NewUserQuery())
checkErr(err)

// THIS WON'T WORK! We've not loaded "Things"
err := store.RemoveThings(user)

user, err := store.FindOne(NewUserQuery().WithThings())
checkErr(err)

// THIS WILL WORK!
err := store.RemoveThings(user)
```

## Query models

### Simple queries

To perform a query you have to do the following things:

- Create a query
- Pass the query to `Find`, `FindOne`, `MustFind` or `MustFindOne` of the store
- Gather the results from the result set, if the used method was `Find` or `MustFind`

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

Next will automatically close the result set when it hits the end. If you have to prematurely exit the iteration you can close it manually with `rs.Close()`.

You can query just a single row with `FindOne`.

```go
q := NewUserQuery().
        Where(kallax.Eq(Schema.User.Username, "Joe"))

user, err := store.FindOne(q)
```

You can also get all of the rows in a result without having to manually iterate the result set with `FindAll`.

```go
q := NewUserQuery().
        Where(kallax.Like(Schema.User.Username, "joe%")).
        Order(kallax.Asc(Schema.User.Username)).
        Limit(20).
        Offset(2)

users, err := store.FindAll(q)
if err != nil {
        // handle error
}
```

By default, all columns in a row are retrieved. To not retrieve all of them, you can specify the columns to include/exclude. Take into account that partial records retrieved from the database will not be writable. To make them writable you will need to [`Reload`](#reloading-a-model) the object.

```go
// Select only Username and password
NewUserQuery().Select(Schema.User.Username, Schema.User.Password)

// Select all but password
NewUserQuery().SelectNot(Schema.User.Password)
```

### Generated findbys

Kallax generates a `FindBy` for every field of your model for which it makes sense to do so. What is a `FindBy`? It is a shorthand to add a condition to the query for a specific field.

Consider the following model:

```go
type Person struct {
        kallax.Model
        ID        int64     `pk:"autoincr"`
        Name      string
        BirthDate time.Time
        Age       int
}
```

Four `FindBy`s will be generated for this model:

```go
func (*PersonQuery) FindByID(...int64) *PersonQuery
func (*PersonQuery) FindByName(string) *PersonQuery
func (*PersonQuery) FindByBirthDate(kallax.ScalarCond, time.Time) *PersonQuery
func (*PersonQuery) FindByAge(kallax.ScalarCond, int) *PersonQuery
```

That way, you can just do the following:

```go
NewPersonQuery().
        FindByAge(kallax.GtOrEq, 18).
        FindByName("Bobby")
```

instead of:

```go
NewPersonQuery().
        Where(kallax.GtOrEq(Schema.Person.Age, 18)).
        Where(kallax.Eq(Schema.Person.Name, "Bobby"))
```

Why are there three different types of methods generated?

- The primary key field is treated in a special way and allows multiple IDs to be passed, since searching by multiple IDs is a common operation.
- Types that are not often searched by equality (integers, floats, times, ...) allow an operator to be passed to them to determine the operator to use.
- Types that can only be searched by value (strings, bools, ...) only allow a value to be passed.

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

**NOTE:** if a filter is passed to a `With{Name}` method we can no longer guarantee that all related objects are there and, therefore, the retrieved records will **not** be writable.

### Reloading a model

If, for example, you have a model that is not writable because you only selected one field you can always reload it and have the full object. When the object is reloaded, all the changes made to the object that have not been saved will be discarded and overwritten with the values in the database.

```go
err := store.Reload(user)
```

Reload will not reload any relationships, just the model itself. After a `Reload` the model will **always** be writable.

### Querying JSON

You can query arbitrary JSON using the JSON operators defined in the [kallax](https://godoc.org/github.com/networkteam/go-kallax) package. The schema of the JSON (if it's a struct, obviously for maps it is not) is also generated.

```go
q := NewPostQuery().Where(kallax.JSONContainsAnyKey(
        Schema.Post.Metadata,
        "foo", "bar",
))
```

## Transactions

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

The fact that a transaction receives a store with the type of the model can be a problem if you want to store several models of different types. Kallax has a method named `StoreFrom` that initializes a store of the type you want to have the same underlying store as some other.

```go
store.Transaction(func(s *UserStore) error {
        var postStore PostStore
        kallax.StoreFrom(&postStore, s)

        for _, p := range posts {
                if err := postStore.Insert(p); err != nil {
                        return err
                }
        }

        return s.Insert(user)
})
```

`Transaction` can be used inside a transaction, but it does not open a new one, reuses the existing one.

## Caveats

- It is not possible to use slices or arrays of types that are not one of these types:
  - Basic types (e.g. `[]string`, `[]int64`) (except for `rune`, `complex64` and `complex128`)
  - Types that implement `sql.Scanner` and `driver.Valuer`
    The reason why this is not possible is because kallax implements support for arrays of all basic Go types by hand and also for types implementing `sql.Scanner` and `driver.Valuer` (using reflection in this case), but without having a common interface to operate on them, arbitrary types can not be supported.
    For example, consider the following type `type Foo string`, using `[]Foo` would not be supported. Know that this will fail during the scanning of rows and not in code-generation time for now. In the future, might be moved to a warning or an error during code generation.
    Aliases of slice types are supported, though. If we have `type Strings []string`, using `Strings` would be supported, as a cast like this `([]string)(&slice)` it's supported and `[]string` is supported.
- `time.Time` and `url.URL` need to be used as is. That is, you can not use a type `Foo` being `type Foo time.Time`. `time.Time` and `url.URL` are types that are treated in a special way, if you do that, it would be the same as saying `type Foo struct { ... }` and kallax would no longer be able to identify the correct type.
- `time.Time` fields will be truncated to remove its nanoseconds on `Save`, `Insert` or `Update`, since PostgreSQL will not be able to store them. PostgreSQL stores times with timezones as UTC internally. So, times will come back as UTC (you can use `Local` method to convert them back to the local timezone). You can change the timezone that will be used to bring times back from the database in [the PostgreSQL configuration](https://www.postgresql.org/docs/9.6/static/datatype-datetime.html).
- Multidimensional arrays or slices are **not supported** except inside a JSON field.

## Migrations

Kallax can generate migrations for your schema automatically, if you want to. It is a process completely separated from the model generation, so it does not force you to generate your migrations using kallax.

Sometimes, kallax won't be able to infer a type or you will want a specific column type for a field. You can specify so with the `sqltype` struct tag on a field.

```go
type Model struct {
        kallax.Model `table:"foo"`
        Stuff SuperCustomType `sqltype:"bytea"`
}
```

You can see the [**full list of default type mappings**](#type-mappings) between Go and SQL.

### Generate migrations

To generate a migration, you have to run the command `kallax migrate`.

```
kallax migrate --input ./users/ --input ./posts/ --out ./migrations --name initial_schema
```

The `migrate` command accepts the following flags:

| Name              | Repeated | Description                                                                                                                                                                                    | Default        |
| ----------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| `--name` or `-n`  | no       | name of the migration file (will be converted to `a_snakecase_name`)                                                                                                                           | `migration`    |
| `--input` or `-i` | yes      | every occurrence of this flag will specify a directory in which kallax models can be found. You can specify multiple times this flag if you have your models scattered across several packages | required       |
| `--out` or `-o`   | no       | destination folder where the migrations will be generated                                                                                                                                      | `./migrations` |

Every single migration consists of 2 files:

- `TIMESTAMP_NAME.up.sql`: script that will upgrade your database to this version.
- `TIMESTAMP_NAME.down.sql`: script that will downgrade your database to this version.

Additionally, there is a `lock.json` file where schema of the last migration is store to diff against the current models.

### Run migrations

To run a migration you can either use `kallax migrate up` or `kallax migrate down`. `up` will upgrade your database and `down` will downgrade it.

These are the flags available for `up` and `down`:

| Name                | Description                                                                                                                               | Default        |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| `--dir` or `-d`     | directory where your migrations are stored                                                                                                | `./migrations` |
| `--dsn`             | database connection string                                                                                                                | required       |
| `--steps` or `-s`   | maximum number of migrations to run                                                                                                       | `0`            |
| `--all`             | migrate all the way up (only available for `up`                                                                                           |
| `--version` or `-v` | final version of the database we want after running the migration. The version is the timestamp value at the beginning of migration files | `0`            |

- If no `--steps` or `--version` are provided to `down`, they will do nothing. If `--all` is provided to `up`, it will upgrade the database all the way up.
- If `--steps` and `--version` are provided to either `up` or `down` it will use only `--version`, as it is more specific.

**Example:**

```
kallax migrate up --dir ./my-migrations --dsn 'user:pass@localhost:5432/dbname?sslmode=disable' --version 1493991142
```

### Type mappings

| Go type                                  | SQL type                                                                  |
| ---------------------------------------- | ------------------------------------------------------------------------- |
| `kallax.ULID`                            | `uuid`                                                                    |
| `kallax.UUID`                            | `uuid`                                                                    |
| `kallax.NumericID`                       | `serial` on primary keys, `bigint` on foreign keys                        |
| `int64` on primary keys                  | `serial`                                                                  |
| `int64` on foreign keys and other fields | `bigint`                                                                  |
| `string`                                 | `text`                                                                    |
| `rune`                                   | `char(1)`                                                                 |
| `uint8`                                  | `smallint`                                                                |
| `int8`                                   | `smallint`                                                                |
| `byte`                                   | `smallint`                                                                |
| `uint16`                                 | `integer`                                                                 |
| `int16`                                  | `smallint`                                                                |
| `uint32`                                 | `bigint`                                                                  |
| `int32`                                  | `integer`                                                                 |
| `uint`                                   | `numeric(20)`                                                             |
| `int`                                    | `bigint`                                                                  |
| `int64`                                  | `bigint`                                                                  |
| `uint64`                                 | `numeric(20)`                                                             |
| `float32`                                | `real`                                                                    |
| `float64`                                | `double`                                                                  |
| `bool`                                   | `boolean`                                                                 |
| `url.URL`                                | `text`                                                                    |
| `time.Time`                              | `timestamptz`                                                             |
| `time.Duration`                          | `bigint`                                                                  |
| `[]byte`                                 | `bytea`                                                                   |
| `[]T`                                    | `T'[]` \* where `T'` is the SQL type of type `T`, except for `T` = `byte` |
| `map[K]V`                                | `jsonb`                                                                   |
| `struct`                                 | `jsonb`                                                                   |
| `*struct`                                | `jsonb`                                                                   |

Any other type must be explicitly specified.

All types that are not pointers will be `NOT NULL`.

## Custom operators

You can create custom operators with kallax using the `NewOperator` and `NewMultiOperator` functions.

`NewOperator` creates an operator with the specified format. It returns a function that given a schema field and a value returns a condition.

The format is a string in which `:col:` will get replaced with the schema field and `:arg:` will be replaced with the value.

```go
var Gt = kallax.NewOperator(":col: > :arg:")

// can be used like this:
query.Where(Gt(SomeSchemaField, 9000))
```

`NewMultiOperator` does exactly the same as the previous one, but it accepts a variable number of values.

```go
var In = kallax.NewMultiOperator(":col: IN :arg:")

// can be used like this:
query.Where(In(SomeSchemaField, 4, 5, 6))
```

This function already takes care of wrapping `:arg:` with parenthesis.

### Further customization

If you need further customization, you can create your own custom operator.

You need these things:

- A condition constructor (the operator itself) that takes the field and the values to create the proper SQL expression.
- A `ToSqler` that yields your SQL expression.

Imagine we want a greater than operator that only works with integers.

```go
func GtInt(col kallax.SchemaField, n int) kallax.Condition {
        return func(schema kallax.Schema) kallax.ToSqler {
                // it is VERY important that all SchemaFields
                // are qualified using the schema
                return &gtInt{col.QualifiedName(schema), n}
        }
}

type gtInt struct {
        col string
        val int
}

func (g *gtInt) ToSql() (sql string, params []interface{}, err error) {
        return fmt.Sprintf("%s > ?", g.col), []interface{}{g.val}, nil
}

// can be used like this:
query.Where(GtInt(SomeSchemaField, 9000))
```

For most of the operators, `NewOperator` and `NewMultiOperator` are enough, so the usage of these functions is preferred over the completely custom approach. Use it only if there is no other way to build your custom operator.

## Debug SQL queries

It is possible to debug the SQL queries being executed with kallax. To do that, you just need to call the `Debug` method of a store. This returns a new store with debugging enabled.

```go
store.Debug().Find(myQuery)
```

This will log to stdout using `log.Printf` `kallax: Query: THE QUERY SQL STATEMENT, args: [arg1 arg2]`.

You can use a custom logger (any function with a type `func(string, ...interface{})` using the `DebugWith` method instead.

```go
func myLogger(message string, args ...interface{}) {
        myloglib.Debugf("%s, args: %v", message, args)
}

store.DebugWith(myLogger).Find(myQuery)
```

## Acknowledgements

- Big thank you to the [Masterminds/squirrel](https://github.com/Masterminds/squirrel) library, which is an awesome query builder used internally in this ORM.
- [lib/pq](https://github.com/lib/pq), the Golang PostgreSQL driver that ships with a ton of support for builtin Go types.
- [mattes/migrate](https://github.com/mattes/migrate), a Golang library to manage database migrations.

## Contributing

### Reporting bugs

Kallax is a code generation tool, so it obviously has not been tested with all possible types and cases. If you find a case where the code generation is broken, please report an issue providing a minimal snippet for us to be able to reproduce the issue and fix it.

### Suggesting features

Kallax is a very opinionated ORM that works for us, so changes that make things not work for us or add complexity via configuration will not be considered for adding.
If we decide not to implement the feature you're suggesting, just keep in mind that it might not be because it is not a good idea, but because it does not work for us or is not aligned with the direction we want kallax to be moving forward.

### Running tests

For obvious reasons, an instance of PostgreSQL is required to run the tests of this package.

By default, it assumes that an instance exists at `0.0.0.0:5432` with an user, password and database name all equal to `testing`.

If that is not the case you can set the following environment variables:

- `DBNAME`: name of the database
- `DBUSER`: database user
- `DBPASS`: database user password

#### Docker PostgreSQL

If you have docker, you may run an instance of postgres in a container:

```
docker run -it --rm --name kallax \
 -e POSTGRES_PASSWORD=testing \
 -e POSTGRES_USER=testing \
 -e POSTGRES_DB=testing \
 -v `pwd`/.pgdata:/var/lib/postgresql/data \
 -p 127.0.0.1:5432:5432 \
 postgres:11
```

Remove `.pgdata` after you are done.

## License

MIT, see [LICENSE](LICENSE)
