package config

import (
	"github.com/efegreat/forgo/server/worker"

	"example/internal/domain/order"
	"example/internal/domain/shipping"
	"example/internal/event"
)

// Consumer handles the messages of the queues shipping and order, through the subscribers of Broker.
var Consumer = worker.ConsumerConfig{
	SensitiveFields: SensitiveFields,
	Subscribers: []worker.SubscriberConfig{
		{
			Name: "shipping", // rabbitmq.SubscriberConfig.Name
			Handlers: []worker.HandlerConfig{
				{Name: event.OrderPlacedName, Handler: worker.HandlerFor((*shipping.Service).OnOrderPlaced)},
			},
		},
		{
			Name: "order",
			Handlers: []worker.HandlerConfig{
				{Name: event.ShipmentDispatchedName, Handler: worker.HandlerFor((*order.Service).OnShipmentDispatched)},
			},
		},
	},
}

// Producer relays the outbox to the broker, every message through the default publisher.
var Producer = worker.ProducerConfig{}
