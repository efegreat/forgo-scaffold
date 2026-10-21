package shipping

import (
	"context"

	"github.com/efegreat/forgo/logger"
	"github.com/efegreat/forgo/operation"
	"github.com/efegreat/forgo/server/cron"

	"example/internal/port"
)

// dispatchedLogKey is the log key of the number of shipments a run of the dispatch job dispatched.
const dispatchedLogKey = "dispatched"

// DispatchJob dispatches the pending shipments, as a carrier's pickup would.
type DispatchJob struct {
	shipments port.ShipmentRepository
}

// NewDispatchJob returns the job, for cron.As.
func NewDispatchJob(shipments port.ShipmentRepository) cron.IJob {
	return &DispatchJob{shipments: shipments}
}

// Tolerance retries nothing: the next run dispatches what this one did not.
func (*DispatchJob) Tolerance() operation.Tolerance {
	return operation.Tolerance{}
}

func (j *DispatchJob) Run(ctx context.Context) error {
	n, err := j.shipments.DispatchPending(ctx)
	if err != nil {
		return err
	}
	logger.FromCtx(ctx).Info().Int64(dispatchedLogKey, n).Msg("shipments dispatched")
	return nil
}
