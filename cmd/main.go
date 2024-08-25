package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/davidsteinsland/ynab-go/ynab"
	"github.com/jrh3k5/cryptonabber-offramp/v2/config"
	"github.com/jrh3k5/cryptonabber-offramp/v2/math"
	"github.com/jrh3k5/cryptonabber-offramp/v2/qr"
	"github.com/manifoldco/promptui"
	"github.com/mdp/qrterminal"
	"gopkg.in/yaml.v3"

	cliynab "github.com/jrh3k5/cryptonabber-offramp/v2/ynab"
)

func main() {
	ctx := context.Background()

	var file string
	flag.StringVar(&file, "file", "", "the location of the file to be read in as configuration")

	var accessToken string
	flag.StringVar(&accessToken, "access-token", "", "the personal access token used to interact with YNAB's APIs")

	flag.Parse()

	fmt.Printf("Reading configuration from '%s'\n", file)

	appConfig, err := readConfiguration(file)
	if err != nil {
		panic(fmt.Sprintf("Failed to read configuration: %v", err))
	}

	ynabURL, err := url.Parse("https://api.ynab.com/v1/")
	if err != nil {
		// ??? how?
		panic(fmt.Sprintf("unable to parse hard-coded YNAB URL: %v", err))
	}
	ynabClient := ynab.NewClient(ynabURL, http.DefaultClient, accessToken)

	budget, err := getBudget(ynabClient, appConfig.YNABBudgetName)
	if err != nil {
		panic(fmt.Sprintf("Failed to get budget: %v", err))
	} else if budget == nil {
		panic(fmt.Sprintf("No budget found for name '%s'", appConfig.YNABBudgetName))
	}

	offrampAccountNames := make([]string, len(appConfig.YNABAccounts.OfframpAccounts))
	for accountIndex, offrampAccount := range appConfig.YNABAccounts.OfframpAccounts {
		offrampAccountNames[accountIndex] = offrampAccount.Name
	}

	allAccountNames := toUnique(append(offrampAccountNames, appConfig.YNABAccounts.FundsOriginAccount, appConfig.YNABAccounts.FundsRecipientAccount))
	accountNamesByID, err := mapAccountNamesByID(ynabClient, budget.Id, allAccountNames)
	if err != nil {
		panic(fmt.Sprintf("Failed to map offramp accounts by ID: %v", err))
	}

	allAccountIDs, err := getAccountIDs(accountNamesByID, allAccountNames)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve all account IDs: %v", err))
	}

	offrampAccountIDs, err := getAccountIDs(accountNamesByID, offrampAccountNames)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve account IDs for offramp accounts: %v", err))
	}

	var fundsOriginAccountID string
	if accountIDs, err := getAccountIDs(accountNamesByID, []string{appConfig.YNABAccounts.FundsOriginAccount}); err != nil {
		panic(fmt.Sprintf("Failed to resolve funds origin account ID: %v", err))
	} else {
		fundsOriginAccountID = accountIDs[0]
	}

	var recipientAccountID string
	if accountIDs, err := getAccountIDs(accountNamesByID, []string{appConfig.YNABAccounts.FundsRecipientAccount}); err != nil {
		panic(fmt.Sprintf("Failed to resolve recipient account ID: %v", err))
	} else {
		recipientAccountID = accountIDs[0]
	}

	now := time.Now().Local()
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := nowDay.Add(7 * 24 * time.Hour)

	isValidDate := func(v string) error {
		_, parseErr := time.Parse(time.DateOnly, v)
		return parseErr
	}

	startDatePrompt := &promptui.Prompt{
		Label:    "Start date",
		Default:  startDate.Format(time.DateOnly),
		Validate: isValidDate,
	}
	startDateStr, startDatePromptErr := startDatePrompt.Run()
	if startDatePromptErr != nil {
		panic(fmt.Sprintf("Failed to get start date: %v", startDatePromptErr))
	}
	// the Validate function in the prompt ensures that it's a valid date value
	startDate, _ = time.Parse(time.DateOnly, startDateStr)

	endDate := startDate.Add(6 * 24 * time.Hour)
	endDatePrompt := &promptui.Prompt{
		Label:    "End date",
		Default:  endDate.Format(time.DateOnly),
		Validate: isValidDate,
	}
	endDateStr, endDatePromptErr := endDatePrompt.Run()
	if endDatePromptErr != nil {
		panic(fmt.Sprintf("Failed to get end date: %v", endDatePromptErr))
	}
	// the Validate function in the prompt ensures that it's a valid date value
	endDate, _ = time.Parse(time.DateOnly, endDateStr)

	scheduledTransactions, err := ynabClient.ScheduledTransactionsService.List(budget.Id)
	if err != nil {
		panic(fmt.Sprintf("Failed to get scheduled transactions: %v", err))
	}

	excludedColorsByAccountID := make(map[string][]string)
	for _, offrampAccount := range appConfig.YNABAccounts.OfframpAccounts {
		for accountID, accountName := range accountNamesByID {
			if accountName == offrampAccount.Name {
				excludedColorsByAccountID[accountID] = offrampAccount.ExcludedFlagColors
			}
		}
	}

	outboundBalances, err := math.CalculateOutboundTransactions(offrampAccountIDs, excludedColorsByAccountID, scheduledTransactions, startDate, endDate)
	if err != nil {
		panic(fmt.Sprintf("Failed to calculate outbound transactions: %v", err))
	}

	outboundCents := 0
	fmt.Printf("Outbound Account Balances for [%s, %s]:\n", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
	for accountID, outboundBalance := range outboundBalances {
		outboundCents += outboundBalance.ToCents()
		accountName := accountNamesByID[accountID]
		fmt.Printf("  %s: $%d.%02d\n", accountName, outboundBalance.Dollars, outboundBalance.Cents)
	}

	if outboundCents == 0 {
		fmt.Println("No upcoming transactions require funding; exiting")
		return
	}

	fmt.Println("Creating transactions in YNAB...")

	payeeIDsByAccountIDs, err := getTransferPayeeIDsByAccountID(ynabClient, budget.Id, allAccountIDs, accountNamesByID)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve transfer payee IDs by account ID: %v", err))
	}

	transactions, err := cliynab.CreateTransactions(
		fundsOriginAccountID,
		recipientAccountID,
		outboundBalances,
		accountNamesByID,
		payeeIDsByAccountIDs,
		startDate,
		endDate,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create transactions to send to YNAB: %v", err))
	}

	_, err = ynabClient.TransactionsService.CreateBulk(budget.Id, transactions)
	if err != nil {
		panic(fmt.Sprintf("Failed to create transfer transactions in YNAB: %v", err))
	}

	totalCents := outboundCents % 100
	totalDollars := (outboundCents - totalCents) / 100

	fmt.Printf("Scan the following QR code and send $%d.%02d to the address it presents:\n", totalDollars, totalCents)

	var urlGenerator qr.URLGenerator
	qrCodeType := appConfig.GetQRCodeType()
	switch qrCodeType {
	case "erc681":
		urlGenerator = qr.NewERC681URLGenerator()
	case "recipient_only":
		urlGenerator = qr.NewRecipientAddressURLGenerator()
	default:
		panic(fmt.Sprintf("Unsupported QR code type: %v", qrCodeType))
	}

	qrDetails := &qr.Details{
		ChainID:           appConfig.ChainID,
		ContactAddress:    appConfig.ContractAddress,
		Decimals:          appConfig.Decimals,
		ReceipientAddress: appConfig.RecipientAddress,
		Dollars:           totalDollars,
		Cents:             totalCents,
	}

	url, err := urlGenerator.Generate(ctx, qrDetails)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate QR code URL: %v", err))
	}

	qrterminal.Generate(url, qrterminal.M, os.Stdout)
}

func getAccountIDs(accountNamesByID map[string]string, accountNames []string) ([]string, error) {
	accountIDs := make([]string, 0, len(accountNames))
	for accountID, accountName := range accountNamesByID {
		for _, desiredAccountName := range accountNames {
			if accountName == desiredAccountName {
				accountIDs = append(accountIDs, accountID)
				break
			}
		}
	}

	if len(accountIDs) != len(accountNames) {
		return nil, fmt.Errorf("%d account names (['%s']) were requested, but only %d account IDs (['%s']) were resolved.",
			len(accountNames),
			strings.Join(accountNames, "', '"),
			len(accountIDs),
			strings.Join(accountIDs, "', '"))
	}

	return accountIDs, nil
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

func getTransferPayeeIDsByAccountID(ynabClient *ynab.Client, budgetID string, accountIDs []string, accountNamesByID map[string]string) (map[string]string, error) {
	allPayees, err := ynabClient.PayeesService.List(budgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all payees: %w", err)
	}

	mappedPayeeIDs := make(map[string]string)
	for _, accountID := range accountIDs {
		accountName, hasAccountName := accountNamesByID[accountID]
		if !hasAccountName {
			return nil, fmt.Errorf("can't resolve payee ID; account ID '%s' has no known name", accountID)
		}

		payeeName := "Transfer : " + accountName
		for _, payee := range allPayees {
			if payee.Name == payeeName {
				mappedPayeeIDs[accountID] = payee.Id
			}
		}
	}

	if len(mappedPayeeIDs) != len(accountIDs) {
		return nil, fmt.Errorf("%d account IDs were requested for payee mapping, but only %d were resolved", len(accountIDs), len(mappedPayeeIDs))
	}

	return mappedPayeeIDs, nil
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

func toUnique(values []string) []string {
	uniqueMap := make(map[string]any)
	for _, value := range values {
		uniqueMap[value] = nil
	}

	uniqueValues := make([]string, 0, len(uniqueMap))
	for mapKey := range uniqueMap {
		uniqueValues = append(uniqueValues, mapKey)
	}

	return uniqueValues
}
