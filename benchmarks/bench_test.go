package benchmark

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/vattle/sqlboiler/queries/qm"
	null "gopkg.in/nullbio/null.v6"
	"github.com/zbyte/go-kallax/benchmarks/models"
)

func envOrDefault(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		v = def
	}
	return v
}

func dbURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		envOrDefault("DBUSER", "testing"),
		envOrDefault("DBPASS", "testing"),
		envOrDefault("DBHOST", "0.0.0.0:5432"),
		envOrDefault("DBNAME", "testing"),
	)
}

func openTestDB(b *testing.B) *sql.DB {
	db, err := sql.Open("postgres", dbURL())
	if err != nil {
		b.Fatalf("error opening db: %s", err)
	}
	return db
}

func openGormTestDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open("postgres", dbURL())
	if err != nil {
		b.Fatalf("error opening db: %s", err)
	}
	return db
}

var schemas = []string{
	`CREATE TABLE IF NOT EXISTS people (
			id serial primary key,
			name text
		)`,
	`CREATE TABLE IF NOT EXISTS pets (
			id serial primary key,
			name text,
			kind text,
			person_id integer references people(id)
		)`,
}

var tables = []string{"pets", "people"}

func setupDB(b *testing.B, db *sql.DB) *sql.DB {
	for _, s := range schemas {
		_, err := db.Exec(s)
		if err != nil {
			b.Fatalf("error creating schema: %s", err)
		}
	}

	return db
}

func teardownDB(b *testing.B, db *sql.DB) {
	for _, t := range tables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", t))
		if err != nil {
			b.Fatalf("error dropping table: %s", err)
		}
	}

	if err := db.Close(); err != nil {
		b.Fatalf("error closing db: %s", err)
	}
}

func mkPersonWithRels() *Person {
	return &Person{
		Name: "Dolan",
		Pets: []*Pet{
			{Name: "Garfield", Kind: Cat},
			{Name: "Oddie", Kind: Dog},
			{Name: "Reptar", Kind: Fish},
		},
	}
}

func mkGormPersonWithRels() *GORMPerson {
	return &GORMPerson{
		Name: "Dolan",
		Pets: []*GORMPet{
			{Name: "Garfield", Kind: string(Cat)},
			{Name: "Oddie", Kind: string(Dog)},
			{Name: "Reptar", Kind: string(Fish)},
		},
	}
}

func BenchmarkKallaxInsertWithRelationships(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	for i := 0; i < b.N; i++ {
		if err := store.Insert(mkPersonWithRels()); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}
}

func BenchmarkKallaxUpdateWithRelationships(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	pers := mkPersonWithRels()
	if err := store.Insert(pers); err != nil {
		b.Fatalf("error inserting: %s", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := store.Update(pers); err != nil {
			b.Fatalf("error updating: %s", err)
		}
	}
}

func BenchmarkSQLBoilerInsertWithRelationships(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	for i := 0; i < b.N; i++ {
		tx, _ := db.Begin()
		person := &models.Person{Name: null.StringFrom("Dolan")}
		if err := person.Insert(tx); err != nil {
			b.Fatalf("error inserting: %s", err)
		}

		err := person.SetPets(tx, true, []*models.Pet{
			{Name: null.StringFrom("Garfield"), Kind: null.StringFrom("cat")},
			{Name: null.StringFrom("Oddie"), Kind: null.StringFrom("dog")},
			{Name: null.StringFrom("Reptar"), Kind: null.StringFrom("fish")},
		}...)
		if err != nil {
			b.Fatalf("error inserting relationships: %s", err)
		}

		tx.Commit()
	}
}

func BenchmarkRawSQLInsertWithRelationships(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	for i := 0; i < b.N; i++ {
		p := mkPersonWithRels()
		tx, err := db.Begin()

		err = tx.QueryRow("INSERT INTO people (name) VALUES ($1) RETURNING id", p.Name).
			Scan(&p.ID)
		if err != nil {
			b.Fatalf("error inserting: %s", err)
		}

		for _, pet := range p.Pets {
			err := tx.QueryRow(
				"INSERT INTO pets (name, kind, person_id) VALUES ($1, $2, $3) RETURNING id",
				pet.Name, string(pet.Kind), p.ID,
			).Scan(&pet.ID)
			if err != nil {
				b.Fatalf("error inserting rel: %s", err)
			}
		}

		if err := tx.Commit(); err != nil {
			b.Fatalf("error committing transaction: %s", err)
		}
	}
}

func BenchmarkGORMInsertWithRelationships(b *testing.B) {
	store := openGormTestDB(b)
	setupDB(b, store.DB())
	defer teardownDB(b, store.DB())

	for i := 0; i < b.N; i++ {
		if db := store.Create(mkGormPersonWithRels()); db.Error != nil {
			b.Fatalf("error inserting: %s", db.Error)
		}
	}
}

func BenchmarkKallaxInsert(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	for i := 0; i < b.N; i++ {
		if err := store.Insert(&Person{Name: "foo"}); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}
}

func BenchmarkKallaxUpdate(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	pers := &Person{Name: "foo"}
	if err := store.Insert(pers); err != nil {
		b.Fatalf("error inserting: %s", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := store.Update(pers); err != nil {
			b.Fatalf("error updating: %s", err)
		}
	}
}

func BenchmarkSQLBoilerInsert(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	for i := 0; i < b.N; i++ {
		if err := (&models.Person{Name: null.StringFrom("foo")}).Insert(db); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}
}

func BenchmarkRawSQLInsert(b *testing.B) {
	db := setupDB(b, openTestDB(b))
	defer teardownDB(b, db)

	for i := 0; i < b.N; i++ {
		p := &Person{Name: "foo"}

		err := db.QueryRow("INSERT INTO people (name) VALUES ($1) RETURNING id", p.Name).
			Scan(&p.ID)
		if err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}
}

func BenchmarkGORMInsert(b *testing.B) {
	store := openGormTestDB(b)
	setupDB(b, store.DB())
	defer teardownDB(b, store.DB())

	for i := 0; i < b.N; i++ {
		if db := store.Create(&GORMPerson{Name: "foo"}); db.Error != nil {
			b.Fatalf("error inserting: %s", db.Error)
		}
	}
}

func BenchmarkKallaxQueryRelationships(b *testing.B) {
	db := openTestDB(b)
	setupDB(b, db)
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	for i := 0; i < 200; i++ {
		if err := store.Insert(mkPersonWithRels()); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := store.FindAll(NewPersonQuery().WithPets(nil).Limit(100))
			if err != nil {
				b.Fatalf("error retrieving persons: %s", err)
			}
		}
	})
}

func BenchmarkSQLBoilerQueryRelationships(b *testing.B) {
	db := openTestDB(b)
	setupDB(b, db)
	defer teardownDB(b, db)

	for i := 0; i < 200; i++ {
		person := &models.Person{Name: null.StringFrom("Dolan")}
		if err := person.Insert(db); err != nil {
			b.Fatalf("error inserting: %s", err)
		}

		err := person.SetPets(db, true, []*models.Pet{
			{Name: null.StringFrom("Garfield"), Kind: null.StringFrom("cat")},
			{Name: null.StringFrom("Oddie"), Kind: null.StringFrom("dog")},
			{Name: null.StringFrom("Reptar"), Kind: null.StringFrom("fish")},
		}...)
		if err != nil {
			b.Fatalf("error inserting relationships: %s", err)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := models.People(db, qm.Load("Pets"), qm.Limit(100)).All()
			if err != nil {
				b.Fatalf("error retrieving persons: %s", err)
			}
		}
	})
}

func BenchmarkRawSQLQueryRelationships(b *testing.B) {
	db := openTestDB(b)
	setupDB(b, db)
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	for i := 0; i < 200; i++ {
		if err := store.Insert(mkPersonWithRels()); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("SELECT * FROM people")
			if err != nil {
				b.Fatalf("error querying: %s", err)
			}

			var people []*GORMPerson
			for rows.Next() {
				var p GORMPerson
				if err := rows.Scan(&p.ID, &p.Name); err != nil {
					b.Fatalf("error scanning: %s", err)
				}

				r, err := db.Query("SELECT * FROM pets WHERE person_id = $1", p.ID)
				if err != nil {
					b.Fatalf("error querying relationships: %s", err)
				}

				for r.Next() {
					var pet GORMPet
					if err := r.Scan(&pet.ID, &pet.Name, &pet.Kind, &pet.PersonID); err != nil {
						b.Fatalf("error scanning relationship: %s", err)
					}
					p.Pets = append(p.Pets, &pet)
				}

				r.Close()
				people = append(people, &p)
			}

			_ = people
			rows.Close()
		}
	})
}

func BenchmarkGORMQueryRelationships(b *testing.B) {
	store := openGormTestDB(b)
	setupDB(b, store.DB())
	defer teardownDB(b, store.DB())

	for i := 0; i < 300; i++ {
		if db := store.Create(mkGormPersonWithRels()); db.Error != nil {
			b.Fatalf("error inserting: %s", db.Error)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var persons []*GORMPerson
			db := store.Preload("Pets").Limit(100).Find(&persons)
			if db.Error != nil {
				b.Fatalf("error retrieving persons: %s", db.Error)
			}
		}
	})
}

func BenchmarkKallaxQuery(b *testing.B) {
	db := openTestDB(b)
	setupDB(b, db)
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	for i := 0; i < 300; i++ {
		if err := store.Insert(&Person{Name: "foo"}); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := store.FindAll(NewPersonQuery())
			if err != nil {
				b.Fatalf("error retrieving persons: %s", err)
			}
		}
	})
}

func BenchmarkSQLBoilerQuery(b *testing.B) {
	db := openTestDB(b)
	setupDB(b, db)
	defer teardownDB(b, db)

	for i := 0; i < 300; i++ {
		if err := (&models.Person{Name: null.StringFrom("foo")}).Insert(db); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := models.People(db).All()
			if err != nil {
				b.Fatalf("error retrieving persons: %s", err)
			}
		}
	})
}

func BenchmarkRawSQLQuery(b *testing.B) {
	db := openTestDB(b)
	setupDB(b, db)
	defer teardownDB(b, db)

	store := NewPersonStore(db)
	for i := 0; i < 300; i++ {
		if err := store.Insert(&Person{Name: "foo"}); err != nil {
			b.Fatalf("error inserting: %s", err)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("SELECT * FROM people")
			if err != nil {
				b.Fatalf("error querying: %s", err)
			}

			var people []*Person
			for rows.Next() {
				var p Person
				err := rows.Scan(&p.ID, &p.Name)
				if err != nil {
					b.Fatalf("error scanning: %s", err)
				}
				people = append(people, &p)
			}

			_ = people
			rows.Close()
		}
	})
}

func BenchmarkGORMQuery(b *testing.B) {
	store := openGormTestDB(b)
	setupDB(b, store.DB())
	defer teardownDB(b, store.DB())

	for i := 0; i < 200; i++ {
		if db := store.Create(&GORMPerson{Name: "foo"}); db.Error != nil {
			b.Fatalf("error inserting: %s", db.Error)
		}
	}

	b.Run("query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var persons []*GORMPerson
			db := store.Find(&persons)
			if db.Error != nil {
				b.Fatal("error retrieving persons:", db.Error)
			}
		}
	})
}
