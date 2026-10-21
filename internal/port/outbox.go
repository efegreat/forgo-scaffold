package port

import (
	"context"

	"github.com/efegreat/forgo/server/worker"
)

// Outbox saves the messages to publish, in the transaction of ctx. The worker's producer relays them to
// the broker once that transaction commits, so a message is published if and only if the writes of its
// transaction are stored. forgo-adapter's outbox implements it: no adapter of this application has to.
type Outbox interface {
	Save(ctx context.Context, msg worker.Message) error
}
