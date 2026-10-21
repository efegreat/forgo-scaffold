package postgres

import (
	"context"
	"database/sql"

	"github.com/efegreat/forgo-adapter/provider/sqltx"
	"github.com/efegreat/forgo/server/worker"

	"example/internal/port"
)

// outbox is the port.Outbox on the outbox_events table, which the worker's producer relays through
// forgo-adapter's outbox.NewPostgres.
type outbox struct {
	db *sql.DB
}

func NewOutbox(db *sql.DB) port.Outbox {
	return &outbox{db: db}
}

func (o *outbox) Save(ctx context.Context, msg worker.Message) error {
	_, err := sqltx.From(ctx, o.db).ExecContext(ctx,
		`INSERT INTO outbox_events (id, name, payload, headers, created_at, message_key, destination)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		msg.ID, msg.Name, string(msg.Payload), msg.Headers, msg.CreatedAt, msg.Key, msg.Destination)
	return err
}
