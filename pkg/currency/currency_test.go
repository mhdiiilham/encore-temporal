package currency_test

import (
	"testing"

	"encore.app/pkg/currency"
)

func TestFormatString(t *testing.T) {
	tests := []struct {
		name           string
		currency       string
		originalAmount int64
		want           string
	}{
		{
			name:           "USD with cents",
			currency:       "USD",
			originalAmount: 12345,
			want:           "$123.45",
		},
		{
			name:           "GEL with cents",
			currency:       "GEL",
			originalAmount: 500,
			want:           "₾5.00",
		},
		{
			name:           "Unknown currency no division",
			currency:       "XYZ",
			originalAmount: 999,
			want:           "999.00",
		},
		{
			name:           "USD exact dollar",
			currency:       "USD",
			originalAmount: 100,
			want:           "$1.00",
		},
		{
			name:           "GEL zero value",
			currency:       "GEL",
			originalAmount: 0,
			want:           "₾0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := currency.FormatString(tt.currency, tt.originalAmount)
			if got != tt.want {
				t.Errorf("FormatString(%q, %d) = %q, want %q",
					tt.currency, tt.originalAmount, got, tt.want)
			}
		})
	}
}
