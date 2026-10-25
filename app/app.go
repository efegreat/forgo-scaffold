// Package app builds the application of each deployable, one file each. Its main runs it, and its tests
// run the same app through forgo's forgotest, so the tests cover the wiring that production runs.
package app

import (
	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/otel/observer"
	"github.com/efegreat/forgo-adapter/sql/inbox"
	"github.com/efegreat/forgo-adapter/sql/liveconfig"
	"github.com/efegreat/forgo-adapter/sql/outbox"
	"github.com/efegreat/forgo-adapter/sql/sqltx"
	_ "github.com/lib/pq" // the PostgreSQL driver of config.DB

	"example/config"
	"example/internal/adapter/postgres"
	"example/internal/domain/order"
	"example/internal/domain/shipping"
)

// providers are the providers every deployable shares. A provider may only depend on the ones before it,
// so forgo's storages and the adapters come before the domains. None of them connects to anything, so a
// deployable that does not use some of them pays nothing for them. Every provider is invoked when an app
// starts, though, so one that needs a resource only some deployables have, such as Redis, goes to the apps
// of those deployables instead, with forgo.WithProviders.
var providers = []any{
	// forgo's transactions, inbox, outbox, live configs and metrics, from forgo-adapter. The outbox is
	// port.Outbox too.
	sqltx.NewTransactor,
	inbox.NewPostgres,
	outbox.NewPostgres,
	liveconfig.NewPostgres,
	forgo.NewUncachedLiveConfigs,
	observer.New(observer.Config{}),

	// the ports, from adapter
	postgres.NewOrders,
	postgres.NewShipments,
	postgres.NewFailedMessages,

	// the domains
	order.NewService,
	shipping.NewService,
}

// options returns the options of the app named name: the logger, the shared providers, and the resources
// every deployable has, the database and the meter of its metrics. opts follow: the app's other resources,
// its servers, and the providers only it has, which come after the shared ones.
func options(name string, opts ...forgo.AppOption) []forgo.AppOption {
	return append([]forgo.AppOption{
		forgo.WithName(name),
		forgo.WithLogger(config.Logger),
		forgo.WithResources(config.DB(), config.Meter(name)),
		forgo.WithProviders(providers...),
	}, opts...)
}
