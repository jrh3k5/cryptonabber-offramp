package ynab

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
)

// CreateTransactions creates all of the necessary transactions to record the transfers between accounts.
func CreateTransactions(
	fundsOriginAccountID string,
	recipientAccountID string,
	outboundBalancesByAccountID map[string]*OutboundTransactionBalance,
	balanceAdjustmentsByAccountID map[string]*MinimumBalanceAdjustment,
	accountNamesByID map[string]string,
	payeeIDsByAccountID map[string]string, // mapping account ID to the payee ID to use to write a transfer to that account
	startDate time.Time,
	endDate time.Time,
) ([]ynab.SaveTransaction, error) {
	nowDate := time.Now().Format(time.DateOnly)

	uniqueAccountIDs := make(map[string]any)

	sumCents := 0
	for accountID, outboundBalance := range outboundBalancesByAccountID {
		sumCents += outboundBalance.ToCents()

		uniqueAccountIDs[accountID] = nil
	}

	for accountID, balanceAdjustment := range balanceAdjustmentsByAccountID {
		sumCents += balanceAdjustment.ToCents()

		uniqueAccountIDs[accountID] = nil
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
			Memo:      buildSummaryMemo(startDate, endDate, outboundBalancesByAccountID, balanceAdjustmentsByAccountID, accountNamesByID),
		},
	}

	// Create a transaction from the recipient account to each off the offramp account
	for offrampAccountID := range uniqueAccountIDs {
		// If this balance is to be left in the recipient account, create no new transactions
		if offrampAccountID == recipientAccountID {
			continue
		}

		var totalTransfer int

		outboundBalance, hasBalance := outboundBalancesByAccountID[offrampAccountID]
		if hasBalance {
			totalTransfer += outboundBalance.ToCents()
		}

		balanceAdjustment, hasBalanceAdjustment := balanceAdjustmentsByAccountID[offrampAccountID]
		if hasBalanceAdjustment {
			totalTransfer += balanceAdjustment.ToCents()
		}

		// If nothing's moving, don't create a transaction for it.
		if totalTransfer == 0 {
			continue
		}

		transactions = append(transactions, ynab.SaveTransaction{
			AccountId: offrampAccountID,
			PayeeId:   recipientAccountPayeeID,
			Amount:    totalTransfer * 10,
			Date:      nowDate,
			Memo:      buildBasicTransferMemo(startDate, endDate, balanceAdjustment),
		})
	}

	return transactions, nil
}

func buildBasicTransferMemo(startDate, endDate time.Time, minimumBalanceAdjustment *MinimumBalanceAdjustment) string {
	memoString := fmt.Sprintf("Bills %s - %s", startDate.Format("01/02"), endDate.Format("01/02"))
	if minimumBalanceAdjustment != nil {
		memoString += fmt.Sprintf(" (minimum balance adjustment: %s)", minimumBalanceAdjustment.String())
	}

	return memoString
}

func buildSummaryMemo(
	startDate time.Time,
	endDate time.Time,
	outboundBalancesByAccountID map[string]*OutboundTransactionBalance,
	minimumBalanceAdjustmentsByAccountID map[string]*MinimumBalanceAdjustment,
	accountNamesByID map[string]string,
) string {
	perAccountTotalAmounts := make(map[string]int)
	for accountID, outboundBalance := range outboundBalancesByAccountID {
		perAccountTotalAmounts[accountID] = outboundBalance.ToCents()
	}

	for accountID, minimumBalanceAdjustment := range minimumBalanceAdjustmentsByAccountID {
		perAccountTotalAmounts[accountID] += minimumBalanceAdjustment.ToCents()
	}

	accountTotals := make([]string, 0, len(perAccountTotalAmounts))
	for accountID, totalTransferAmount := range perAccountTotalAmounts {
		accountName, hasName := accountNamesByID[accountID]
		if !hasName {
			// There should be enough guards to prevent this from happening,
			// but account for it, nonetheless, just in case
			accountName = accountID
		}

		totalTransferCentsRemainder := totalTransferAmount % 100
		totalTransferDollars := (totalTransferAmount - totalTransferCentsRemainder) / 100

		accountTotals = append(accountTotals, fmt.Sprintf("%s: $%d.%02d", accountName, totalTransferDollars, totalTransferCentsRemainder))
	}

	// Get some kind of consistency in ordering, if just to help tests
	sort.Strings(accountTotals)

	return fmt.Sprintf("%s: %s", buildBasicTransferMemo(startDate, endDate, nil), strings.Join(accountTotals, "; "))
}
