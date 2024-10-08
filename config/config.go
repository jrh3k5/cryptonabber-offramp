package config

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
	Name               string   `yaml:"name"`
	ExcludedFlagColors []string `yaml:"excluded_flag_colors"`
}
