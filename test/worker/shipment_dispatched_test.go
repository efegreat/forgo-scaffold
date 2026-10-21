package worker

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"
	"github.com/efegreat/forgo/server/worker"

	"example/internal/event"
	"example/test/support"
)

const orderID = "0199f0a0-0000-7000-8000-000000000001"

// dispatchedMessage is the message of a dispatched shipment, as the dispatch job saves it.
func dispatchedMessage(t *testing.T, orderID string) worker.Message {
	t.Helper()
	msg, err := event.ShipmentDispatched{OrderID: orderID}.Message()
	if err != nil {
		t.Fatal(err)
	}
	return msg
}

func TestShipmentDispatchedShipsOrder(t *testing.T) {
	support.Fixture(t, "orders/placed_order.sql")

	got, err := env.Deliver(t.Context(), "order", dispatchedMessage(t, orderID))

	if err != nil || got.Settlement != forgotest.Completed {
		t.Fatalf("delivery %+v, %v", got, err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM orders WHERE id = $1 AND status = 'shipped'`, orderID); n != 1 {
		t.Fatalf("%d orders shipped", n)
	}
}

// Retrying cannot help a message about an order that does not exist, or that is shipped already: it is
// rejected at once, to the dead-letter queue, and its order is left as it was.
func TestShipmentDispatchedRejected(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fixtures []string
		shipped  int
	}{
		{"unknown order", nil, 0},
		{"shipped order", []string{"orders/shipped_order.sql"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			support.Fixture(t, tc.fixtures...)

			got, err := env.Deliver(t.Context(), "order", dispatchedMessage(t, orderID))

			if err != nil || got.Settlement != forgotest.Rejected {
				t.Fatalf("delivery %+v, %v", got, err)
			}
			if n := support.Count(t, `SELECT COUNT(*) FROM orders WHERE status = 'shipped'`); n != tc.shipped {
				t.Fatalf("%d orders shipped, want %d", n, tc.shipped)
			}
		})
	}
}

// A message whose order did not exist yet fails. Once the order exists, the message is still rejected on
// every redelivery, since the inbox keeps it failed, until retry-failed sets it back to waiting: then its
// replay from the dead-letter queue ships the order.
func TestRetryFailedMessage(t *testing.T) {
	support.Fixture(t)
	msg := dispatchedMessage(t, orderID)
	if got, err := env.Deliver(t.Context(), "order", msg); err != nil || got.Settlement != forgotest.Rejected {
		t.Fatalf("first delivery %+v, %v", got, err)
	}
	support.Load(t, "orders/placed_order.sql")
	if got, err := env.Deliver(t.Context(), "order", msg); err != nil || got.Settlement != forgotest.Rejected {
		t.Fatalf("replay before retry-failed %+v, %v", got, err)
	}

	if err := support.Command("retry-failed", "-subscriber=order", "-message="+msg.ID); err != nil {
		t.Fatal(err)
	}

	if got, err := env.Deliver(t.Context(), "order", msg); err != nil || got.Settlement != forgotest.Completed {
		t.Fatalf("replay after retry-failed %+v, %v", got, err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM orders WHERE id = $1 AND status = 'shipped'`, orderID); n != 1 {
		t.Fatalf("%d orders shipped", n)
	}
}
