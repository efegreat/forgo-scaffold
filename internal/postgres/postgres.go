// Package postgres stores the domain on PostgreSQL. Every statement runs through sqltx.From, so it joins
// the transaction its context carries: that of a service's operation.Transactor, or the consumer's, which
// commits a handler's writes together with the inbox.
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/efegreat/forgo-adapter/provider/sqltx"
	"github.com/efegreat/forgo/server/worker"

	"example/internal/order"
	"example/internal/shipping"
)

// Orders is the order.Repository on the orders table.
type Orders struct {
	db *sql.DB
}

func NewOrders(db *sql.DB) *Orders {
	return &Orders{db: db}
}

func (r *Orders) Insert(ctx context.Context, o order.Order) error {
	_, err := sqltx.From(ctx, r.db).ExecContext(ctx,
		`INSERT INTO orders (id, sku, quantity, status) VALUES ($1, $2, $3, $4)`,
		o.ID, o.SKU, o.Quantity, o.Status)
	return err
}

func (r *Orders) Get(ctx context.Context, id string) (order.Order, bool, error) {
	var o order.Order
	err := sqltx.From(ctx, r.db).QueryRowContext(ctx,
		`SELECT id, sku, quantity, status, created_at FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.SKU, &o.Quantity, &o.Status, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return order.Order{}, false, nil
	}
	return o, err == nil, err
}

// Shipments is the shipping.Repository on the shipments table.
type Shipments struct {
	db *sql.DB
}

func NewShipments(db *sql.DB) *Shipments {
	return &Shipments{db: db}
}

func (r *Shipments) Insert(ctx context.Context, s shipping.Shipment) error {
	_, err := sqltx.From(ctx, r.db).ExecContext(ctx,
		`INSERT INTO shipments (order_id, sku, quantity, status) VALUES ($1, $2, $3, $4)`,
		s.OrderID, s.SKU, s.Quantity, s.Status)
	return err
}

func (r *Shipments) DispatchPending(ctx context.Context) (int64, error) {
	res, err := sqltx.From(ctx, r.db).ExecContext(ctx,
		`UPDATE shipments SET status = $1, dispatched_at = CURRENT_TIMESTAMP WHERE status = $2`,
		shipping.StatusDispatched, shipping.StatusPending)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Outbox is the order.Outbox on the outbox_events table, which the worker's producer relays through
// forgo-adapter's outbox.NewPostgres.
type Outbox struct {
	db *sql.DB
}

func NewOutbox(db *sql.DB) *Outbox {
	return &Outbox{db: db}
}

func (o *Outbox) Save(ctx context.Context, msg worker.Message) error {
	_, err := sqltx.From(ctx, o.db).ExecContext(ctx,
		`INSERT INTO outbox_events (id, name, payload, headers, created_at, message_key, destination)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		msg.ID, msg.Name, string(msg.Payload), msg.Headers, msg.CreatedAt, msg.Key, msg.Destination)
	return err
}
