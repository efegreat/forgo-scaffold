package app

import (
	"github.com/efegreat/forgo/cli"
	"github.com/efegreat/forgo/commands/migration"
	"github.com/efegreat/forgo/commands/openapi"

	"example"
	"example/config"
	"example/internal/adapter/postgres"
	"example/internal/domain/messaging"
)

// CLI runs one command, which cmd/cli runs. opts are added last, such as cli.WithArgs in tests.
//
//   - migrate applies the migrations of migrations/postgres, and -a=new writes a new one there;
//   - openapi writes the OpenAPI documents of config.HTTP, such as api/openapi-v1.json;
//   - retry-failed sets the failed messages of a subscriber back to waiting, before they are replayed from
//     its dead-letter queue.
func CLI(opts ...cli.Option) *cli.CLI {
	return cli.New(append([]cli.Option{
		cli.WithName("example-cli"),
		cli.WithLogger(config.Logger),
		cli.WithResources(config.DB()),
		cli.WithProviders(postgres.NewFailedMessages),
		cli.WithConfig(cli.Config{Commands: []cli.CommandConfig{
			{
				Name: "migrate",
				Command: migration.Factory(migration.CommandConfig{
					Config: migration.Config{Dialect: "postgres"},
					FS:     example.Migrations,
					Dir:    "migrations/postgres", // also where -a=new writes, run from the root of the module
				}),
			},
			{
				Name: "openapi",
				Command: openapi.Factory(openapi.CommandConfig{
					HTTP:    config.HTTP,
					Options: []openapi.Option{openapi.WithInfo(openapi.Info{Title: "Orders API", Description: "Places and reads orders."})},
				}),
			},
			{Name: "retry-failed", Command: cli.As(messaging.NewRetryCommand)},
		}}),
	}, opts...)...)
}
