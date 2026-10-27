package features

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	orderv1 "example/api/order/v1"
	"example/internal/errs"
	"example/test/support"
)

func TestGetOrder(t *testing.T) {
	support.Fixture(t, "orders/placed_order.sql")

	got, err := support.Orders(t, env).GetOrder(t.Context(), &orderv1.GetOrderRequest{Id: "0199f0a0-0000-7000-8000-000000000001"})

	if err != nil {
		t.Fatal(err)
	}
	want := &orderv1.GetOrderResponse{Id: "0199f0a0-0000-7000-8000-000000000001", Sku: "a-1", Quantity: 2, Status: "placed"}
	if !proto.Equal(got, want) {
		t.Fatalf("order %v, want %v", got, want)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	support.Fixture(t)

	_, err := support.Orders(t, env).GetOrder(t.Context(), &orderv1.GetOrderRequest{Id: "missing"})

	if code, errCode := support.Status(err); code != codes.NotFound || errCode != errs.ErrOrderNotFound.Code {
		t.Fatalf("status %v, error code %q: %v", code, errCode, err)
	}
}

// Over HTTP, the id is a segment of the path, which no route matches empty. A gRPC client can send an empty
// one, which the validate tag of order.GetReq rejects before order.Service.Get reads it.
func TestGetOrderWithoutID(t *testing.T) {
	support.Fixture(t)

	_, err := support.Orders(t, env).GetOrder(t.Context(), &orderv1.GetOrderRequest{})

	if code, errCode := support.Status(err); code != codes.InvalidArgument || errCode != "validation" {
		t.Fatalf("status %v, error code %q: %v", code, errCode, err)
	}
}
