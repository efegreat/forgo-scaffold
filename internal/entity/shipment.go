package entity

// ShipmentStatus is the status of a Shipment.
type ShipmentStatus string

const (
	// ShipmentStatusPending is the status of a shipment that waits for its dispatch.
	ShipmentStatusPending ShipmentStatus = "pending"
	// ShipmentStatusDispatched is the status of a shipment that was dispatched.
	ShipmentStatusDispatched ShipmentStatus = "dispatched"
)

// Shipment is the shipment of an order.
type Shipment struct {
	OrderID  string
	SKU      string
	Quantity int
	Status   ShipmentStatus
}
