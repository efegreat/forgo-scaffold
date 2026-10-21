package config

import (
	"github.com/efegreat/forgo/server/cron"

	"example/internal/domain/shipping"
)

// Cron dispatches the pending shipments every five minutes.
var Cron = cron.Config{
	Jobs: []cron.JobConfig{
		{Name: "dispatch_shipments", Schedule: "@every 5m", Job: cron.As(shipping.NewDispatchJob)},
	},
}
