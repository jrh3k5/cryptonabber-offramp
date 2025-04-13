package math

// toDollarsAndCents returns the dollars and cents of the given amount,
// converting YNAB's three-significant-place precision to two-significant-place.
func toDollarsAndCents(amount int) (int, int) {
	if amount == 0 {
		return 0, 0
	}

	// YNAB expresses cents to the third decimal place
	cents := amount % 1000

	dollars := (amount - cents) / 1000
	cents = (cents - (cents % 10)) / 10

	return dollars, cents
}
