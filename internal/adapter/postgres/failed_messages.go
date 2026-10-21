package postgres

import (
	"context"
	"database/sql"

	"github.com/efegreat/forgo-adapter/provider/sqltx"
	"github.com/efegreat/forgo/server/worker"

	"example/internal/port"
)

// failedMessages is the port.FailedMessages on forgo-adapter's inbox_events table, as forgo's REFERENCE.md
// retries failed messages.
type failedMessages struct {
	db *sql.DB
}

func NewFailedMessages(db *sql.DB) port.FailedMessages {
	return &failedMessages{db: db}
}

func (f *failedMessages) Retry(ctx context.Context, subscriber, messageID string) (int64, error) {
	res, err := sqltx.From(ctx, f.db).ExecContext(ctx, `
		UPDATE inbox_events SET status = $1, attempts = 0, updated_at = CURRENT_TIMESTAMP
		WHERE subscriber_name = $2 AND status = $3 AND ($4 = '' OR message_id = $4)`,
		int(worker.InboxWaiting), subscriber, int(worker.InboxFailed), messageID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
