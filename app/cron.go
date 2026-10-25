package app

import (
	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/redis/locker"

	"example/config"
)

// Cron is the app of the scheduled jobs, which cmd/cron runs. Only it has Redis, which holds the locks of its
// jobs, so the locker is a provider of its own.
func Cron() *forgo.App {
	return forgo.NewApp(options("example-cron",
		forgo.WithResources(config.Redis()),
		forgo.WithProviders(locker.NewRedis),
		forgo.WithServers(config.Cron, config.Ops(":8083", 0)),
	)...)
}
