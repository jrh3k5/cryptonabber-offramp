package ynab

import "fmt"

// MinimumBalanceAdjustment represents the minimum balance adjustment for an account
// to maintain a minimum balance after a set of scheduled transactions have been applied.
type MinimumBalanceAdjustment struct {
	Dollars int
	Cents   int
}

// ToCents expresses the amount in just cents.
func (m *MinimumBalanceAdjustment) ToCents() int {
	return (m.Dollars * 100) + m.Cents
}

func (m *MinimumBalanceAdjustment) String() string {
	return fmt.Sprintf("$%d.%02d", m.Dollars, m.Cents)
}
