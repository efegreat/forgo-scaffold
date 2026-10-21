package app

import (
	"github.com/efegreat/forgo"

	"example/config"
)

// Cron is the app of the scheduled jobs, which cmd/cron runs.
func Cron() *forgo.App {
	return forgo.NewApp(options("example-cron",
		forgo.WithResources(config.DB()),
		forgo.WithServers(config.Cron),
	)...)
}
