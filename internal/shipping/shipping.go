// Package shipping ships placed orders: it records a shipment for each order.placed message, and a job
// dispatches the pending shipments.
package shipping

import (
	"context"

	"github.com/efegreat/forgo/logger"
	"github.com/efegreat/forgo/operation"
	"github.com/efegreat/forgo/server/cron"

	"example/internal/order"
)

const (
	// StatusPending is the status of a shipment that waits for its dispatch.
	StatusPending = "pending"
	// StatusDispatched is the status of a shipment that was dispatched.
	StatusDispatched = "dispatched"
)

// dispatchedLogKey is the log key of the number of shipments a run of the dispatch job dispatched.
const dispatchedLogKey = "dispatched"

// Shipment is the shipment of an order.
type Shipment struct {
	OrderID  string
	SKU      string
	Quantity int
	Status   string
}

// Repository stores shipments.
type Repository interface {
	Insert(ctx context.Context, s Shipment) error
	// DispatchPending dispatches every pending shipment, and returns how many it dispatched.
	DispatchPending(ctx context.Context) (int64, error)
}

// Service handles the messages of orders.
type Service struct {
	shipments Repository
}

// NewService returns a Service that stores shipments in shipments.
func NewService(shipments Repository) *Service {
	return &Service{shipments: shipments}
}

// OnOrderPlaced records the pending shipment of a placed order. The consumer runs it in a transaction
// together with the inbox, so a message delivered twice ships its order once.
func (s *Service) OnOrderPlaced(ctx context.Context, e order.Placed) error {
	return s.shipments.Insert(ctx, Shipment{OrderID: e.OrderID, SKU: e.SKU, Quantity: e.Quantity, Status: StatusPending})
}

// DispatchJob dispatches the pending shipments, as a carrier's pickup would.
type DispatchJob struct {
	shipments Repository
}

// NewDispatchJob returns the job, for cron.As.
func NewDispatchJob(shipments Repository) cron.IJob {
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
