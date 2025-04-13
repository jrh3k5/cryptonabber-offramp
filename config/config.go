package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	ChainID          int                 `yaml:"chain_id"`
	ContractAddress  string              `yaml:"contract_address"`
	Decimals         int                 `yaml:"decimals"`
	QRCodeType       *string             `yaml:"qr_code_type"`
	RecipientAddress string              `yaml:"recipient_address"`
	YNABBudgetName   string              `yaml:"ynab_budget_name"`
	YNABAccounts     *YNABAccountsConfig `yaml:"ynab_accounts"`
}

type YNABAccountsConfig struct {
	FundsOriginAccount    string                      `yaml:"funds_origin_account"`
	FundsRecipientAccount string                      `yaml:"funds_recipient_account"`
	OfframpAccounts       []*YNABOfframpAccountConfig `yaml:"offramp_accounts"`
}

func (c *Config) GetQRCodeType() string {
	if c.QRCodeType == nil {
		return "erc681"
	}

	return *c.QRCodeType
}

type YNABOfframpAccountConfig struct {
	Name               string       `yaml:"name"`                 // The name of the account as it appears in YNAB
	ExcludedFlagColors []string     `yaml:"excluded_flag_colors"` // If specified, this is a list of flag colors to exclude from calculations
	MinimumBalance     *json.Number `yaml:"minimum_balance"`      // If specified, this is the minimum balance to be maintained between now and the given end billing date
}

// MinimumBalanceAsCents returns the minimum balance as cents.
// If there is no minimum balance specified, the returned boolean is false; otherwise, it is true.
func (y *YNABOfframpAccountConfig) MinimumBalanceAsCents() (int, bool, error) {
	if y.MinimumBalance == nil {
		return 0, false, nil
	}

	minimumBalanceString := y.MinimumBalance.String()

	var parsedDollars int
	var parsedCents int

	periodPos := strings.Index(minimumBalanceString, ".")
	if periodPos == -1 {
		parsedDollarsInt64, err := strconv.ParseInt(minimumBalanceString, 10, 64)
		if err != nil {
			return 0, false, fmt.Errorf("failed to parse minimum balance '%s' as an integer: %v", minimumBalanceString, err)
		}

		return int(parsedDollarsInt64) * 100, true, nil
	}

	parsedDollars, err := strconv.Atoi(minimumBalanceString[:periodPos])
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse minimum balance '%s' as an integer: %v", minimumBalanceString, err)
	}

	parsedCents, err = strconv.Atoi(minimumBalanceString[periodPos+1:])
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse minimum balance '%s' as an integer: %v", minimumBalanceString, err)
	}

	return parsedDollars*100 + parsedCents, true, nil
}
