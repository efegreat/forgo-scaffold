package config

import (
	"time"

	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/otel/meter"
	"github.com/efegreat/forgo/server/http"

	"example/internal/domain/order"
)

// LiveConfigs are the live configs that admins read and change through the ops server.
var LiveConfigs = forgo.LiveConfigCatalog{
	forgo.LiveConfigOf[order.Limits](),
}

// Ops is the operations server of every deployable, on addr unless OPS_ADDR is set, which the cluster must
// not expose. It answers:
//
//   - GET /health, the liveness probe, which keeps passing while the deployable drains;
//   - GET /metrics, the metrics of the deployable, for Prometheus;
//   - GET /live-configs and PUT /live-configs/{name}, which read and change the LiveConfigs.
//
// Once its deployable terminates, it keeps serving for drain: the API's Drain in the http deployable, so that
// the liveness probe keeps passing and Prometheus scrapes the last metrics while the API drains, and zero in
// the deployables that stop at once.
//
// Anyone who reaches the port reads every metric and changes every live config: before exposing it, protect
// its routes with middlewares, such as http.JWTAuth and a role check.
func Ops(addr string, drain time.Duration) http.Config {
	var term forgo.TermConfig
	if drain > 0 {
		term.TermSig = &forgo.TermSig{Wait: drain}
	}
	return http.Config{
		ServerConfig:    forgo.ServerConfig{Name: "ops", Term: term},
		Addr:            forgo.Env("OPS_ADDR", addr),
		SensitiveFields: SensitiveFields,
		Routes: []http.RouteConfig{
			{Name: "health", Method: "GET", Path: "/health", Handler: http.EmptyHandler(), DisableAPIDoc: true,
				Args: map[any]any{http.ArgKeyRecorderDisabled: true}},
			meter.PrometheusRoute(),
			{Name: "live_configs", Method: "GET", Path: "/live-configs", Handler: http.LiveConfigsHandler(LiveConfigs), DisableAPIDoc: true},
			{Name: "set_live_config", Method: "PUT", Path: "/live-configs/{name}", Handler: http.SetLiveConfigHandler(LiveConfigs), DisableAPIDoc: true},
		},
	}
}
