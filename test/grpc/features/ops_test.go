package features

import (
	"net/http"
	"testing"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	orderv1 "example/api/order/v1"
	"example/test/support"
)

// The gRPC server answers its readiness probe, the health service, and the ops server the liveness probe.
func TestProbes(t *testing.T) {
	resp, err := healthpb.NewHealthClient(support.Conn(t, env)).Check(t.Context(), &healthpb.HealthCheckRequest{})
	if err != nil || resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("health %v, %v", resp, err)
	}
	support.Ops(t, env).Get(t, "/health").Status(t, http.StatusOK)
}

// The ops server serves the metrics of the gRPC server's operations, for Prometheus, under the names of the
// HTTP routes that run the same methods.
func TestMetrics(t *testing.T) {
	support.Fixture(t)
	if _, err := support.Orders(t, env).PlaceOrder(t.Context(), &orderv1.PlaceOrderRequest{Sku: "a-1", Quantity: 2}); err != nil {
		t.Fatal(err)
	}

	resp := support.Ops(t, env).Get(t, "/metrics")

	resp.Status(t, http.StatusOK)
	if !resp.HasSample("forgo_operation_duration_seconds_count",
		`forgo_operation_name="place_order"`, `forgo_operation_status="ok"`, `forgo_operation_type="grpc_handler"`) {
		t.Fatalf("no count of place_order in:\n%s", resp.Text())
	}
}
