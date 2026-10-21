package config

import (
	"github.com/efegreat/forgo/server/worker"

	"example/internal/domain/shipping"
	"example/internal/event"
)

// Consumer handles the messages of the queue shipping.
var Consumer = worker.ConsumerConfig{
	Subscribers: []worker.SubscriberConfig{
		{
			Name: "shipping", // rabbitmq.SubscriberConfig.Name
			Handlers: []worker.HandlerConfig{
				{Name: event.OrderPlacedName, Handler: worker.HandlerFor((*shipping.Service).OnOrderPlaced)},
			},
		},
	},
}

// Producer relays the outbox to the broker, every message through the default publisher.
var Producer = worker.ProducerConfig{}
