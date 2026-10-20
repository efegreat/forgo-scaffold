// Package app builds the application of each deployable. Its main runs it, and its tests run the same app
// through forgo's forgotest, so the tests cover the wiring that production runs.
package app

import (
	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/provider/inbox"
	"github.com/efegreat/forgo-adapter/provider/outbox"
	"github.com/efegreat/forgo-adapter/provider/sqltx"
	"github.com/efegreat/forgo/cli"
	"github.com/efegreat/forgo/commands/migration"
	_ "github.com/lib/pq" // the PostgreSQL driver of config.DB

	"example"
	"example/config"
	"example/internal/order"
	"example/internal/postgres"
	"example/internal/shipping"
)

// Providers are the providers of every deployable, innermost first. None of them connects to anything,
// so a deployable that does not use some of them pays nothing for them.
var Providers = []any{
	sqltx.NewTransactor,
	inbox.NewPostgres,
	outbox.NewPostgres,
	postgres.NewOrders,
	postgres.NewShipments,
	postgres.NewOutbox,
	order.NewService,
	shipping.NewService,
}

// HTTP is the app of the orders API. It needs no broker: it saves messages to the outbox, which the worker
// relays.
func HTTP() *forgo.App {
	return newApp("example-http", []forgo.ResourceFactory{config.DB()}, config.HTTP)
}

// Worker is the app that handles messages and relays the outbox.
func Worker() *forgo.App {
	return newApp("example-worker", []forgo.ResourceFactory{config.DB(), config.Broker()}, config.Consumer, config.Producer)
}

// Cron is the app of the scheduled jobs.
func Cron() *forgo.App {
	return newApp("example-cron", []forgo.ResourceFactory{config.DB()}, config.Cron)
}

func newApp(name string, resources []forgo.ResourceFactory, servers ...forgo.ServerFactory) *forgo.App {
	return forgo.NewApp(
		forgo.WithName(name),
		forgo.WithLogger(config.Logger),
		forgo.WithResources(resources...),
		forgo.WithProviders(Providers...),
		forgo.WithServers(servers...),
	)
}

// CLI runs one command: migrate, the migrations of migrations/postgres. opts are added last, such as
// cli.WithArgs in tests.
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
