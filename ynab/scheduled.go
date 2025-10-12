package ynab

import (
	"fmt"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
)

// IsScheduledAfterInclusive returns true if the scheduled transaction is scheduled for after the given date.
func IsScheduledAfterInclusive(
	scheduledTransactionSummary ynab.ScheduledTransactionSummary,
	dateTime time.Time,
) (bool, error) {
	nextDate, err := parseScheduledDate(scheduledTransactionSummary)
	if err != nil {
		return false, err
	}

	return nextDate.Equal(dateTime) || nextDate.After(dateTime), nil
}

// IsScheduledBeforeInclusive returns true if the scheduled transaction is scheduled for before the given date.
func IsScheduledBeforeInclusive(
	scheduledTransactionSummary ynab.ScheduledTransactionSummary,
	dateTime time.Time,
) (bool, error) {
	nextDate, err := parseScheduledDate(scheduledTransactionSummary)
	if err != nil {
		return false, err
	}

	return nextDate.Equal(dateTime) || nextDate.Before(dateTime), nil
}

func parseScheduledDate(
	scheduledTransactionSummary ynab.ScheduledTransactionSummary,
) (time.Time, error) {
	nextDate, err := time.Parse(time.DateOnly, scheduledTransactionSummary.DateNext)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse 'date next' value of '%s': %w", scheduledTransactionSummary.DateNext, err)
	}

	return nextDate, nil
}
