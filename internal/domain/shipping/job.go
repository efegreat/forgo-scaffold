package shipping

import (
	"context"

	"github.com/efegreat/forgo/logger"
	"github.com/efegreat/forgo/operation"
	"github.com/efegreat/forgo/server/cron"

	"example/internal/event"
	"example/internal/port"
)

const (
	// dispatchedLogKey is the log key of the number of shipments a run of the dispatch job dispatched.
	dispatchedLogKey = "dispatched"
	// dispatchBatch is the number of shipments dispatched in one transaction.
	dispatchBatch = 100
)

// DispatchJob dispatches the pending shipments, as a carrier's pickup would, and announces each with an
// event.ShipmentDispatched message, saved to the outbox in the transaction that dispatches it.
type DispatchJob struct {
	tx        operation.Transactor
	shipments port.ShipmentRepository
	outbox    port.Outbox
}

// NewDispatchJob returns the job, for cron.As.
func NewDispatchJob(tx operation.Transactor, shipments port.ShipmentRepository, outbox port.Outbox) cron.IJob {
	return &DispatchJob{tx: tx, shipments: shipments, outbox: outbox}
}

// Tolerance retries nothing: the next run dispatches what this one did not.
func (*DispatchJob) Tolerance() operation.Tolerance {
	return operation.Tolerance{}
}

// Run dispatches batches of shipments until a batch is short, so that no transaction grows with the backlog.
func (j *DispatchJob) Run(ctx context.Context) error {
	total := 0
	for {
		n, err := j.dispatch(ctx)
		total += n
		if err != nil {
			return err
		}
		if n < dispatchBatch {
			break
		}
	}
	logger.FromCtx(ctx).Info().Int(dispatchedLogKey, total).Msg("shipments dispatched")
	return nil
}

// dispatch dispatches one batch, and saves its messages, in one transaction.
func (j *DispatchJob) dispatch(ctx context.Context) (int, error) {
	n := 0
	err := j.tx.Do(ctx, func(ctx context.Context) error {
		dispatched, err := j.shipments.DispatchPending(ctx, dispatchBatch)
		if err != nil {
			return err
		}
		for _, s := range dispatched {
			msg, err := event.ShipmentDispatched{OrderID: s.OrderID}.Message()
			if err != nil {
				return err
			}
			if err := j.outbox.Save(ctx, msg); err != nil {
				return err
			}
		}
		n = len(dispatched)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}
