package order

// Limits are the limits of orders: a live config, which admins change while the app runs through the ops
// server's live config endpoints, and which Place reads on every order. It has no MemoryTTL, so a change
// applies at once on every instance. A config that can take a delay of a few seconds would set one, and
// save each Place a read of the store.
type Limits struct {
	// MaxQuantity is the largest quantity of one order.
	MaxQuantity int `json:"max_quantity" validate:"min=1"`
}

// Name is the key of the config in the store.
func (Limits) Name() string { return "order_limits" }

// Default holds the limits until an admin saves others, and the fields a saved document lacks.
func (Limits) Default() Limits { return Limits{MaxQuantity: 100} }
