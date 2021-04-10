package util

const (
	RMB = "RMB"
	USD = "USD"
	EUR = "EUR"
)

// IsSupporttedCurrency validates the currency is supportted
func IsSupporttedCurrency(currency string) bool {
	switch currency {
	case RMB, USD, EUR:
		return true
	}
	return false
}
