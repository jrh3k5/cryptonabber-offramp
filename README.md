# cryptonabber-offramp

A CLI tool to make offramping with YNAB integration easier. It does the following:

* For the configured offramp accounts, calculate the total outbound transactions, per account, that are outbound for a week, starting from a week from today
  * e.g., if you are executing this on February 2, it will retrieve outbound transactions for February 9 through Febrary 15, inclusive
* For each account and for the configured source and destination account, generate the transactions in YNAB tracking:
 * The send of the funds from your wallet to the address being used for offramping
 * Transactions from the offramp address to each of the accounts to which funds are being offramped
* Generates a QR code that can be scanned to send the funds to be used for offramping

## Usage

This tool receives two parameters:

* An `access-token` parameter containing your YNAB personal access token
* A `file` parameter describing the location of the YAML file to drive the behavior of this tool

It will look like:

```
./cryptonabber-offramp --access-token="FAKE04678" --file="config.yaml"
```

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
    - "<a list of the names of accounts to which the funds are to be offramped>"
```

#### QR Code Type

By default, this tool generates an ERC-681-compliant QR code. You can set the YAML file with the following values to change that:

* `erc681`: the default; this generates an ERC-681-compliant QR code
* `recipient_only`: the QR code will merely contain the address to which the funds are to be sent