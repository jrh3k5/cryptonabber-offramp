package ynab

import "fmt"

// OutboundTransactionBalance represents the balance of outbound transactions
// for a particular account.
type OutboundTransactionBalance struct {
	Dollars int
	Cents   int
}

// ToCents expresses the amount in just cents.
func (o *OutboundTransactionBalance) ToCents() int {
	return (o.Dollars * 100) + o.Cents
}

func (o *OutboundTransactionBalance) String() string {
	return fmt.Sprintf("$%d.%02d", o.Dollars, o.Cents)
}
