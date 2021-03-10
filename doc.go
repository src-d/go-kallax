// Kallax is a PostgreSQL typesafe ORM for the Go language.
//
// Kallax aims to provide a way of programmatically write queries and interact
// with a PostgreSQL database without having to write a single line of SQL,
// use strings to refer to columns and use values of any type in queries.
// For that reason, the first priority of kallax is to provide type safety to
// the data access layer.
// Another of the goals of kallax is make sure all models are, first and
// foremost, Go structs without having to use database-specific types such as,
// for example, `sql.NullInt64`.
// Support for arrays of all basic Go types and all JSON and arrays operators is
// provided as well.
package kallax // import "github.com/networkteam/go-kallax"
