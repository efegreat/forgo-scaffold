package features

import (
	"net/http"
	"testing"

	"example/internal/order"
	"example/test/support"
)

func TestGetOrder(t *testing.T) {
	support.Fixture(t, "orders/placed_order.sql")

	resp := support.Router(t, env).Get(t, "/v1/orders/0199f0a0-0000-7000-8000-000000000001")

	resp.Status(t, http.StatusOK)
	got := support.Result[order.GetResp](t, resp)
	want := order.GetResp{ID: "0199f0a0-0000-7000-8000-000000000001", SKU: "a-1", Quantity: 2, Status: order.StatusPlaced}
	if got != want {
		t.Fatalf("order %+v, want %+v", got, want)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	support.Fixture(t)

	resp := support.Router(t, env).Get(t, "/v1/orders/missing")

	resp.Status(t, http.StatusNotFound)
	if resp.ErrCode() != order.ErrNotFound.Code {
		t.Fatalf("error code %q", resp.ErrCode())
	}
}
