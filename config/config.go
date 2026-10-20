// Package config declares the resources and servers of the application.
//
// The resources are functions, not vars: Go evaluates package-level vars before TestMain, so the address
// of a test's database container, which TestMain puts in DATABASE_URL, could never reach them. The
// servers can stay vars, since tests serve them on ports of their own.
package config

import (
	"time"

	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/resource/broker/rabbitmq"
	"github.com/efegreat/forgo-adapter/resource/sqldb"
	"github.com/efegreat/forgo/logger"
	"github.com/efegreat/forgo/server/cron"
	"github.com/efegreat/forgo/server/http"
	"github.com/efegreat/forgo/server/worker"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"

	"example/internal/order"
	"example/internal/shipping"
)

// Logger writes JSON lines to stdout.
var Logger = logger.Config{Level: zerolog.InfoLevel}

// DB is the PostgreSQL database of the orders and shipments, and of the outbox and the inbox.
func DB() sqldb.Config {
	return sqldb.Config{
		Name:           "main",
		DriverName:     "postgres",
		DSN:            forgo.Env("DATABASE_URL", "postgres://example:example@localhost:5432/example?sslmode=disable"),
		ConnectTimeout: 5 * time.Second,
	}
}

// Broker is the RabbitMQ broker of the worker. It declares the topology: order.placed messages go to the
// exchange orders, and from there to the queue shipping, which the subscriber shipping consumes. Tests do
// not run it, so they do not check this topology either: check it against a real broker.
func Broker() rabbitmq.Config {
	return rabbitmq.Config{
		Name:     "rabbitmq",
		Addr:     forgo.Env("RABBITMQ_HOST", "localhost"),
		Username: forgo.Env("RABBITMQ_USER", "guest"),
		Password: forgo.Env("RABBITMQ_PASSWORD", "guest"),
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: order.Destination, Kind: amqp.ExchangeTopic, Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{
				Name:       "shipping",
				Durable:    true,
				Bindings:   []rabbitmq.BindingConfig{{Exchange: order.Destination, RoutingKey: order.PlacedMessage}},
				DeadLetter: rabbitmq.DeadLetterConfig{Enabled: true},
				Retry:      rabbitmq.RetryConfig{Initial: time.Second},
			},
		},
		Subscribers: []rabbitmq.SubscriberConfig{{Name: "shipping", Queue: "shipping"}},
	}
}

// HTTP serves the orders API.
var HTTP = http.Config{
	Addr: forgo.Env("HTTP_ADDR", ":8080"),
	Routes: []http.RouteConfig{
		{Name: "health", Method: "GET", Path: "/health", Handler: http.HealthHandler()},
		{Name: "place_order", Method: "POST", Path: "/orders", Version: "v1", Handler: http.HandlerFor((*order.Service).Place)},
		{Name: "get_order", Method: "GET", Path: "/orders/{id}", Version: "v1", Handler: http.HandlerFor((*order.Service).Get)},
	},
}

// Cron dispatches the pending shipments every five minutes.
var Cron = cron.Config{
	Jobs: []cron.JobConfig{
		{Name: "dispatch_shipments", Schedule: "@every 5m", Job: cron.As(shipping.NewDispatchJob)},
	},
}

// Consumer handles the messages of the queue shipping.
var Consumer = worker.ConsumerConfig{
	Subscribers: []worker.SubscriberConfig{
		{
			Name: "shipping", // rabbitmq.SubscriberConfig.Name
			Handlers: []worker.HandlerConfig{
				{Name: order.PlacedMessage, Handler: worker.HandlerFor((*shipping.Service).OnOrderPlaced)},
			},
		},
	},
}

// Producer relays the outbox to the broker, every message through the default publisher.
var Producer = worker.ProducerConfig{}
