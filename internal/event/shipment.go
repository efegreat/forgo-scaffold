package event

import (
	"encoding/json"

	"github.com/efegreat/forgo/server/worker"
)

// ShipmentsDestination is the destination of the messages of shipments: the exchange shipments on RabbitMQ.
const ShipmentsDestination = "shipments"

// ShipmentDispatchedName is the name of a ShipmentDispatched message, and the routing key the order queue is
// bound to.
const ShipmentDispatchedName = "shipment.dispatched"

// ShipmentDispatched announces that the shipment of an order was dispatched. The order domain handles it, so
// the two domains exchange messages both ways without importing each other.
type ShipmentDispatched struct {
	OrderID string `json:"order_id" validate:"required"`
}

// Message returns the message of e. Its key is its name, the routing key that the order queue is bound to.
func (e ShipmentDispatched) Message() (worker.Message, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return worker.Message{}, err
	}
	return worker.NewMessage(ShipmentDispatchedName, ShipmentDispatchedName, ShipmentsDestination, payload), nil
}
