package math_test

import (
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-offramp/v3/math"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Balance", func() {
	Context("CalculateMinimumBalanceAdjustment", func() {
		When("there are no applicable transactions", func() {
			It("calculates the amount needed to adjust the existing balance up to the minimum balance", func() {
				account := ynab.Account{
					Id:      "6dba52e7-3367-4d29-b246-95f7ff83495d",
					Balance: 3000, // 3.00 USD
				}

				now := time.Now()

				adjustment, err := math.CalculateMinimumBalanceAdjustment(
					account,
					[]ynab.ScheduledTransactionDetail{
						{
							ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
								AccountId: "other-account",
								Amount:    -1000, // 1.00 USD
								DateNext:  now.Format(time.DateOnly),
							},
						},
					},
					1000, // 10.00 USD
					now.Add(24*time.Hour),
					false,
				)

				Expect(err).NotTo(HaveOccurred(), "calculating the minimum balance adjustment should not fail")
				Expect(adjustment.ToCents()).To(Equal(700), "the minimum balance adjustment should be the amount needed to adjust the existing balance up to the minimum balance")
			})
		})

		When("the post-transaction balance meets or exceeds the minimum balance", func() {
			It("returns no adjustment", func() {
				accountID := "27a41dc9-5d2b-4404-a4aa-b83e28d50586"

				account := ynab.Account{
					Id:      accountID,
					Balance: 20000, // 20.00 USD
				}

				now := time.Now()

				adjustment, err := math.CalculateMinimumBalanceAdjustment(
					account,
					[]ynab.ScheduledTransactionDetail{
						{
							ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
								AccountId: accountID,
								Amount:    -1000, // 1.00 USD
								DateNext:  now.Format(time.DateOnly),
							},
						},
					},
					1000, // 10.00 USD, which should be less than the account balance
					now.Add(24*time.Hour),
					false,
				)

				Expect(err).NotTo(HaveOccurred(), "calculating the minimum balance adjustment should not fail")
				Expect(adjustment.ToCents()).To(Equal(0), "there should be no adjustment necessary to reach the minimum balance")
			})
		})

		When("the post-transaction balance is below the minimum balance", func() {
			It("calculates the amount needed to adjust the existing balance up to the minimum balance", func() {
				accountID := "8e0e8d3f-5411-45e8-b47e-ae0a17486109"

				account := ynab.Account{
					Id:      accountID,
					Balance: 2000, // 2.00 USD
				}

				now := time.Now()

				adjustment, err := math.CalculateMinimumBalanceAdjustment(
					account,
					[]ynab.ScheduledTransactionDetail{
						{
							ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
								AccountId: accountID,
								Amount:    -1000, // 1.00 USD
								DateNext:  now.Format(time.DateOnly),
							},
						},
					},
					1000, // 10.00 USD
					now.Add(24*time.Hour),
					false,
				)

				Expect(err).NotTo(HaveOccurred(), "calculating the minimum balance adjustment should not fail")
				Expect(adjustment.ToCents()).To(Equal(900), "the minimum balance adjustment should be the amount needed to adjust the existing balance up to the minimum balance")
			})
		})
	})

	Context("CalculateEffectiveBalanceThrough", func() {
		When("there are transactions before now", func() {
			When("the transaction is before midnight today", func() {
				It("does not count it in the balance", func() {
					now := time.Now()

					yesterdayDateString := now.Add(-24 * time.Hour).Format(time.DateOnly)

					nowDateString := time.Now().Format(time.DateOnly)

					balance, err := math.CalculateEffectiveBalanceThrough(
						0,
						[]ynab.ScheduledTransactionDetail{
							{
								ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
									DateNext: nowDateString,
									Amount:   -200,
								},
							},
							{
								ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
									DateNext: yesterdayDateString,
									Amount:   -100,
								},
							},
							{
								ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
									DateNext: nowDateString,
									Amount:   -50,
								},
							},
						},
						now.Add(24*time.Hour),
					)

					Expect(err).ToNot(HaveOccurred(), "calculating the balance should not fail")
					Expect(balance).To(Equal(-250), "the transaction from the past should not be included; only today's transactions should be included")
				})
			})
		})

		When("there are transactions after the end date", func() {
			It("does not count them in the balance", func() {
				now := time.Now()

				nowDateString := now.Format(time.DateOnly)
				tomorrowDateString := now.Add(24 * time.Hour).Format(time.DateOnly)
				dayAfterTomorrowDateString := now.Add(48 * time.Hour).Format(time.DateOnly)

				balance, err := math.CalculateEffectiveBalanceThrough(
					0,
					[]ynab.ScheduledTransactionDetail{
						{
							ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
								DateNext: nowDateString,
								Amount:   -200,
							},
						},
						{
							ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
								DateNext: tomorrowDateString,
								Amount:   -100,
							},
						},
						{
							ScheduledTransactionSummary: ynab.ScheduledTransactionSummary{
								DateNext: dayAfterTomorrowDateString,
								Amount:   -50,
							},
						},
					},
					now.Add(24*time.Hour),
				)

				Expect(err).ToNot(HaveOccurred(), "calculating the balance should not fail")
				Expect(balance).To(Equal(-300), "the transaction from the days after the given end date should not be included; only today's and tomorrow's transactions should be included")
			})
		})
	})
})
