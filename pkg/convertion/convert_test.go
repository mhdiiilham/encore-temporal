package convertion_test

import (
	"testing"

	"encore.app/pkg/convertion"
	"github.com/stretchr/testify/assert"
)

func TestConvertAmount(t *testing.T) {
	tests := []struct {
		name           string
		amount         int64
		baseCurrency   string
		targetCurrency string
		expectedAmount int64
		expectedRate   float64
		expectError    bool
	}{
		{
			name:           "GEL to USD",
			amount:         10000,
			baseCurrency:   "GEL",
			targetCurrency: "USD",
			expectedAmount: 3600,
			expectedRate:   0.36,
			expectError:    false,
		},
		{
			name:           "USD to GEL",
			amount:         10000,
			baseCurrency:   "USD",
			targetCurrency: "GEL",
			expectedAmount: 27778,
			expectedRate:   1 / 0.36,
			expectError:    false,
		},
		{
			name:           "Same currency",
			amount:         12345,
			baseCurrency:   "USD",
			targetCurrency: "USD",
			expectedAmount: 12345,
			expectedRate:   1.0,
			expectError:    false,
		},
		{
			name:           "Unsupported base currency",
			amount:         100,
			baseCurrency:   "EUR",
			targetCurrency: "USD",
			expectError:    true,
		},
		{
			name:           "Unsupported target currency",
			amount:         100,
			baseCurrency:   "USD",
			targetCurrency: "EUR",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, rate, err := convertion.ConvertAmount(tt.amount, tt.baseCurrency, tt.targetCurrency)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAmount, amount)
				assert.Equal(t, tt.expectedRate, rate)
			}
		})
	}
}
