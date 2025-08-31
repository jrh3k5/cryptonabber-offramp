package math

import (
	"fmt"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-offramp/v3/currency"
	offrampynab "github.com/jrh3k5/cryptonabber-offramp/v3/ynab"
)

// CalculateMinimumBalanceAdjustment returns the minimum balance adjustment
// needed to, after all of the given transactions between now and the given date/time (inclusive),
// maintain the given minimum account balance (expressed in cents).
func CalculateMinimumBalanceAdjustment(
	account ynab.Account,
	transactions []ynab.ScheduledTransactionDetail,
	minimumAccountBalanceCents int,
	endDateTime time.Time,
	debug bool,
) (*offrampynab.MinimumBalanceAdjustment, error) {
	filteredTransactions := filterToAccountIDs(transactions, []string{account.Id})

	// Print project bill expenses and time range (only in debug mode)
	if debug {
		yesterdayYear, yesterMonth, yesterday := time.Now().Add(-24 * time.Hour).Date()
		yesterdayDate := time.Date(yesterdayYear, yesterMonth, yesterday, 0, 0, 0, 0, time.UTC)

		fmt.Printf("Balance adjustment calculation for account '%s':\n", account.Name)
		fmt.Printf("Time range: %s to %s\n", yesterdayDate.Format(time.DateOnly), endDateTime.Format(time.DateOnly))
		fmt.Printf("Project bill expenses:\n")

		totalExpenses := 0
		for _, transaction := range filteredTransactions {
			isBefore, err := offrampynab.IsScheduledBeforeInclusive(transaction.ScheduledTransactionSummary, yesterdayDate)
			if err != nil {
				continue
			}
			if isBefore {
				continue
			}

			dayAfterEndYear, dayAfterEndMonth, dayAfterEndDate := endDateTime.Add(24 * time.Hour).Date()
			endDate := time.Date(dayAfterEndYear, dayAfterEndMonth, dayAfterEndDate, 0, 0, 0, 0, time.UTC)

			isAfter, err := offrampynab.IsScheduledAfterInclusive(transaction.ScheduledTransactionSummary, endDate)
			if err != nil {
				continue
			}
			if isAfter {
				continue
			}

			amountDollars, amountCents := toDollarsAndCents(transaction.Amount)

			fmt.Printf("  - %s: %s\n", transaction.PayeeName, currency.FormatDollarsAndCents(amountDollars, amountCents))
			totalExpenses += transaction.Amount
		}

		totalDollars, totalCents := toDollarsAndCents(totalExpenses)

		fmt.Printf("Total expenses: %s\n", currency.FormatDollarsAndCents(totalDollars, totalCents))
	}

	var effectiveBalanceThrough int
	if len(filteredTransactions) >= 0 {
		var err error
		effectiveBalanceThrough, err = CalculateEffectiveBalanceThrough(account.Balance, filteredTransactions, endDateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate effective balance through: %w", err)
		}
	}

	effectiveBalanceDollar, effectiveBalanceRemainingCents := toDollarsAndCents(effectiveBalanceThrough)

	if debug {
		startingDollars, startingCents := toDollarsAndCents(account.Balance)
		fmt.Printf("Starting balance of account: %s\n", currency.FormatDollarsAndCents(startingDollars, startingCents))
		fmt.Printf("Projected ending balance of account: %s\n", currency.FormatDollarsAndCents(effectiveBalanceDollar, effectiveBalanceRemainingCents))
	}

	effectiveBalanceCents := effectiveBalanceDollar*100 + effectiveBalanceRemainingCents

	if effectiveBalanceCents >= minimumAccountBalanceCents {
		if debug {
			fmt.Printf("Projected account balance meets or exceeds minimum account requirement (%s), so no balance adjustment will be created", currency.FormatCents(minimumAccountBalanceCents))
		}

		return &offrampynab.MinimumBalanceAdjustment{}, nil
	}

	adjustmentTotalCents := minimumAccountBalanceCents - effectiveBalanceCents
	adjustmentRemainingCents := adjustmentTotalCents % 100
	adjustmentDollars := (adjustmentTotalCents - adjustmentRemainingCents) / 100

	if debug {
		// Print a blank line for ease of reading
		fmt.Println()
	}

	return &offrampynab.MinimumBalanceAdjustment{
		Dollars: adjustmentDollars,
		Cents:   adjustmentRemainingCents,
	}, nil
}

// CalculateEffectiveBalanceThrough returns the effective balance through the given end date.
// This is expressed as a YNAB transaction amount, not a number of cents.
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
