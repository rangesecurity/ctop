package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/rangesecurity/ctop/db"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
)

func DBCommand(migrations *migrate.Migrations) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "manage database migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "db.url",
				Value: "postgres://postgres:password123@localhost:5432/ctop",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)
					return migrator.Init(c.Context)
				},
			},
			{
				Name:  "reset",
				Usage: "resets migration table and recreates schema",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					bunDb := db.OpenDB(dbUrl)
					defer bunDb.Close()
					models := []interface{}{
						(*db.VoteEvent)(nil),
						(*db.NewRoundEvent)(nil),
						(*db.NewRoundStepEvent)(nil),
						(*db.Validators)(nil),
					}
					for _, model := range models {
						_, _ = bunDb.NewDropTable().Model(model).Exec(context.Background())
					}
					migrator := migrate.NewMigrator(bunDb, migrations)
					if err := migrator.Reset(c.Context); err != nil {
						return err
					}
					_, err := migrator.Migrate(c.Context)
					return err
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)

					group, err := migrator.Migrate(c.Context)
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no new migrations to run\n")
						return nil
					}

					fmt.Printf("migrated to %s\n", group)
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)

					group, err := migrator.Rollback(c.Context)
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no groups to roll back\n")
						return nil
					}

					fmt.Printf("rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)

					return migrator.Lock(c.Context)
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)
					return migrator.Unlock(c.Context)
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)

					name := strings.Join(c.Args().Slice(), "_")
					mf, err := migrator.CreateGoMigration(c.Context, name)
					if err != nil {
						return err
					}
					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)

					return nil
				},
			},
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)

					name := strings.Join(c.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(c.Context, name)
					if err != nil {
						return err
					}

					for _, mf := range files {
						fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrations)
					ms, err := migrator.MigrationsWithStatus(c.Context)
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())

					return nil
				},
			},
			{
				Name:  "mark_applied",
				Usage: "mark migrations as applied without actually running them",
				Action: func(c *cli.Context) error {
					dbUrl := c.String("db.url")
					db := db.OpenDB(dbUrl)
					defer db.Close()
					migrator := migrate.NewMigrator(db, migrations)

					group, err := migrator.Migrate(c.Context, migrate.WithNopMigration())
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no new migrations to mark as applied\n")
						return nil
					}

					fmt.Printf("marked as applied %s\n", group)
					return nil
				},
			},
		},
	}
}
