package config

import (
	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo/server/http"

	"example/internal/domain/order"
)

// HTTP serves the orders API.
var HTTP = http.Config{
	Addr: forgo.Env("HTTP_ADDR", ":8080"),
	Routes: []http.RouteConfig{
		{Name: "health", Method: "GET", Path: "/health", Handler: http.HealthHandler()},
		{Name: "place_order", Method: "POST", Path: "/orders", Version: "v1", Handler: http.HandlerFor((*order.Service).Place)},
		{Name: "get_order", Method: "GET", Path: "/orders/{id}", Version: "v1", Handler: http.HandlerFor((*order.Service).Get)},
	},
}
