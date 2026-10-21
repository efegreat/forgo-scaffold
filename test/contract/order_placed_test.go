package contract

import (
	"net/http"
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/internal/domain/order"
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
