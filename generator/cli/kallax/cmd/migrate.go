package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/urfave/cli"

	"github.com/networkteam/go-kallax/generator"
)

var Migrate = cli.Command{
	Name:   "migrate",
	Usage:  "Generate migrations for current kallax models",
	Action: migrateAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "out, o",
			Usage: "Output directory of migrations",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "Descriptive name for the migration",
			Value: "migration",
		},
		cli.StringSliceFlag{
			Name:  "input, i",
			Usage: "List of directories to scan models from. You can use this flag as many times as you want.",
		},
	},
	Subcommands: cli.Commands{
		Up,
		Down,
	},
}

var migrationFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "dir, d",
		Value: "./migrations",
		Usage: "Directory where your migrations are stored",
	},
	cli.StringFlag{
		Name:  "dsn",
		Usage: "PostgreSQL data source name. Example: `user:pass@localhost:5432/database?sslmode=enable`",
	},
	cli.UintFlag{
		Name:  "steps, n",
		Usage: "Number of migrations to run",
	},
	cli.UintFlag{
		Name:  "version, v",
		Usage: "Migrate to a specific version. If `steps` and this flag are given, this will be used.",
	},
}

var Up = cli.Command{
	Name:   "up",
	Usage:  "Executes the migrations from the current version until the specified version.",
	Action: runMigrationAction(upAction),
	Flags: append(migrationFlags, cli.BoolFlag{
		Name:  "all",
		Usage: "If this flag is used, the database will be migrated all the way up.",
	}),
}

var Down = cli.Command{
	Name:   "down",
	Usage:  "Downgrades the database a certain number of migrations or until a certain version.",
	Action: runMigrationAction(downAction),
	Flags:  migrationFlags,
}

func upAction(m *migrate.Migrate, steps, version uint, all bool) error {
	if all {
		if err := m.Up(); err != nil {
			return fmt.Errorf("kallax: unable to upgrade the database all the way up: %s", err)
		}
	} else if version > 0 {
		if err := m.Migrate(version); err != nil {
			return fmt.Errorf("kallax: unable to upgrade up to version %d: %s", version, err)
		}
	} else if steps > 0 {
		if err := m.Steps(int(steps)); err != nil {
			return fmt.Errorf("kallax: unable to execute %d migration(s) up: %s", steps, err)
		}
	} else {
		return fmt.Errorf("WARN: No `version` or `steps` provided")
	}
	reportMigrationSuccess(m)
	return nil
}

func downAction(m *migrate.Migrate, steps, version uint, all bool) error {
	if version > 0 {
		if err := m.Migrate(version); err != nil {
			return fmt.Errorf("kallax: unable to rollback to version %d: %s", version, err)
		}
	} else if steps > 0 {
		if err := m.Steps(-int(steps)); err != nil {
			return fmt.Errorf("kallax: unable to execute %d migration(s) down: %s", steps, err)
		}
	} else {
		return fmt.Errorf("kallax: no `version` or `steps` provided, you need to specify one of them")
	}
	reportMigrationSuccess(m)
	return nil
}

func reportMigrationSuccess(m *migrate.Migrate) {
	fmt.Println("Success! the migration has been run.")

	if v, _, err := m.Version(); err != nil {
		fmt.Printf("Unable to check the latest version of the database: %s.\n", err)
	} else {
		fmt.Printf("Database is now at version %d.\n", v)
	}
}

type runMigrationFunc func(m *migrate.Migrate, steps, version uint, all bool) error

func runMigrationAction(fn runMigrationFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			dir     = c.String("dir")
			dsn     = c.String("dsn")
			steps   = c.Uint("steps")
			version = c.Uint("version")
			all     = c.Bool("all")
		)

		ok, err := isDirectory(dir)
		if err != nil {
			return fmt.Errorf("kallax: cannot check if `dir` is a directory: %s", err)
		}

		if !ok {
			return fmt.Errorf("kallax: argument `dir` must be a valid directory")
		}

		dir, err = filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("kallax: cannot get absolute path of `dir`: %s", err)
		}

		m, err := migrate.New(pathToFileURL(dir), fmt.Sprintf("postgres://%s", dsn))
		if err != nil {
			return fmt.Errorf("kallax: unable to open a connection with the database: %s", err)
		}

		return fn(m, steps, version, all)
	}
}

func pathToFileURL(path string) string {
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return ""
		}
	}
	return fmt.Sprintf("file://%s", filepath.ToSlash(path))
}

func migrateAction(c *cli.Context) error {
	dirs := c.StringSlice("input")
	dir := c.String("out")
	name := c.String("name")

	var pkgs []*generator.Package
	for _, dir := range dirs {
		ok, err := isDirectory(dir)
		if err != nil {
			return fmt.Errorf("kallax: cannot check directory in `input`: %s", err)
		}

		if !ok {
			return fmt.Errorf("kallax: `input` must be a valid directory")
		}

		p := generator.NewProcessor(dir, nil)
		p.Silent()
		pkg, err := p.Do()
		if err != nil {
			return err
		}

		pkgs = append(pkgs, pkg)
	}

	ok, err := isDirectory(dir)
	if err != nil {
		return fmt.Errorf("kallax: cannot check directory in `out`: %s", err)
	}

	if !ok {
		return fmt.Errorf("kallax: `out` must be a valid directory")
	}

	g := generator.NewMigrationGenerator(name, dir)
	migration, err := g.Build(pkgs...)
	if err != nil {
		return err
	}

	return g.Generate(migration)
}
