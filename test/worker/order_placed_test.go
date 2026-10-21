package worker

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"
	"github.com/efegreat/forgo/server/worker"

	"example/internal/event"
	"example/test/support"
)

// placedMessage is the message of a placed order, as order.Service.Place saves it.
func placedMessage(t *testing.T, e event.OrderPlaced) worker.Message {
	t.Helper()
	msg, err := e.Message()
	if err != nil {
		t.Fatal(err)
	}
	return msg
}

func TestOrderPlacedShips(t *testing.T) {
	support.Fixture(t)
	msg := placedMessage(t, event.OrderPlaced{OrderID: orderID, SKU: "a-1", Quantity: 2})

	got, err := env.Deliver(t.Context(), "shipping", msg)

	if err != nil || got.Settlement != forgotest.Completed {
		t.Fatalf("delivery %+v, %v", got, err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE order_id = $1 AND sku = 'a-1' AND quantity = 2 AND status = 'pending'`, orderID); n != 1 {
		t.Fatalf("%d shipments", n)
	}
}

// A message delivered twice, as brokers do, ships its order once: the inbox skips the second delivery.
func TestOrderPlacedTwiceShipsOnce(t *testing.T) {
	support.Fixture(t)
	msg := placedMessage(t, event.OrderPlaced{OrderID: orderID, SKU: "a-1", Quantity: 2})

	for range 2 {
		if got, err := env.Deliver(t.Context(), "shipping", msg); err != nil || got.Settlement != forgotest.Completed {
			t.Fatalf("delivery %+v, %v", got, err)
		}
	}

	if n := support.Count(t, `SELECT COUNT(*) FROM shipments`); n != 1 {
		t.Fatalf("%d shipments", n)
	}
}

// A message that fails validation is rejected, to the dead-letter queue in production, and never retried.
func TestOrderPlacedInvalid(t *testing.T) {
	support.Fixture(t)
	msg := placedMessage(t, event.OrderPlaced{SKU: "a-1", Quantity: 2}) // no order ID

	got, err := env.Deliver(t.Context(), "shipping", msg)

	if err != nil || got.Settlement != forgotest.Rejected {
		t.Fatalf("delivery %+v, %v", got, err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM shipments`); n != 0 {
		t.Fatalf("%d shipments", n)
	}
}
