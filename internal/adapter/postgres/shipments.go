package postgres

import (
	"context"
	"database/sql"

	"github.com/efegreat/forgo-adapter/sql/sqltx"

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

// DispatchPending reads the oldest pending shipments through the shipments_pending index, and skips those
// another transaction locked, so that two dispatches never take the same shipment.
func (r *shipments) DispatchPending(ctx context.Context, limit int) ([]entity.Shipment, error) {
	rows, err := sqltx.From(ctx, r.db).QueryContext(ctx, `
		UPDATE shipments SET status = $1, dispatched_at = CURRENT_TIMESTAMP
		WHERE order_id = ANY(ARRAY(
			SELECT order_id FROM shipments WHERE status = $2 ORDER BY created_at LIMIT $3 FOR UPDATE SKIP LOCKED
		))
		RETURNING order_id, sku, quantity, status`,
		entity.ShipmentStatusDispatched, entity.ShipmentStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dispatched []entity.Shipment
	for rows.Next() {
		var s entity.Shipment
		if err := rows.Scan(&s.OrderID, &s.SKU, &s.Quantity, &s.Status); err != nil {
			return nil, err
		}
		dispatched = append(dispatched, s)
	}
	return dispatched, rows.Err()
}
