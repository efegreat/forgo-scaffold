package app

import (
	"github.com/efegreat/forgo"

	"example/config"
)

// Worker is the app that handles messages and relays the outbox, which cmd/worker runs.
func Worker() *forgo.App {
	return forgo.NewApp(options("example-worker",
		forgo.WithResources(config.DB(), config.Broker()),
		forgo.WithServers(config.Consumer, config.Producer),
	)...)
}
