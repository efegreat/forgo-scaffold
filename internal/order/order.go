// Package order places orders. Each order is announced with an order.placed message, which the
// service saves to the outbox in the transaction of the order, so the message is published if and only
// if the order is stored.
package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/efegreat/forgo/operation"
	"github.com/efegreat/forgo/server/worker"
	"github.com/google/uuid"
)

const (
	// PlacedMessage is the name of the message of a placed order. The consumer routes it to its handler
	// by this name, and RabbitMQ routes it to the queues bound to this routing key.
	PlacedMessage = "order.placed"
	// Destination is the exchange the messages of orders are published to.
	Destination = "orders"
)

// StatusPlaced is the status of an order that was placed.
const StatusPlaced = "placed"

// ErrNotFound is the error of an order that does not exist.
var ErrNotFound = operation.Error{
	Code:     "order_not_found",
	Status:   operation.StatusNotFound,
	Message:  "order not found",
	Severity: operation.SeverityLight,
}

// Order is an order of a quantity of one SKU.
type Order struct {
	ID        string
	SKU       string
	Quantity  int
	Status    string
	CreatedAt time.Time
}

// Placed is the payload of a PlacedMessage. The shipping handler binds the same struct, so the compiler
// checks that both ends agree on its fields.
type Placed struct {
	OrderID  string `json:"order_id" validate:"required"`
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"quantity" validate:"min=1"`
}

// Repository stores orders.
type Repository interface {
	Insert(ctx context.Context, o Order) error
	// Get returns the order of id, and false when there is none.
	Get(ctx context.Context, id string) (Order, bool, error)
}

// Outbox saves the messages to publish, in the transaction of ctx.
type Outbox interface {
	Save(ctx context.Context, msg worker.Message) error
}

// Service places and reads orders.
type Service struct {
	tx     operation.Transactor
	orders Repository
	outbox Outbox
}

// NewService returns a Service that stores orders in orders and saves their messages to outbox, in
// transactions of tx.
func NewService(tx operation.Transactor, orders Repository, outbox Outbox) *Service {
	return &Service{tx: tx, orders: orders, outbox: outbox}
}

type PlaceReq struct {
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"quantity" validate:"min=1"`
}

type PlaceResp struct {
	ID string `json:"id"`
}

// Place stores the order and saves its PlacedMessage to the outbox, in one transaction.
func (s *Service) Place(ctx context.Context, req PlaceReq) (PlaceResp, error) {
	o := Order{ID: uuid.Must(uuid.NewV7()).String(), SKU: req.SKU, Quantity: req.Quantity, Status: StatusPlaced}
	payload, err := json.Marshal(Placed{OrderID: o.ID, SKU: o.SKU, Quantity: o.Quantity})
	if err != nil {
		return PlaceResp{}, err
	}
	// The key is the routing key on RabbitMQ: the shipping queue is bound to PlacedMessage.
	msg := worker.NewMessage(PlacedMessage, PlacedMessage, Destination, payload)

	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.orders.Insert(ctx, o); err != nil {
			return err
		}
		return s.outbox.Save(ctx, msg)
	})
	if err != nil {
		return PlaceResp{}, err
	}
	return PlaceResp{ID: o.ID}, nil
}

type GetReq struct {
	ID string `uri:"id" validate:"required"`
}

type GetResp struct {
	ID       string `json:"id"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
}

// Get returns the order of req.ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, req GetReq) (GetResp, error) {
	o, ok, err := s.orders.Get(ctx, req.ID)
	if err != nil {
		return GetResp{}, err
	}
	if !ok {
		return GetResp{}, ErrNotFound
	}
	return GetResp{ID: o.ID, SKU: o.SKU, Quantity: o.Quantity, Status: o.Status}, nil
}
