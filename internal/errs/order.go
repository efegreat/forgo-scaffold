package errs

import "github.com/efegreat/forgo/operation"

var (
	// ErrOrderNotFound is the error of an order that does not exist.
	ErrOrderNotFound = operation.Error{
		Code:     "order_not_found",
		Status:   operation.StatusNotFound,
		Message:  "order not found",
		Severity: operation.SeverityLight,
	}
	// ErrQuantityAboveLimit is the error of an order of more than the live config order.Limits allows.
	ErrQuantityAboveLimit = operation.Error{
		Code:     "order_quantity_above_limit",
		Status:   operation.StatusInvalidArgument,
		Message:  "quantity above the limit of an order",
		Severity: operation.SeverityLight,
	}
	// ErrOrderNotShippable is the error of shipping an order that is not placed, such as one already shipped.
	ErrOrderNotShippable = operation.Error{
		Code:     "order_not_shippable",
		Status:   operation.StatusConflict,
		Message:  "order cannot be shipped",
		Severity: operation.SeverityModerate,
	}
)
