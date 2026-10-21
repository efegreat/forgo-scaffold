package event

import (
	"encoding/json"

	"github.com/efegreat/forgo/server/worker"
)

// OrdersDestination is the destination of the messages of orders: the exchange orders on RabbitMQ.
const OrdersDestination = "orders"

// OrderPlacedName is the name of an OrderPlaced message. The consumer routes it to its handler by this
// name, and RabbitMQ routes it to the queues bound to this routing key.
const OrderPlacedName = "order.placed"

// OrderPlaced announces a placed order. The shipping handler binds the same struct, so the compiler checks
// that both ends agree on its fields.
type OrderPlaced struct {
	OrderID  string `json:"order_id" validate:"required"`
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"quantity" validate:"min=1"`
}

// Message returns the message of e. Its key is its name, the routing key that the shipping queue is bound
// to.
func (e OrderPlaced) Message() (worker.Message, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return worker.Message{}, err
	}
	return worker.NewMessage(OrderPlacedName, OrderPlacedName, OrdersDestination, payload), nil
}
