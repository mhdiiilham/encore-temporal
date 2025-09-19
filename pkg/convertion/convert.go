package convertion

import (
	"fmt"
	"math"
)

var exchangeRates = map[string]map[string]float64{
	"GEL": {
		"USD": 0.36,
	},
	"USD": {
		"GEL": 1 / 0.36,
	},
}

// ConvertAmount converts an amount from baseCurrency to targetCurrency.
// Amount is expected in the smallest unit of baseCurrency (e.g., cents).
// Returns the converted amount in the smallest unit of targetCurrency,
// the conversion rate, and an error if currencies are unsupported.
func ConvertAmount(amount int64, baseCurrency, targetCurrency string) (int64, float64, error) {
	if baseCurrency == targetCurrency {
		return amount, 1.0, nil
	}

	ratesFromBase, ok := exchangeRates[baseCurrency]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported base currency %s", baseCurrency)
	}

	rate, ok := ratesFromBase[targetCurrency]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported target currency %s -> %s", baseCurrency, targetCurrency)
	}

	baseAsFloat := float64(amount) / 100.0
	targetValue := baseAsFloat * rate

	targetSmallestUnit := int64(math.Round(targetValue * 100))

	return targetSmallestUnit, rate, nil
}
