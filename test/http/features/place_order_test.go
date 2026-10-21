package features

import (
	"encoding/json"
	"net/http"
	"testing"

	"example/internal/domain/order"
	"example/internal/event"
	"example/test/support"
)

func TestPlaceOrder(t *testing.T) {
	support.Fixture(t)

	resp := support.Router(t, env).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 2})

	resp.Status(t, http.StatusOK)
	placed := support.Result[order.PlaceResp](t, resp)
	if n := support.Count(t, `SELECT COUNT(*) FROM orders WHERE id = $1 AND sku = 'a-1' AND quantity = 2 AND status = 'placed'`, placed.ID); n != 1 {
		t.Fatalf("%d orders %s stored", n, placed.ID)
	}

	// The order is announced: the outbox holds the message the worker's producer will publish.
	msgs, err := env.Outbox(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Name != event.OrderPlacedName || msgs[0].Destination != event.OrdersDestination {
		t.Fatalf("outbox %+v", msgs)
	}
	var e event.OrderPlaced
	if err := json.Unmarshal(msgs[0].Payload, &e); err != nil {
		t.Fatal(err)
	}
	if e != (event.OrderPlaced{OrderID: placed.ID, SKU: "a-1", Quantity: 2}) {
		t.Fatalf("payload %+v", e)
	}
}

// An order that fails validation is neither stored nor announced.
func TestPlaceOrderInvalid(t *testing.T) {
	support.Fixture(t)

	resp := support.Router(t, env).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 0})

	resp.Status(t, http.StatusBadRequest)
	if resp.ErrCode() != "validation" {
		t.Fatalf("error code %q", resp.ErrCode())
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM orders`); n != 0 {
		t.Fatalf("%d orders stored", n)
	}
	if msgs, err := env.Outbox(t.Context()); err != nil || len(msgs) != 0 {
		t.Fatalf("outbox %v, %v", msgs, err)
	}
}
