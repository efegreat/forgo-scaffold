package app

import (
	"github.com/efegreat/forgo/cli"
	"github.com/efegreat/forgo/commands/migration"

	"example"
	"example/config"
)

// CLI runs one command, which cmd/cli runs: migrate, the migrations of migrations/postgres. opts are added
// last, such as cli.WithArgs in tests.
func CLI(opts ...cli.Option) *cli.CLI {
	return cli.New(append([]cli.Option{
		cli.WithName("example-cli"),
		cli.WithLogger(config.Logger),
		cli.WithResources(config.DB()),
		cli.WithConfig(cli.Config{Commands: []cli.CommandConfig{
			{
				Name: "migrate",
				Command: migration.Factory(migration.CommandConfig{
					Config: migration.Config{Dialect: "postgres"},
					FS:     example.Migrations,
					Dir:    "migrations/postgres", // also where -a=new writes, run from the root of the module
				}),
			},
		}}),
	}, opts...)...)
}
