package config

import (
	"time"

	"github.com/efegreat/forgo/operation"
	"github.com/efegreat/forgo/server/cron"

	"example/internal/domain/shipping"
)

// Cron dispatches the pending shipments every five minutes. The lock, held in Redis, keeps two replicas of
// the cron deployable from dispatching at once: a run that finds it held is skipped. It expires after a
// minute, longer than a run takes, so a replica that crashed holding it never blocks the next runs.
var Cron = cron.Config{
	Jobs: []cron.JobConfig{
		{
			Name:     "dispatch_shipments",
			Schedule: "@every 5m",
			Lock:     operation.Lock{Key: "dispatch_shipments", TTL: time.Minute},
			Job:      cron.As(shipping.NewDispatchJob),
		},
	},
}
