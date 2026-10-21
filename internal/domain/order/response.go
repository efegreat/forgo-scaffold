package order

import "example/internal/entity"

// The responses are what clients receive. They copy the fields of the entities they show, never embed
// them, so that a change to an entity does not change the API.

type PlaceResp struct {
	ID string `json:"id"`
}

type GetResp struct {
	ID       string `json:"id"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
}

func newGetResp(o entity.Order) GetResp {
	return GetResp{ID: o.ID, SKU: o.SKU, Quantity: o.Quantity, Status: string(o.Status)}
}
