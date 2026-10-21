package errs

import "github.com/efegreat/forgo/operation"

// ErrOrderNotFound is the error of an order that does not exist.
var ErrOrderNotFound = operation.Error{
	Code:     "order_not_found",
	Status:   operation.StatusNotFound,
	Message:  "order not found",
	Severity: operation.SeverityLight,
}
