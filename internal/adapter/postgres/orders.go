package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/efegreat/forgo-adapter/provider/sqltx"

	"example/internal/entity"
	"example/internal/port"
)

// orders is the port.OrderRepository on the orders table.
type orders struct {
	db *sql.DB
}

func NewOrders(db *sql.DB) port.OrderRepository {
	return &orders{db: db}
}

func (r *orders) Insert(ctx context.Context, o entity.Order) error {
	_, err := sqltx.From(ctx, r.db).ExecContext(ctx,
		`INSERT INTO orders (id, sku, quantity, status) VALUES ($1, $2, $3, $4)`,
		o.ID, o.SKU, o.Quantity, o.Status)
	return err
}

func (r *orders) Get(ctx context.Context, id string) (entity.Order, bool, error) {
	var o entity.Order
	err := sqltx.From(ctx, r.db).QueryRowContext(ctx,
		`SELECT id, sku, quantity, status, created_at FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.SKU, &o.Quantity, &o.Status, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Order{}, false, nil
	}
	return o, err == nil, err
}

func (r *orders) Update(ctx context.Context, o entity.Order) error {
	_, err := sqltx.From(ctx, r.db).ExecContext(ctx, `UPDATE orders SET status = $2 WHERE id = $1`, o.ID, o.Status)
	return err
}
