# cryptonabber-offramp

A CLI tool to make offramping with YNAB integration easier. It does the following:

* For the configured offramp accounts, calculate the total outbound transactions, per account, that are outbound for a week, starting from a week from today
  * e.g., if you are executing this on February 2, it will retrieve outbound transactions for February 9 through Febrary 15, inclusive
* For each account and for the configured source and destination account, generate the transactions in YNAB tracking:
 * The send of the funds from your wallet to the address being used for offramping
 * Transactions from the offramp address to each of the accounts to which funds are being offramped
* Generates a QR code that can be scanned to send the funds to be used for offramping

## Usage

### Prerequisites

* You must have a [YNAB](https://ynab.com) account with a budget and accounts to set up
* You must have a registered OAuth client ID and secret as described [here](https://api.ynab.com/#oauth-applications).

### Executing the Program

You can either supply the OAuth credentials interactively by executing this application as:

```
/cryptonabber-offramp --interactive
```

...or you can supply the OAuth credentials non-interactively by executing this application as:

```
/cryptonabber-offramp --oauth-client-id=<client ID> --oauth-client-secret=<client secret>
```

You can provide the following optional arguments:

* `--file`: by default, this application looks for a file called `config.yaml` in the local directory; if you would like to use a different filename or location, you can use this parameter to specify that

### Configuration

Below describes the expected structure of the YAML configuration file:

```yaml
recipient_address: "<the address to which the funds are to be sent for offramping>"
contract_address: "<the contract adrdess of funds to be sent>"
decimals: <the number of decimals for the funds to be sent>
chain_id: <the ID of the chain on which the funds are to be sent>
qr_code_type: "<optional; the type of QR code to be generated; defaults to erc681 if not specified>"
ynab_budget_name: "<the name of the budget under which the involved accounts reside>"
ynab_accounts:
  funds_origin_account: "<the name of the account you use to track the wallet from which you'll be sending funds>"
  funds_recipient_account: "<the name of the account you use to track the address to which you'll be sending funds for offboarding>"
  offramp_accounts:
    - name: "<the name of the offramp destination account as it appears in YNAB>"
      minimum_balance: <optional; the minimum balance that should be left in the account after all transactions through the given end date have been executed>
      excluded_flag_colors:
        - green
        - <optional flag colors of transactions to be excluded from the calculation>
```

#### QR Code Type

By default, this tool generates an ERC-681-compliant QR code. You can set the YAML file with the following values to change that:

* `erc681`: the default; this generates an ERC-681-compliant QR code
* `recipient_only`: the QR code will merely contain the address to which the funds are to be sent

### Optional Arguments

You can provide the following optional arguments at runtime to control the behavior of the application:

* `dry-run`: if provided, the application will only calculate the outbound balances and print them; no QR code or YNAB transactions will be generated

## Privacy Policy

This application does not persist any information given to this application. It only uses the access granted to your account within YNAB to read upcoming transactions and create inter-account transfers funding those upcoming transactions, as defined by the configuration you provide to this tool.

No data given to this application or read from YNAB is shared with any third parties.
