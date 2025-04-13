package math

import (
	"fmt"
	"math"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	offrampynab "github.com/jrh3k5/cryptonabber-offramp/v3/ynab"
)

// CalculateOutboundTransactions will pull, from the given scheduled transactions, all outbound transactions that are happening
// within the given start and end date/time (inclusive) for the given account IDs.
func CalculateOutboundTransactions(
	accountIDs []string,
	excludedColorsByAccountID map[string][]string,
	transactions []ynab.ScheduledTransactionDetail,
	startDate time.Time,
	endDate time.Time,
) (map[string]*offrampynab.OutboundTransactionBalance, error) {
	filteredByAccount := filterToAccountIDs(transactions, accountIDs)

	filteredByDate, err := filterTransactionsByDateRange(filteredByAccount, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to filter transactions by date range: %w", err)
	}

	outboundOnly := filterToOutboundOnly(filteredByDate)

	onlyAllowedFlags := filterToOnlyAllowedFlags(outboundOnly, excludedColorsByAccountID)

	grouped := groupTransactionsByAccountID(accountIDs, onlyAllowedFlags)

	balances := make(map[string]*offrampynab.OutboundTransactionBalance)
	for accountID, accountTransactions := range grouped {
		sum := sumTransactions(accountTransactions)
		sum = int(math.Abs(float64(sum)))
		dollars, cents := toDollarsAndCents(sum)
		balances[accountID] = &offrampynab.OutboundTransactionBalance{
			Dollars: dollars,
			Cents:   cents,
		}
	}

	return balances, nil
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
		isAfterInclusive, err := offrampynab.IsScheduledAfterInclusive(transaction.ScheduledTransactionSummary, startDate)
		if err != nil {
			return nil, fmt.Errorf("failed to check if transaction to payee '%s' is after inclusive: %w", transaction.PayeeName, err)
		}

		if !isAfterInclusive {
			continue
		}

		isBeforeInclusive, err := offrampynab.IsScheduledBeforeInclusive(transaction.ScheduledTransactionSummary, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to check if transaction to payee '%s' is before inclusive: %w", transaction.PayeeName, err)
		}

		if !isBeforeInclusive {
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
