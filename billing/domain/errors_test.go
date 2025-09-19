package domain_test

import (
	"testing"

	"encore.app/billing/domain"
	"github.com/stretchr/testify/assert"
)

func TestDomainErrors(t *testing.T) {
	assert.EqualError(t, domain.ErrBillNotFound, "bill not found")
	assert.EqualError(t, domain.ErrBillClosed, "bill is already closed")
	assert.EqualError(t, domain.ErrInvalidCurrency, "invalid currency")
	assert.EqualError(t, domain.ErrInvalidPrice, "invalid price")
	assert.EqualError(t, domain.ErrInvalidItemName, "invalid item name")
	assert.EqualError(t, domain.ErrWorkflowNotFound, "workflow not found")
}

func TestValidationError(t *testing.T) {
	err := domain.ValidationError{
		Field:   "amount",
		Message: "must be greater than zero",
	}

	assert.EqualError(t, err, "amount: must be greater than zero")
}
