package app

import (
	"github.com/efegreat/forgo"

	"example/config"
)

// HTTP is the app of the orders API, which cmd/http runs. It needs no broker: it saves messages to the
// outbox, which the worker relays.
func HTTP() *forgo.App {
	return forgo.NewApp(options("example-http",
		forgo.WithResources(config.DB()),
		forgo.WithServers(config.HTTP),
	)...)
}
