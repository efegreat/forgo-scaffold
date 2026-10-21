package config

import (
	"time"

	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo/server/http"

	"example/internal/domain/order"
)

// Drain is how long the API keeps serving once it starts terminating, while load balancers stop routing to
// it: longer than the readiness probe takes to fail, and than the proxies take to drop the instance.
const Drain = 10 * time.Second

// HTTP serves the orders API.
//
// Its readiness probe, GET /ready, fails once termination starts, while the server keeps serving for Drain,
// so that load balancers stop routing to the instance before it stops: deployments then fail no request
// (forgo's REFERENCE.md, "Zero-downtime deployments on Kubernetes"). The ops server answers the liveness
// probe.
var HTTP = http.Config{
	ServerConfig: forgo.ServerConfig{
		Term: forgo.TermConfig{TermSig: &forgo.TermSig{Wait: Drain}, TermTimeout: 20 * time.Second},
	},
	Addr:            forgo.Env("HTTP_ADDR", ":8080"),
	SensitiveFields: SensitiveFields,
	Routes: []http.RouteConfig{
		{Name: "ready", Method: "GET", Path: "/ready", Handler: http.HealthHandler(), DisableAPIDoc: true,
			Args: map[any]any{http.ArgKeyRecorderDisabled: true}},
		{Name: "place_order", Method: "POST", Path: "/orders", Version: "v1", Handler: http.HandlerFor((*order.Service).Place),
			Middlewares: []forgo.MiddlewareFactory{placeRateLimit}},
		{Name: "get_order", Method: "GET", Path: "/orders/{id}", Version: "v1", Handler: http.HandlerFor((*order.Service).Get)},
	},
}

// placeRateLimit admits 60 orders a minute per client IP, 10 of them at once. It counts per instance: with
// several instances, count in Redis with http.RateLimitWith and forgo-adapter's ratelimit.NewRedis. Behind
// a reverse proxy, set TrustedProxies, or every client counts as the proxy.
var placeRateLimit = http.RateLimit(http.RateLimitConfig{Limit: 60, Period: time.Minute, Burst: 10})
