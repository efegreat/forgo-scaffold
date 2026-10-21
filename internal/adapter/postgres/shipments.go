package postgres

import (
	"context"
	"database/sql"

	"github.com/efegreat/forgo-adapter/provider/sqltx"

	"example/internal/entity"
	"example/internal/port"
)

// shipments is the port.ShipmentRepository on the shipments table.
type shipments struct {
	db *sql.DB
}

func NewShipments(db *sql.DB) port.ShipmentRepository {
	return &shipments{db: db}
}

func (r *shipments) Insert(ctx context.Context, s entity.Shipment) error {
	_, err := sqltx.From(ctx, r.db).ExecContext(ctx,
		`INSERT INTO shipments (order_id, sku, quantity, status) VALUES ($1, $2, $3, $4)`,
		s.OrderID, s.SKU, s.Quantity, s.Status)
	return err
}

func (r *shipments) DispatchPending(ctx context.Context) (int64, error) {
	res, err := sqltx.From(ctx, r.db).ExecContext(ctx,
		`UPDATE shipments SET status = $1, dispatched_at = CURRENT_TIMESTAMP WHERE status = $2`,
		entity.ShipmentStatusDispatched, entity.ShipmentStatusPending)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
