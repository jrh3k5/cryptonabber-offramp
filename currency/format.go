package currency

import (
	"fmt"
	"math"
)

// FormatCents formats the given expression of USD in cents to a dollar-and-cents string.
func FormatCents(cents int) string {
	remainderCents := cents % 100
	dollars := (cents - remainderCents) / 100

	return FormatDollarsAndCents(dollars, remainderCents)
}

// FormatDollarsAndCents formats the given USD dollars and cents to a dollar-and-cents string.
func FormatDollarsAndCents(dollars, cents int) string {
	absoluteDollars := int(math.Abs(float64(dollars)))
	absoluteCents := int(math.Abs(float64(cents)))

	formattedDollars := fmt.Sprintf("$%d.%02d", absoluteDollars, absoluteCents)

	if dollars < 0 || cents < 0 {
		return "-" + formattedDollars
	}

	return formattedDollars
}
