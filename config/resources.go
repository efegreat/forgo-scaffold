package config

import (
	"time"

	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/resource/broker/rabbitmq"
	"github.com/efegreat/forgo-adapter/resource/sqldb"
	amqp "github.com/rabbitmq/amqp091-go"

	"example/internal/event"
)

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
			{Name: event.OrdersDestination, Kind: amqp.ExchangeTopic, Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			{
				Name:       "shipping",
				Durable:    true,
				Bindings:   []rabbitmq.BindingConfig{{Exchange: event.OrdersDestination, RoutingKey: event.OrderPlacedName}},
				DeadLetter: rabbitmq.DeadLetterConfig{Enabled: true},
				Retry:      rabbitmq.RetryConfig{Initial: time.Second},
			},
		},
		Subscribers: []rabbitmq.SubscriberConfig{{Name: "shipping", Queue: "shipping"}},
	}
}
