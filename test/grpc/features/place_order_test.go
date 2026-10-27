package features

import (
	"encoding/json"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"

	orderv1 "example/api/order/v1"
	"example/internal/domain/order"
	"example/internal/errs"
	"example/internal/event"
	"example/test/support"
)

func TestPlaceOrder(t *testing.T) {
	support.Fixture(t)

	resp, err := support.Orders(t, env).PlaceOrder(t.Context(), &orderv1.PlaceOrderRequest{Sku: "a-1", Quantity: 2})

	if err != nil {
		t.Fatal(err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM orders WHERE id = $1 AND sku = 'a-1' AND quantity = 2 AND status = 'placed'`, resp.GetId()); n != 1 {
		t.Fatalf("%d orders %s stored", n, resp.GetId())
	}

	// The order is announced as over HTTP: the outbox holds the message the worker's producer will publish.
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
	if e != (event.OrderPlaced{OrderID: resp.GetId(), SKU: "a-1", Quantity: 2}) {
		t.Fatalf("payload %+v", e)
	}
}

// An order that fails validation, the validate tags of order.PlaceReq, is neither stored nor announced.
func TestPlaceOrderInvalid(t *testing.T) {
	support.Fixture(t)

	_, err := support.Orders(t, env).PlaceOrder(t.Context(), &orderv1.PlaceOrderRequest{Sku: "a-1", Quantity: 0})

	if code, errCode := support.Status(err); code != codes.InvalidArgument || errCode != "validation" {
		t.Fatalf("status %v, error code %q: %v", code, errCode, err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM orders`); n != 0 {
		t.Fatalf("%d orders stored", n)
	}
	if msgs, err := env.Outbox(t.Context()); err != nil || len(msgs) != 0 {
		t.Fatalf("outbox %v, %v", msgs, err)
	}
}

// An admin lowers the limit of an order through the ops server, and the API applies it to the next order:
// above the limit, an order fails with the business error's code and reason, and is neither stored nor
// announced.
func TestPlaceOrderAboveLimit(t *testing.T) {
	support.Fixture(t)
	support.Ops(t, env).Put(t, "/live-configs/order_limits", order.Limits{MaxQuantity: 2}, http.Header{"If-Match": {"0"}}).
		Status(t, http.StatusOK)
	client := support.Orders(t, env)

	_, err := client.PlaceOrder(t.Context(), &orderv1.PlaceOrderRequest{Sku: "a-1", Quantity: 3})

	if code, errCode := support.Status(err); code != codes.InvalidArgument || errCode != errs.ErrQuantityAboveLimit.Code {
		t.Fatalf("status %v, error code %q: %v", code, errCode, err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM orders`); n != 0 {
		t.Fatalf("%d orders stored", n)
	}
	if msgs, err := env.Outbox(t.Context()); err != nil || len(msgs) != 0 {
		t.Fatalf("outbox %v, %v", msgs, err)
	}
	if _, err := client.PlaceOrder(t.Context(), &orderv1.PlaceOrderRequest{Sku: "a-1", Quantity: 2}); err != nil {
		t.Fatal(err)
	}
}
