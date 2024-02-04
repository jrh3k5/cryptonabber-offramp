package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-offramp/config"
	"github.com/jrh3k5/cryptonabber-offramp/math"
	"github.com/jrh3k5/cryptonabber-offramp/qr"
	"github.com/mdp/qrterminal"
	"gopkg.in/yaml.v3"
)

func main() {
	ctx := context.Background()

	var file string
	flag.StringVar(&file, "file", "", "the location of the file to be read in as configuration")

	var accessToken string
	flag.StringVar(&accessToken, "access-token", "", "the personal access token used to interact with YNAB's APIs")

	flag.Parse()

	fmt.Printf("Reading configuration from '%s'\n", file)

	config, err := readConfiguration(file)
	if err != nil {
		panic(fmt.Sprintf("Failed to read configuration: %v", err))
	}

	ynabURL, err := url.Parse("https://api.ynab.com/v1/")
	if err != nil {
		// ??? how?
		panic(fmt.Sprintf("unable to parse hard-coded YNAB URL: %v", err))
	}
	ynabClient := ynab.NewClient(ynabURL, http.DefaultClient, accessToken)

	budget, err := getBudget(ynabClient, config.YNABBudgetName)
	if err != nil {
		panic(fmt.Sprintf("Failed to get budget: %v", err))
	} else if budget == nil {
		panic(fmt.Sprintf("No budget found for name '%s'", config.YNABBudgetName))
	}

	offrampAccountsByID, err := mapAccountNamesByID(ynabClient, budget.Id, config.YNABAccounts.OfframpAccounts)
	if err != nil {
		panic(fmt.Sprintf("Failed to map offramp accounts by ID: %v", err))
	}
	offrampAccountIDs := make([]string, 0, len(offrampAccountsByID))
	for offrampAccountID := range offrampAccountsByID {
		offrampAccountIDs = append(offrampAccountIDs, offrampAccountID)
	}

	now := time.Now().UTC()
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := nowDay.Add(7 * 24 * time.Hour)
	endDate := startDate.Add(6 * 24 * time.Hour)

	scheduledTransactions, err := ynabClient.ScheduledTransactionsService.List(budget.Id)
	if err != nil {
		panic(fmt.Sprintf("Failed to get scheduled transactions: %v", err))
	}

	outboundBalances, err := math.CalculateOutboundTransactions(offrampAccountIDs, scheduledTransactions, startDate, endDate)
	if err != nil {
		panic(fmt.Sprintf("Failed to calculate outbound transactions: %v", err))
	}

	outboundCents := 0
	fmt.Printf("Outbound Account Balances for [%s, %s]:\n", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
	for accountID, outboundBalance := range outboundBalances {
		outboundCents += outboundBalance.ToCents()
		accountName := offrampAccountsByID[accountID]
		fmt.Printf("  %s: $%d.%02d\n", accountName, outboundBalance.Dollars, outboundBalance.Cents)
	}

	totalCents := outboundCents % 100
	totalDollars := (outboundCents - totalCents) / 100

	fmt.Printf("Scan the following QR code and send $%d.%02d to the address it presents:\n", totalDollars, totalCents)

	var urlGenerator qr.URLGenerator
	qrCodeType := config.GetQRCodeType()
	switch qrCodeType {
	case "erc681":
		urlGenerator = qr.NewERC681URLGenerator()
	case "recipient_only":
		urlGenerator = qr.NewRecipientAddressURLGenerator()
	default:
		panic(fmt.Sprintf("Unsupported QR code type: %v", qrCodeType))
	}

	qrDetails := &qr.Details{
		ChainID:           config.ChainID,
		ContactAddress:    config.ContractAddress,
		Decimals:          config.Decimals,
		ReceipientAddress: config.RecipientAddress,
		Dollars:           1,  // TODO: calculate this
		Cents:             52, // TODO: calculate this
	}

	url, err := urlGenerator.Generate(ctx, qrDetails)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate QR code URL: %v", err))
	}

	qrterminal.Generate(url, qrterminal.M, os.Stdout)
}

func getBudget(ynabClient *ynab.Client, budgetName string) (*ynab.BudgetSummary, error) {
	budgets, err := ynabClient.BudgetService.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}

	for _, budget := range budgets {
		if budget.Name == budgetName {
			return &budget, nil
		}
	}

	return nil, nil
}

func mapAccountNamesByID(ynabClient *ynab.Client, budgetID string, accountNames []string) (map[string]string, error) {
	accounts, err := ynabClient.AccountsService.List(budgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	mapped := make(map[string]string)
	for _, account := range accounts {
		for _, accountName := range accountNames {
			if account.Name == accountName {
				mapped[account.Id] = accountName
				break
			}
		}
	}

	return mapped, nil
}

func readConfiguration(file string) (*config.Config, error) {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", file, err)
	}

	config := &config.Config{}
	if err := yaml.Unmarshal(fileBytes, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML in file '%s': %w", file, err)
	}

	return config, nil
}
