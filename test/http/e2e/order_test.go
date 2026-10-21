package e2e

import (
	"net/http"
	"testing"

	"example/internal/domain/order"
	"example/test/support"
)

// A client places an order, then reads it back, over the network.
func TestPlaceAndGetOrder(t *testing.T) {
	support.Fixture(t)
	client := support.Live(t, env)

	id := placeOrder(t, client, order.PlaceReq{SKU: "a-1", Quantity: 2})
	got := getOrder(t, client, id)

	if got.SKU != "a-1" || got.Quantity != 2 || got.Status != "placed" {
		t.Fatalf("order %+v", got)
	}
}

func placeOrder(t *testing.T, client support.Client, req order.PlaceReq) string {
	t.Helper()
	resp := client.Post(t, "/v1/orders", req)
	resp.Status(t, http.StatusOK)
	return support.Result[order.PlaceResp](t, resp).ID
}

func getOrder(t *testing.T, client support.Client, id string) order.GetResp {
	t.Helper()
	resp := client.Get(t, "/v1/orders/"+id)
	resp.Status(t, http.StatusOK)
	return support.Result[order.GetResp](t, resp)
}
