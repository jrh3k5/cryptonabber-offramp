package math_test

import (
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-offramp/v3/math"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutboundTransactions", func() {
	Context("CalculateOutboundTransactions", func() {
		It("groups outbound transactions together by account", func() {
			accountID0 := "account0"
			accountID1 := "account1"

			startDate, _ := time.Parse(time.DateOnly, "2020-01-01")
			endDate, _ := time.Parse(time.DateOnly, "2020-01-03")

			transactions := []ynab.ScheduledTransactionDetail{
				{
					ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
						AccountId: accountID0,
						Amount:    -1230,
						DateNext:  startDate.Format(time.DateOnly),
					},
				},
				{
					ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
						AccountId: accountID1,
						Amount:    -456780,
						DateNext:  endDate.Format(time.DateOnly),
					},
				},
				{
					ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
						AccountId: accountID0,
						Amount:    -560,
						DateNext:  startDate.Add(24 * time.Hour).Format(time.DateOnly),
					},
				},
			}

			grouped, err := math.CalculateOutboundTransactions([]string{accountID0, accountID1}, nil, transactions, startDate, endDate)
			Expect(err).ToNot(HaveOccurred(), "the calculation should not fail")
			Expect(grouped).To(And(
				HaveLen(2),
				HaveKey(accountID0),
				HaveKey(accountID1),
			), "all of the given accounts should be returned")

			account0Balance := grouped[accountID0]
			Expect(account0Balance.ToCents()).To(Equal(179), "account0 should have $1.79 outbound")

			account1Balance := grouped[accountID1]
			Expect(account1Balance.ToCents()).To(Equal(45678), "account1 should have $456.78 outbound")
		})

		When("the transactions include non-outbound transactions", func() {
			It("filters out those transactions", func() {
				accountID := "not-all-outbound"
				dateRange, _ := time.Parse(time.DateOnly, "2020-01-01")
				transactions := []ynab.ScheduledTransactionDetail{
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    12300,
							DateNext:  dateRange.Format(time.DateOnly),
						},
					},
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    -4560,
							DateNext:  dateRange.Format(time.DateOnly),
						},
					},
				}

				grouped, err := math.CalculateOutboundTransactions([]string{accountID}, nil, transactions, dateRange, dateRange)
				Expect(err).ToNot(HaveOccurred(), "calcuating the transactions should not fail")
				Expect(grouped).To(HaveKey(accountID), "the account should be returned")
				Expect(grouped[accountID].ToCents()).To(Equal(456), "the balance should not include the inbound transaction")
			})
		})

		When("the transactions include transactions for accounts not in the given list", func() {
			It("filters out those transactions", func() {
				accountID := "actually-desired-account"
				dateRange, _ := time.Parse(time.DateOnly, "2021-02-01")
				transactions := []ynab.ScheduledTransactionDetail{
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    -1230,
							DateNext:  dateRange.Format(time.DateOnly),
						},
					},
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: "excluded",
							Amount:    -4560,
							DateNext:  dateRange.Format(time.DateOnly),
						},
					},
				}

				grouped, err := math.CalculateOutboundTransactions([]string{accountID}, nil, transactions, dateRange, dateRange)
				Expect(err).ToNot(HaveOccurred(), "calulating the transactions should not fail")
				Expect(grouped).To(And(HaveLen(1), HaveKey(accountID)), "only the desired account should be in the returned amounts")
			})
		})

		When("there are transactions before the start date", func() {
			It("filters out those transactions", func() {
				accountID := "some-before-start"
				dateRange, _ := time.Parse(time.DateOnly, "2020-01-01")
				transactions := []ynab.ScheduledTransactionDetail{
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    -1230,
							DateNext:  dateRange.Format(time.DateOnly),
						},
					},
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    -4560,
							DateNext:  dateRange.Add(-24 * time.Hour).Format(time.DateOnly),
						},
					},
				}

				grouped, err := math.CalculateOutboundTransactions([]string{accountID}, nil, transactions, dateRange, dateRange)
				Expect(err).ToNot(HaveOccurred(), "calculating the outbound transactions should not fail")
				Expect(grouped).To(HaveKey(accountID), "the account should be in the returned transactions")
				Expect(grouped[accountID].ToCents()).To(Equal(123), "only the amount that fits in the date range should be accepted")
			})
		})

		When("there are transactions after the end date", func() {
			It("filters out those transactions", func() {
				accountID := "some-before-start"
				dateRange, _ := time.Parse(time.DateOnly, "2020-01-01")
				transactions := []ynab.ScheduledTransactionDetail{
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    -1230,
							DateNext:  dateRange.Format(time.DateOnly),
						},
					},
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID,
							Amount:    -4560,
							DateNext:  dateRange.Add(24 * time.Hour).Format(time.DateOnly),
						},
					},
				}

				grouped, err := math.CalculateOutboundTransactions([]string{accountID}, nil, transactions, dateRange, dateRange)
				Expect(err).ToNot(HaveOccurred(), "calculating the outbound transactions should not fail")
				Expect(grouped).To(HaveKey(accountID), "the account should be in the returned transactions")
				Expect(grouped[accountID].ToCents()).To(Equal(123), "only the amount that fits in the date range should be accepted")
			})
		})

		When("the transaction has an excluded flag color", func() {
			It("filters out those transactions", func() {
				accountID0 := "account0"
				accountID1 := "account1"

				excludedFlagColor := "red"

				startDate, _ := time.Parse(time.DateOnly, "2020-01-01")
				endDate, _ := time.Parse(time.DateOnly, "2020-01-03")
				transactions := []ynab.ScheduledTransactionDetail{
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID0,
							Amount:    -1230,
							DateNext:  startDate.Format(time.DateOnly),
						},
					},
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID1,
							Amount:    -456780,
							DateNext:  endDate.Format(time.DateOnly),
						},
					},
					{
						ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
							AccountId: accountID0,
							Amount:    -560,
							DateNext:  startDate.Add(24 * time.Hour).Format(time.DateOnly),
							FlagColor: &excludedFlagColor, // this excludes this tranasction from consideration
						},
					},
				}

				grouped, err := math.CalculateOutboundTransactions([]string{accountID0, accountID1}, map[string][]string{
					accountID0: {excludedFlagColor},
				}, transactions, startDate, endDate)
				Expect(err).ToNot(HaveOccurred(), "the calculation should not fail")
				Expect(grouped).To(And(
					HaveLen(2),
					HaveKey(accountID0),
					HaveKey(accountID1),
				), "all of the given accounts should be returned")

				account0Balance := grouped[accountID0]
				Expect(account0Balance.ToCents()).To(Equal(123), "account0 should have $1.23 outbound")

				account1Balance := grouped[accountID1]
				Expect(account1Balance.ToCents()).To(Equal(45678), "account1 should have $456.78 outbound")
			})
		})
	})
})
