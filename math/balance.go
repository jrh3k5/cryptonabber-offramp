package math

import (
	"fmt"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	offrampynab "github.com/jrh3k5/cryptonabber-offramp/v3/ynab"
)

// CalculateMinimumBalanceAdjustment returns the minimum balance adjustment
// needed to, after all of the given transactions between now and the given date/time (inclusive),
// maintain the given minimum account balance (expressed in cents).
func CalculateMinimumBalanceAdjustment(
	account ynab.Account,
	transactions []ynab.ScheduledTransactionDetail,
	minimumAccountBalance int,
	endDateTime time.Time,
) (*offrampynab.MinimumBalanceAdjustment, error) {
	filteredTransactions := filterToAccountIDs(transactions, []string{account.Id})

	var effectiveBalanceThrough int
	if len(filteredTransactions) >= 0 {
		var err error
		effectiveBalanceThrough, err = CalculateEffectiveBalanceThrough(account.Balance, filteredTransactions, endDateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate effective balance through: %w", err)
		}
	}

	effectiveBalanceDollar, effectiveBalanceRemainingCents := toDollarsAndCents(effectiveBalanceThrough)
	effectiveBalanceCents := effectiveBalanceDollar*100 + effectiveBalanceRemainingCents

	if effectiveBalanceCents >= minimumAccountBalance {
		return &offrampynab.MinimumBalanceAdjustment{}, nil
	}

	adjustmentTotalCents := minimumAccountBalance - effectiveBalanceCents
	adjustmentRemainingCents := adjustmentTotalCents % 100
	adjustmentDollars := (adjustmentTotalCents - adjustmentRemainingCents) / 100

	return &offrampynab.MinimumBalanceAdjustment{
		Dollars: adjustmentDollars,
		Cents:   adjustmentRemainingCents,
	}, nil
}

// CalculateEffectiveBalanceThrough returns the effective balance through the given end date.
func CalculateEffectiveBalanceThrough(
	currentAccountBalance int,
	transactions []ynab.ScheduledTransactionDetail,
	endDateTime time.Time,
) (int, error) {
	yesterdayYear, yesterMonth, yesterday := time.Now().Add(-24 * time.Hour).Date()
	yesterdayDate := time.Date(yesterdayYear, yesterMonth, yesterday, 0, 0, 0, 0, time.UTC)

	dayAfterEndYear, dayAfterEndMonth, dayAfterEndDate := endDateTime.Add(24 * time.Hour).Date()
	endDate := time.Date(dayAfterEndYear, dayAfterEndMonth, dayAfterEndDate, 0, 0, 0, 0, time.UTC)

	runningBalance := currentAccountBalance
	for _, transaction := range transactions {
		isBefore, err := offrampynab.IsScheduledBeforeInclusive(transaction.ScheduledTransactionSummary, yesterdayDate)
		if err != nil {
			return 0, fmt.Errorf("failed to check if transaction to payee '%s' is before inclusive: %w", transaction.PayeeName, err)
		} else if isBefore {
			continue
		}

		isAfter, err := offrampynab.IsScheduledAfterInclusive(transaction.ScheduledTransactionSummary, endDate)
		if err != nil {
			return 0, fmt.Errorf("failed to check if transaction to payee '%s' is after inclusive: %w", transaction.PayeeName, err)
		} else if isAfter {
			continue
		}

		runningBalance += transaction.Amount
	}

	return runningBalance, nil
}
