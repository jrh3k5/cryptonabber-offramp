package ynab_test

import (
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cliynab "github.com/jrh3k5/cryptonabber-offramp/v3/ynab"
)

var _ = Describe("Transactions", func() {
	Context("CreateTransactions", func() {
		var fundsOriginAccountID string
		var fundsRecipientAccountID string
		var offrampAccountID0 string
		var offrampAccountID1 string

		var outboundBalances map[string]*cliynab.OutboundTransactionBalance
		var balanceAdjustmentsByAccountID map[string]*cliynab.MinimumBalanceAdjustment
		var namesByID map[string]string
		var payeesByAccountID map[string]string
		var startDate time.Time
		var endDate time.Time

		BeforeEach(func() {
			fundsOriginAccountID = "funds-origin"
			fundsRecipientAccountID = "funds-recipient"

			offrampAccountID0 = "offramp0"
			offrampAccountID1 = "offramp1"

			namesByID = map[string]string{
				fundsOriginAccountID:    "Funds Origin",
				fundsRecipientAccountID: "Funds Recipient",
				offrampAccountID0:       "Offramp 0",
				offrampAccountID1:       "Offramp 1",
			}

			payeesByAccountID = map[string]string{
				fundsOriginAccountID:    "payee-origin",
				fundsRecipientAccountID: "payee-recipient",
				offrampAccountID0:       "payee-offramp0",
				offrampAccountID1:       "payee-offramp1",
			}

			outboundBalances = map[string]*cliynab.OutboundTransactionBalance{
				offrampAccountID0: {
					Dollars: 1,
					Cents:   23,
				},
				offrampAccountID1: {
					Dollars: 420,
					Cents:   69,
				},
			}

			balanceAdjustmentsByAccountID = map[string]*cliynab.MinimumBalanceAdjustment{}

			startDate, _ = time.Parse(time.DateOnly, "2024-02-01")
			endDate, _ = time.Parse(time.DateOnly, "2024-02-03")
		})

		It("creates the transactions to transfer funds", func() {
			recipientAccountPayeeID := payeesByAccountID[fundsRecipientAccountID]
			// sanity check
			Expect(recipientAccountPayeeID).ToNot(BeEmpty(), "there should be a payee ID for the recipient account set up")

			transactions, err := cliynab.CreateTransactions(fundsOriginAccountID,
				fundsRecipientAccountID,
				outboundBalances,
				balanceAdjustmentsByAccountID,
				namesByID,
				payeesByAccountID,
				startDate,
				endDate)

			Expect(err).ToNot(HaveOccurred(), "creating the transactions should not fail")
			Expect(transactions).To(HaveLen(3), "the correct number of transactions should be created")

			fundsOriginTransaction := getTransactionByAccountID(fundsOriginAccountID, transactions)
			Expect(fundsOriginTransaction.Amount).To(Equal(-421920), "the funds origin account should be debited the total amount")
			Expect(fundsOriginTransaction.PayeeId).To(Equal(recipientAccountPayeeID), "the funds origin transaction should be a transfer to the funds recipient account")

			offramp0Transaction := getTransactionByAccountID(offrampAccountID0, transactions)
			Expect(offramp0Transaction.Amount).To(Equal(1230), "offramp account 0 should be receiving its outbound amount")
			Expect(offramp0Transaction.PayeeId).To(Equal(recipientAccountPayeeID), "the funds should be coming from the recipient account")

			offramp1Transaction := getTransactionByAccountID(offrampAccountID1, transactions)
			Expect(offramp1Transaction.Amount).To(Equal(420690), "offramp account 1 should be receiving its outbound amount")
			Expect(offramp1Transaction.PayeeId).To(Equal(recipientAccountPayeeID), "the funds should be coming from the recipient account")
		})

		When("there is a minimum balance adjustment for one of the accounts", func() {
			var minimumBalanceAdjustment *cliynab.MinimumBalanceAdjustment

			BeforeEach(func() {
				minimumBalanceAdjustment = &cliynab.MinimumBalanceAdjustment{
					Dollars: 100,
					Cents:   25,
				}

				balanceAdjustmentsByAccountID[offrampAccountID0] = minimumBalanceAdjustment
			})

			It("incorporates the minimum balance adjustment", func() {
				recipientAccountPayeeID := payeesByAccountID[fundsRecipientAccountID]
				// sanity check
				Expect(recipientAccountPayeeID).ToNot(BeEmpty(), "there should be a payee ID for the recipient account set up")

				transactions, err := cliynab.CreateTransactions(fundsOriginAccountID,
					fundsRecipientAccountID,
					outboundBalances,
					balanceAdjustmentsByAccountID,
					namesByID,
					payeesByAccountID,
					startDate,
					endDate)

				Expect(err).ToNot(HaveOccurred(), "creating the transactions should not fail")
				Expect(transactions).To(HaveLen(3), "the correct number of transactions should be created")

				fundsOriginTransaction := getTransactionByAccountID(fundsOriginAccountID, transactions)
				Expect(fundsOriginTransaction.Amount).To(Equal(-522170), "the funds origin account should be debited the total amount (including minimum balance adjustment)")
				Expect(fundsOriginTransaction.PayeeId).To(Equal(recipientAccountPayeeID), "the funds origin transaction should be a transfer to the funds recipient account")

				offramp0Transaction := getTransactionByAccountID(offrampAccountID0, transactions)
				Expect(offramp0Transaction.Amount).To(Equal(101480), "offramp account 0 should be receiving its outbound amount (including minimum balance adjustment)")
				Expect(offramp0Transaction.PayeeId).To(Equal(recipientAccountPayeeID), "the funds should be coming from the recipient account")
				Expect(offramp0Transaction.Memo).To(ContainSubstring("minimum balance adjustment: $100.25"), "the minimum balance should be mentioned in the memo")

				offramp1Transaction := getTransactionByAccountID(offrampAccountID1, transactions)
				Expect(offramp1Transaction.Amount).To(Equal(420690), "offramp account 1 should be receiving its outbound amount")
				Expect(offramp1Transaction.PayeeId).To(Equal(recipientAccountPayeeID), "the funds should be coming from the recipient account")
			})
		})

		When("the recipient account is among the offramp accounts", func() {
			BeforeEach(func() {
				outboundBalances[fundsRecipientAccountID] = &cliynab.OutboundTransactionBalance{
					Dollars: 456,
					Cents:   12,
				}
			})

			It("adds it to the funds transfer, but does not generate a transfer between the recipient account and itself", func() {
				transactions, err := cliynab.CreateTransactions(fundsOriginAccountID,
					fundsRecipientAccountID,
					outboundBalances,
					balanceAdjustmentsByAccountID,
					namesByID,
					payeesByAccountID,
					startDate,
					endDate)

				Expect(err).ToNot(HaveOccurred(), "creating the transactions should not fail")
				Expect(transactions).To(HaveLen(3), "the correct number of transactions should be created")

				fundsOriginTransaction := getTransactionByAccountID(fundsOriginAccountID, transactions)
				Expect(fundsOriginTransaction.Amount).To(Equal(-878040), "the transfer out of the funds origination account should include the amount staying in the recipient account")

				Expect(getTransactionsByAccountID(fundsRecipientAccountID, transactions)).To(BeEmpty(), "there should be no transactions for funds staying in the recipient account")
			})
		})

		When("one of the offramp accounts has zero currency moving", func() {
			It("does not create a transaction for that account", func() {
				outboundBalances[offrampAccountID1].Cents = 0
				outboundBalances[offrampAccountID1].Dollars = 0

				transactions, err := cliynab.CreateTransactions(fundsOriginAccountID,
					fundsRecipientAccountID,
					outboundBalances,
					balanceAdjustmentsByAccountID,
					namesByID,
					payeesByAccountID,
					startDate,
					endDate)

				Expect(err).ToNot(HaveOccurred(), "creating the transactions should not fail")
				Expect(transactions).To(HaveLen(2), "only accounts with a non-zero sum of transactions should be created")

				// This makes sure that only transactions involving the fund source account and offramp account 0.
				// These calls fail the test if the requested account is not present.
				// As the above ensures there are only two transactions, this confirms that those two are
				// for the expected accounts.
				getTransactionByAccountID(offrampAccountID0, transactions)
				getTransactionByAccountID(fundsOriginAccountID, transactions)
			})

			It("does not include the account in the summary memo", func() {
				outboundBalances[offrampAccountID1].Cents = 0
				outboundBalances[offrampAccountID1].Dollars = 0

				transactions, err := cliynab.CreateTransactions(fundsOriginAccountID,
					fundsRecipientAccountID,
					outboundBalances,
					balanceAdjustmentsByAccountID,
					namesByID,
					payeesByAccountID,
					startDate,
					endDate)

				Expect(err).ToNot(HaveOccurred(), "creating the transactions should not fail")

				fundsOriginTransaction := getTransactionByAccountID(fundsOriginAccountID, transactions)
				Expect(fundsOriginTransaction.Memo).ToNot(ContainSubstring("Offramp 1"), "accounts with zero balance transfer should not appear in the memo")
				Expect(fundsOriginTransaction.Memo).To(ContainSubstring("Offramp 0"), "accounts with non-zero balance transfer should appear in the memo")
			})
		})
	})
})

func getTransactionByAccountID(accountID string, transactions []ynab.SaveTransaction) ynab.SaveTransaction {
	matches := getTransactionsByAccountID(accountID, transactions)

	Expect(matches).To(HaveLen(1), "there should only be one transaction for account ID '%s'", accountID)

	return matches[0]
}

func getTransactionsByAccountID(accountID string, transactions []ynab.SaveTransaction) []ynab.SaveTransaction {
	var matches []ynab.SaveTransaction
	for _, transaction := range transactions {
		if transaction.AccountId == accountID {
			matches = append(matches, transaction)
		}
	}

	return matches
}
