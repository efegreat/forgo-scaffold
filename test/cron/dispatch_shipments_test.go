package cron

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"example/config"
	"example/internal/event"
	"example/test/support"
)

// A run dispatches every pending shipment, and saves a shipment.dispatched message for each, for the order
// subscriber.
func TestDispatchShipments(t *testing.T) {
	support.Fixture(t, "shipments/pending_shipments.sql")

	if err := env.RunJob(t.Context(), "dispatch_shipments"); err != nil {
		t.Fatal(err)
	}

	if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE status = 'pending'`); n != 0 {
		t.Fatalf("%d shipments still pending", n)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE status = 'dispatched' AND dispatched_at IS NOT NULL`); n != 3 {
		t.Fatalf("%d shipments dispatched, want 3", n)
	}
	msgs, err := env.Outbox(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var orders []string
	for _, msg := range msgs {
		var e event.ShipmentDispatched
		if err := json.Unmarshal(msg.Payload, &e); err != nil {
			t.Fatal(err)
		}
		if msg.Name != event.ShipmentDispatchedName || msg.Key != event.ShipmentDispatchedName || msg.Destination != event.ShipmentsDestination {
			t.Fatalf("message %+v", msg)
		}
		orders = append(orders, e.OrderID)
	}
	slices.Sort(orders)
	if want := []string{"0199f0a0-0000-7000-8000-000000000001", "0199f0a0-0000-7000-8000-000000000002"}; !slices.Equal(orders, want) {
		t.Fatalf("messages of orders %v, want the two that were pending", orders)
	}
}

// While another replica holds the job's lock, a run is skipped: it dispatches nothing. Once the lock is
// released, the next run dispatches.
func TestDispatchShipmentsLocked(t *testing.T) {
	support.Fixture(t, "shipments/pending_shipments.sql")

	t.Run("held", func(t *testing.T) {
		support.HoldLock(t, config.Cron.Jobs[0].Lock.Key, time.Minute)
		if err := env.RunJob(t.Context(), "dispatch_shipments"); err != nil {
			t.Fatal(err)
		}
		if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE status = 'pending'`); n != 2 {
			t.Fatalf("%d shipments pending, want both: the run should have been skipped", n)
		}
	})

	if err := env.RunJob(t.Context(), "dispatch_shipments"); err != nil {
		t.Fatal(err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE status = 'pending'`); n != 0 {
		t.Fatalf("%d shipments still pending after the lock was released", n)
	}
}
