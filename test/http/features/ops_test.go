package features

import (
	"net/http"
	"testing"

	"example/internal/domain/order"
	"example/test/support"
)

// The API answers its readiness probe, and the ops server the liveness probe.
func TestProbes(t *testing.T) {
	support.Router(t, env).Get(t, "/ready").Status(t, http.StatusOK)
	support.Ops(t, env).Get(t, "/health").Status(t, http.StatusOK)
}

// The ops server serves the metrics of the API's operations, for Prometheus.
func TestMetrics(t *testing.T) {
	support.Fixture(t)
	support.Router(t, env).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 2}).Status(t, http.StatusOK)

	resp := support.Ops(t, env).Get(t, "/metrics")

	resp.Status(t, http.StatusOK)
	if !resp.HasSample("forgo_operation_duration_seconds_count",
		`forgo_operation_name="place_order"`, `forgo_operation_status="ok"`, `forgo_operation_type="http_handler"`) {
		t.Fatalf("no count of place_order in:\n%s", resp.Text())
	}
}

// The ops server lists the live configs with their versions, saves a change, and rejects one that fails
// validation or started from a version that is no longer stored.
func TestLiveConfigs(t *testing.T) {
	support.Fixture(t)
	ops := support.Ops(t, env)
	type stored struct {
		Version uint64       `json:"version"`
		Value   order.Limits `json:"value"`
	}

	resp := ops.Get(t, "/live-configs")
	resp.Status(t, http.StatusOK)
	if got := support.Result[map[string]stored](t, resp)["order_limits"]; got != (stored{Value: order.Limits{}.Default()}) {
		t.Fatalf("order_limits %+v, want the defaults, not stored", got)
	}

	invalid := ops.Put(t, "/live-configs/order_limits", order.Limits{MaxQuantity: 0}, http.Header{"If-Match": {"0"}})
	invalid.Status(t, http.StatusBadRequest)
	if invalid.ErrCode() != "validation" {
		t.Fatalf("error code %q", invalid.ErrCode())
	}

	ops.Put(t, "/live-configs/order_limits", order.Limits{MaxQuantity: 5}, http.Header{"If-Match": {"0"}}).Status(t, http.StatusOK)
	stale := ops.Put(t, "/live-configs/order_limits", order.Limits{MaxQuantity: 7}, http.Header{"If-Match": {"0"}})
	stale.Status(t, http.StatusPreconditionFailed)
	if stale.ErrCode() != "live_config_conflict" {
		t.Fatalf("error code %q", stale.ErrCode())
	}

	if got := support.Result[map[string]stored](t, ops.Get(t, "/live-configs"))["order_limits"]; got != (stored{Version: 1, Value: order.Limits{MaxQuantity: 5}}) {
		t.Fatalf("order_limits %+v, want the change saved as version 1", got)
	}
}
