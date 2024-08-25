package math

import (
	"fmt"
	"math"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
)

type OutboundTransactionBalance struct {
	Dollars int
	Cents   int
}

func (o *OutboundTransactionBalance) ToCents() int {
	return (o.Dollars * 100) + o.Cents
}

// CalculateOutboundTransactions will pull, from the given scheduled transactions, all outbound transactions that are happening
// within the given start and end date/time (inclusive) for the given account IDs.
func CalculateOutboundTransactions(
	accountIDs []string,
	excludedColorsByAccountID map[string][]string,
	transactions []ynab.ScheduledTransactionDetail,
	startDate time.Time,
	endDate time.Time,
) (map[string]*OutboundTransactionBalance, error) {
	filteredByAccount := filterToAccountIDs(transactions, accountIDs)

	filteredByDate, err := filterTransactionsByDateRange(filteredByAccount, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to filter transactions by date range: %w", err)
	}

	outboundOnly := filterToOutboundOnly(filteredByDate)

	onlyAllowedFlags := filterToOnlyAllowedFlags(outboundOnly, excludedColorsByAccountID)

	grouped := groupTransactionsByAccountID(accountIDs, onlyAllowedFlags)

	balances := make(map[string]*OutboundTransactionBalance)
	for accountID, accountTransactions := range grouped {
		sum := sumTransactions(accountTransactions)
		sum = int(math.Abs(float64(sum)))
		dollars, cents := toDollarsAndCents(sum)
		balances[accountID] = &OutboundTransactionBalance{
			Dollars: dollars,
			Cents:   cents,
		}
	}

	return balances, nil
}

func filterToAccountIDs(transactions []ynab.ScheduledTransactionDetail, accountIDs []string) []ynab.ScheduledTransactionDetail {
	var included []ynab.ScheduledTransactionDetail

	for _, transaction := range transactions {
		include := false
		for _, accountID := range accountIDs {
			if accountID == transaction.AccountId {
				include = true
				break
			}
		}

		if !include {
			continue
		}

		included = append(included, transaction)
	}

	return included
}
func filterToOnlyAllowedFlags(transactions []ynab.ScheduledTransactionDetail, excludedColorsByAccountID map[string][]string) []ynab.ScheduledTransactionDetail {
	if excludedColorsByAccountID == nil {
		return transactions
	}

	trimmedDown := make([]ynab.ScheduledTransactionDetail, len(transactions))
	copy(trimmedDown, transactions)

	for i := len(trimmedDown) - 1; i >= 0; i-- {
		transaction := trimmedDown[i]
		if transaction.FlagColor == nil {
			continue
		}

		excludedColors, hasExclusions := excludedColorsByAccountID[transaction.AccountId]
		if !hasExclusions {
			continue
		}

		for _, excludedColor := range excludedColors {
			if excludedColor == *transaction.FlagColor {
				fmt.Printf("Removing transaction: %v\n", transaction)
				trimmedDown = append(trimmedDown[:i], trimmedDown[i+1:]...)

				break
			}
		}
	}

	return trimmedDown
}

func filterToOutboundOnly(transactions []ynab.ScheduledTransactionDetail) []ynab.ScheduledTransactionDetail {
	var included []ynab.ScheduledTransactionDetail

	for _, transaction := range transactions {
		if transaction.Amount >= 0 {
			continue
		}

		included = append(included, transaction)
	}

	return included
}

func filterTransactionsByDateRange(transactions []ynab.ScheduledTransactionDetail, startDate time.Time, endDate time.Time) ([]ynab.ScheduledTransactionDetail, error) {
	var included []ynab.ScheduledTransactionDetail

	for _, transaction := range transactions {
		nextDate, parseErr := time.Parse(time.DateOnly, transaction.DateNext)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse 'date next' value of '%s' for transaction to payee '%s': %w", transaction.DateNext, transaction.PayeeName, parseErr)
		}

		if nextDate.Before(startDate) || nextDate.After(endDate) {
			continue
		}

		included = append(included, transaction)
	}

	return included, nil
}

func groupTransactionsByAccountID(accountIDs []string, transactions []ynab.ScheduledTransactionDetail) map[string][]ynab.ScheduledTransactionDetail {
	grouped := make(map[string][]ynab.ScheduledTransactionDetail)
	for _, accountID := range accountIDs {
		grouped[accountID] = make([]ynab.ScheduledTransactionDetail, 0)
	}

	for _, transaction := range transactions {
		grouped[transaction.AccountId] = append(grouped[transaction.AccountId], transaction)
	}

	return grouped
}

func sumTransactions(transactions []ynab.ScheduledTransactionDetail) int {
	summed := 0
	for _, transaction := range transactions {
		summed += transaction.Amount
	}
	return summed
}

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
