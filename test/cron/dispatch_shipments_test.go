package cron

import (
	"testing"

	"example/test/support"
)

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
}
