package ynab

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-offramp/v3/math"
)

// CreateTransactions creates all of the necessary transactions to record the transfers between accounts.
func CreateTransactions(
	fundsOriginAccountID string,
	recipientAccountID string,
	outboundBalances map[string]*math.OutboundTransactionBalance,
	accountNamesByID map[string]string,
	payeeIDsByAccountID map[string]string, // mapping account ID to the payee ID to use to write a transfer to that account
	startDate time.Time,
	endDate time.Time,
) ([]ynab.SaveTransaction, error) {
	nowDate := time.Now().Format(time.DateOnly)

	sumCents := 0
	for _, outboundBalance := range outboundBalances {
		sumCents += outboundBalance.ToCents()
	}

	// Not sure when this would happen, but, if it does, then just don't do anything
	if sumCents == 0 {
		return nil, nil
	}

	// First build the transactions to move funds from the originating account
	// to the recipient account
	recipientAccountPayeeID, hasID := payeeIDsByAccountID[recipientAccountID]
	if !hasID {
		return nil, fmt.Errorf("unable to resolve transfer payee ID to receive funds at recipient account ID '%s'", recipientAccountID)
	}

	transactions := []ynab.SaveTransaction{
		{
			AccountId: fundsOriginAccountID,
			PayeeId:   recipientAccountPayeeID,
			Amount:    sumCents * -10,
			Date:      nowDate,
			Memo:      buildSummaryMemo(startDate, endDate, outboundBalances, accountNamesByID),
		},
	}

	// Create a transaction from the recipient account to each off the offramp account
	for offrampAccountID, outboundBalance := range outboundBalances {
		// If this balance is to be left in the recipient account, create no new transactions
		if offrampAccountID == recipientAccountID {
			continue
		}

		// If nothing's moving, don't create a transaction for it.
		if outboundBalance.Cents+outboundBalance.Dollars == 0 {
			continue
		}

		transactions = append(transactions, ynab.SaveTransaction{
			AccountId: offrampAccountID,
			PayeeId:   recipientAccountPayeeID,
			Amount:    outboundBalance.ToCents() * 10,
			Date:      nowDate,
			Memo:      buildBasicTransferMemo(startDate, endDate),
		})
	}

	return transactions, nil
}

func buildBasicTransferMemo(startDate time.Time, endDate time.Time) string {
	return fmt.Sprintf("Bills %s - %s", startDate.Format("01/02"), endDate.Format("01/02"))
}

func buildSummaryMemo(startDate time.Time, endDate time.Time, outboundBalances map[string]*math.OutboundTransactionBalance, accountNamesByID map[string]string) string {
	accountTotals := make([]string, 0, len(outboundBalances))
	for accountID, outboundBalance := range outboundBalances {
		accountName, hasName := accountNamesByID[accountID]
		if !hasName {
			// There should be enough guards to prevent this from happening,
			// but account for it, nonetheless, just in case
			accountName = accountID
		}

		accountTotals = append(accountTotals, fmt.Sprintf("%s: $%d.%02d", accountName, outboundBalance.Dollars, outboundBalance.Cents))
	}

	// Get some kind of consistency in ordering, if just to help tests
	sort.Strings(accountTotals)

	return fmt.Sprintf("%s: %s", buildBasicTransferMemo(startDate, endDate), strings.Join(accountTotals, "; "))
}
