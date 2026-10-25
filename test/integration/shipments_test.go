package integration

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/efegreat/forgo-adapter/sql/sqltx"

	"example/internal/adapter/postgres"
	"example/internal/entity"
	"example/test/support"
)

// DispatchPending dispatches the oldest pending shipments up to its limit, and returns them dispatched.
func TestDispatchPending(t *testing.T) {
	support.Fixture(t, "shipments/pending_shipments.sql")
	shipments := postgres.NewShipments(support.DB())

	first, err := shipments.DispatchPending(t.Context(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0].Status != entity.ShipmentStatusDispatched || first[0].SKU == "" {
		t.Fatalf("first batch %+v", first)
	}
	rest, err := shipments.DispatchPending(t.Context(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) != 1 || rest[0].OrderID == first[0].OrderID {
		t.Fatalf("second batch %+v, want the other pending shipment", rest)
	}
	if none, err := shipments.DispatchPending(t.Context(), 10); err != nil || len(none) != 0 {
		t.Fatalf("third batch %+v, %v", none, err)
	}
}

// Two dispatches at once never take the same shipment: the second skips the shipments the first one's
// transaction holds, and those are pending again when it rolls back.
func TestDispatchPendingSkipsLocked(t *testing.T) {
	support.Fixture(t, "shipments/pending_shipments.sql")
	db := support.DB()
	shipments := postgres.NewShipments(db)
	errRollback := errors.New("rollback")

	held, done := make(chan []entity.Shipment), make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- sqltx.Run(context.Background(), db, func(ctx context.Context) error {
			dispatched, err := shipments.DispatchPending(ctx, 1)
			if err != nil {
				return err
			}
			held <- dispatched
			<-done
			return errRollback
		})
	}()
	first := <-held

	second, err := shipments.DispatchPending(t.Context(), 10)
	close(done)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-result; !errors.Is(err, errRollback) {
		t.Fatal(err)
	}
	if len(first) != 1 || len(second) != 1 || first[0].OrderID == second[0].OrderID {
		t.Fatalf("first %+v, second %+v: want one shipment each", first, second)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE status = 'pending' AND order_id = $1`, first[0].OrderID); n != 1 {
		t.Fatalf("the shipment of the rolled back dispatch is not pending")
	}
}

// Retry sets back the failed messages of one subscriber, or one of them, and leaves the others alone.
func TestRetryFailedMessages(t *testing.T) {
	support.Fixture(t, "inbox/failed_messages.sql")
	failed := postgres.NewFailedMessages(support.DB())

	if n, err := failed.Retry(t.Context(), "order", "missing"); err != nil || n != 0 {
		t.Fatalf("Retry of a missing message: %d, %v", n, err)
	}
	if n, err := failed.Retry(t.Context(), "shipping", ""); err != nil || n != 2 {
		t.Fatalf("Retry of shipping: %d, %v; want its two failed messages", n, err)
	}
	if n, err := failed.Retry(t.Context(), "order", "m4"); err != nil || n != 1 {
		t.Fatalf("Retry of m4: %d, %v", n, err)
	}

	var waiting []string
	rows, err := support.DB().QueryContext(t.Context(),
		`SELECT message_id FROM inbox_events WHERE status = 0 AND attempts = 0 ORDER BY message_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		waiting = append(waiting, id)
	}
	if !slices.Equal(waiting, []string{"m1", "m2", "m4"}) {
		t.Fatalf("waiting %v, want the failed messages, never the completed m3", waiting)
	}
}
