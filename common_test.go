package kallax

import "database/sql"

func openTestDB() (*sql.DB, error) {
	return sql.Open("postgres", "postgres://testing:testing@0.0.0.0:5432/testing?sslmode=disable")
}
