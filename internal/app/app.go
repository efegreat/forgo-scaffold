// Package app builds the application of each deployable, one file each. Its main runs it, and its tests
// run the same app through forgo's forgotest, so the tests cover the wiring that production runs.
package app

import (
	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/provider/inbox"
	"github.com/efegreat/forgo-adapter/provider/outbox"
	"github.com/efegreat/forgo-adapter/provider/sqltx"
	_ "github.com/lib/pq" // the PostgreSQL driver of config.DB

	"example/config"
	"example/internal/adapter/postgres"
	"example/internal/domain/order"
	"example/internal/domain/shipping"
)

// providers are the providers every deployable shares. A provider may only depend on the ones before it,
// so forgo's storages and the adapters come before the domains. None of them connects to anything, so a
// deployable that does not use some of them pays nothing for them. Every provider is invoked when an app
// starts, though, so one that needs a resource only some deployables have, such as the broker, goes to
// the apps of those deployables instead, with forgo.WithProviders.
var providers = []any{
	// forgo's transactions, inbox and outbox, from forgo-adapter
	sqltx.NewTransactor,
	inbox.NewPostgres,
	outbox.NewPostgres,

	// the ports, from adapter
	postgres.NewOrders,
	postgres.NewShipments,
	postgres.NewOutbox,

	// the domains
	order.NewService,
	shipping.NewService,
}

// options returns the options of every app, named name, followed by opts: the app's resources and
// servers, and the providers only it has, which come after the shared ones.
func options(name string, opts ...forgo.AppOption) []forgo.AppOption {
	return append([]forgo.AppOption{
		forgo.WithName(name),
		forgo.WithLogger(config.Logger),
		forgo.WithProviders(providers...),
	}, opts...)
}
