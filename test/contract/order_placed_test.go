package contract

import (
	"net/http"
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/internal/domain/order"
	"example/internal/event"
	"example/test/support"
)

// The order.placed message that POST /v1/orders saves is the one the shipping subscriber handles: its
// name has a handler, and its payload passes the handler's validation. What the test cannot see is the
// broker between them: that RabbitMQ routes the exchange orders to the queue shipping (config.Broker).
func TestOrderPlacedReachesShipping(t *testing.T) {
	support.Fixture(t)

	resp := support.Router(t, api).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 2})
	resp.Status(t, http.StatusOK)
	placed := support.Result[order.PlaceResp](t, resp)

	msgs, err := api.Outbox(t.Context())
	if err != nil || len(msgs) != 1 {
		t.Fatalf("outbox %v, %v", msgs, err)
	}
	got, err := worker.Deliver(t.Context(), "shipping", msgs[0])
	if err != nil || got.Settlement != forgotest.Completed {
		t.Fatalf("delivery %+v, %v", got, err)
	}

	if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE order_id = $1 AND status = 'pending'`, placed.ID); n != 1 {
		t.Fatalf("%d shipments of order %s", n, placed.ID)
	}
}

// The shipment.dispatched message that the dispatch job saves is the one the order subscriber handles: an
// order placed through the API, shipped by the worker and dispatched by the job, reads as shipped.
func TestShipmentDispatchedReachesOrder(t *testing.T) {
	support.Fixture(t)
	resp := support.Router(t, api).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 2})
	resp.Status(t, http.StatusOK)
	placed := support.Result[order.PlaceResp](t, resp)
	deliverAll(t, api, "shipping", event.OrderPlacedName)

	if err := jobs.RunJob(t.Context(), "dispatch_shipments"); err != nil {
		t.Fatal(err)
	}
	deliverAll(t, jobs, "order", event.ShipmentDispatchedName)

	got := support.Result[order.GetResp](t, support.Router(t, api).Get(t, "/v1/orders/"+placed.ID))
	if got.Status != "shipped" {
		t.Fatalf("order %+v, want it shipped", got)
	}
}

// deliverAll delivers to subscriber every message from's outbox holds, which must be one, named name.
func deliverAll(t *testing.T, from *forgotest.Env, subscriber, name string) {
	t.Helper()
	msgs, err := from.Outbox(t.Context())
	if err != nil || len(msgs) != 1 || msgs[0].Name != name {
		t.Fatalf("outbox %+v, %v; want one %s", msgs, err, name)
	}
	if got, err := worker.Deliver(t.Context(), subscriber, msgs[0]); err != nil || got.Settlement != forgotest.Completed {
		t.Fatalf("delivery to %s %+v, %v", subscriber, got, err)
	}
}
