package config

import (
	"time"

	"github.com/efegreat/forgo"
	"github.com/efegreat/forgo-adapter/otel/meter"
	"github.com/efegreat/forgo-adapter/rabbitmq"
	"github.com/efegreat/forgo-adapter/redis/redisdb"
	"github.com/efegreat/forgo-adapter/sql/sqldb"
	amqp "github.com/rabbitmq/amqp091-go"

	"example/internal/event"
)

// DB is the PostgreSQL database of the orders and shipments, the outbox and the inbox, and the live configs.
// It logs the statements that are slow or fail, with their arguments masked like the other logs.
func DB() sqldb.Config {
	return sqldb.Config{
		Name:           "main",
		DriverName:     "postgres",
		DSN:            forgo.Env("DATABASE_URL", "postgres://example:example@localhost:5432/example?sslmode=disable"),
		ConnectTimeout: 5 * time.Second,
		Logger:         sqldb.LoggerConfig{Enabled: true, SlowOnly: true, Args: true, SensitiveFields: SensitiveFields},
	}
}

// Broker is the RabbitMQ broker of the worker. It declares the topology, one exchange per domain that
// publishes and one queue per domain that handles:
//
//   - order.placed messages go to the exchange orders, and from there to the queue shipping;
//   - shipment.dispatched messages go to the exchange shipments, and from there to the queue order.
//
// Each queue has its subscriber, of the same name. Tests do not run the broker, so they do not check this
// topology either: check it against a real broker.
func Broker() rabbitmq.Config {
	queue := func(name, exchange, key string) rabbitmq.QueueConfig {
		return rabbitmq.QueueConfig{
			Name:       name,
			Durable:    true,
			Bindings:   []rabbitmq.BindingConfig{{Exchange: exchange, RoutingKey: key}},
			DeadLetter: rabbitmq.DeadLetterConfig{Enabled: true},
			Retry:      rabbitmq.RetryConfig{Initial: time.Second},
		}
	}
	return rabbitmq.Config{
		Name:     "rabbitmq",
		Addr:     forgo.Env("RABBITMQ_HOST", "localhost"),
		Username: forgo.Env("RABBITMQ_USER", "guest"),
		Password: forgo.Env("RABBITMQ_PASSWORD", "guest"),
		Exchanges: []rabbitmq.ExchangeConfig{
			{Name: event.OrdersDestination, Kind: amqp.ExchangeTopic, Durable: true},
			{Name: event.ShipmentsDestination, Kind: amqp.ExchangeTopic, Durable: true},
		},
		Queues: []rabbitmq.QueueConfig{
			queue("shipping", event.OrdersDestination, event.OrderPlacedName),
			queue("order", event.ShipmentsDestination, event.ShipmentDispatchedName),
		},
		Subscribers: []rabbitmq.SubscriberConfig{
			{Name: "shipping", Queue: "shipping"},
			{Name: "order", Queue: "order"},
		},
	}
}

// Redis holds the locks of the scheduled jobs, so that two replicas of the cron deployable never run a job
// at once.
func Redis() redisdb.Config {
	return redisdb.Config{Name: "redis", Addrs: []string{forgo.Env("REDIS_ADDR", "localhost:6379")}}
}

// Meter records the metrics of the deployable named service, which its ops server serves to Prometheus.
func Meter(service string) meter.Config {
	return meter.Config{ServiceName: service, Exporter: meter.ExporterPrometheus}
}
