package math

import "github.com/davidsteinsland/ynab-go/ynab"

// filterToAccountIDs will filter the given transactions to only include those for the given account IDs
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
