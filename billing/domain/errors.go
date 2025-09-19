package domain

import "errors"

// Domain errors
var (
	ErrBillNotFound        = errors.New("bill not found")
	ErrBillClosed          = errors.New("bill is already closed")
	ErrInvalidCurrency     = errors.New("invalid currency")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrInvalidItemName     = errors.New("invalid item name")
	ErrWorkflowNotFound    = errors.New("workflow not found")
	ErrFailedToConvertBill = errors.New("failed to convert bill currency")
)

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
