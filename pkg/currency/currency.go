package currency

import "fmt"

// FormatString formats a currency amount into a human-readable string.
//
// The function expects the amount in the smallest currency unit (e.g. cents for USD).
// It converts the amount to a float with 2 decimal places and prepends the
// appropriate currency symbol.
//
// Supported currencies:
//   - USD: represented with `$`, amount divided by 100
//   - GEL: represented with `₾`, amount divided by 100
//
// If the currency is not recognized, the amount is returned without a symbol
// and is not divided by 100.
//
// Example:
//
//	FormatString("USD", 12345) // "$123.45"
//	FormatString("GEL", 500)   // "₾5.00"
//	FormatString("XYZ", 999)   // "999.00"
func FormatString(currency string, originalAmount int64) string {
	var symbol string
	var amount float64

	switch currency {
	case "USD":
		symbol = "$"
		amount = float64(originalAmount) / 100
	case "GEL":
		symbol = "₾"
		amount = float64(originalAmount) / 100
	default:
		symbol = ""
		amount = float64(originalAmount)
	}

	return fmt.Sprintf("%s%.2f", symbol, amount)
}
