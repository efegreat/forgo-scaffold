package config

import (
	"time"

	"github.com/efegreat/forgo"
	fgrpc "github.com/efegreat/forgo/server/grpc"

	orderv1 "example/api/order/v1"
	"example/internal/domain/order"
)

// GRPC serves the orders API over gRPC: order.v1.OrderService, which api/order/v1/order.proto declares. Its
// methods run the very methods of order.Service that the routes of HTTP run, under the same operation
// names, and the server converts the proto messages from and to their requests and responses. Business
// errors reach clients as the status of their operation.Error, with its Code as the reason of the status's
// ErrorInfo.
//
// It drains like HTTP: its health service, grpc.health.v1.Health, answers NOT_SERVING once termination
// starts, while the server keeps serving for Drain. Point a gRPC readiness probe at it. The ops server
// answers the liveness probe.
//
// Unlike POST /v1/orders, PlaceOrder has no rate limit: the clients of a gRPC API are usually other
// services, which a limit per client IP would count together. Add fgrpc.RateLimit to the method when the
// API is public, with a Key of its own to count per caller.
var GRPC = fgrpc.Config{
	ServerConfig: forgo.ServerConfig{
		Term: forgo.TermConfig{TermSig: &forgo.TermSig{Wait: Drain}, TermTimeout: 20 * time.Second},
	},
	Addr:            forgo.Env("GRPC_ADDR", ":50051"),
	SensitiveFields: SensitiveFields,
	Services: []fgrpc.ServiceConfig{
		grpcHealth(),
		{
			Desc: &orderv1.OrderService_ServiceDesc,
			Methods: []fgrpc.MethodConfig{
				{Method: "PlaceOrder", Name: "place_order", Handler: fgrpc.HandlerFor((*order.Service).Place)},
				{Method: "GetOrder", Name: "get_order", Handler: fgrpc.HandlerFor((*order.Service).Get)},
			},
		},
	},
}

// grpcHealth is the health service of GRPC, which logs no line per probe.
func grpcHealth() fgrpc.ServiceConfig {
	health := fgrpc.HealthService()
	health.Args = map[any]any{fgrpc.ArgKeyRecorderDisabled: true}
	return health
}
