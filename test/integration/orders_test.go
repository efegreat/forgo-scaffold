package integration

import (
	"testing"

	"example/internal/order"
	"example/internal/postgres"
	"example/test/support"
)

// The repository against the database alone: no app and no server, the smallest boundary that covers
// the SQL.
func TestOrders(t *testing.T) {
	support.Fixture(t)
	orders := postgres.NewOrders(support.DB())
	o := order.Order{ID: "0199f0a0-0000-7000-8000-000000000001", SKU: "a-1", Quantity: 2, Status: order.StatusPlaced}

	if err := orders.Insert(t.Context(), o); err != nil {
		t.Fatal(err)
	}
	got, ok, err := orders.Get(t.Context(), o.ID)
	if err != nil || !ok {
		t.Fatalf("Get: %v, %t, %v", got, ok, err)
	}
	if got.ID != o.ID || got.SKU != o.SKU || got.Quantity != o.Quantity || got.Status != o.Status || got.CreatedAt.IsZero() {
		t.Fatalf("order %+v", got)
	}

	if _, ok, err := orders.Get(t.Context(), "missing"); err != nil || ok {
		t.Fatalf("Get of a missing order: %t, %v", ok, err)
	}
}
