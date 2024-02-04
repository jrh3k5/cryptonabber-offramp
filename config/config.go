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
	OfframpAccounts []string `yaml:"offramp_accounts"`
}

func (c *Config) GetQRCodeType() string {
	if c.QRCodeType == nil {
		return "erc681"
	}

	return *c.QRCodeType
}
