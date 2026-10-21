package order

// The requests are what the servers bind and validate before a method of Service runs. Their tags name
// where each field comes from on any transport, such as the JSON body or the path, without importing one.

type PlaceReq struct {
	SKU      string `json:"sku" validate:"required"`
	Quantity int    `json:"quantity" validate:"min=1"`
}

type GetReq struct {
	ID string `uri:"id" validate:"required"`
}
