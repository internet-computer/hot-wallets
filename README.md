# Hot Wallet Detection Script (for Exchange Accounts?)

This script identifies and filters accounts based on transaction count and balance thresholds. It fetches account data from the Internet Computer Ledger API and outputs the results to a JSON file.

## Quick Start

To run the script with default settings:

```sh
go run main.go
```

The script should complete its execution in approximately **10 seconds** with the current configuration.

## Features

- Filters accounts based on a minimum transaction count (`MIN_TX_COUNT`).
- Optionally filters accounts with large balances (`MIN_ICP_BAL`) using the `-enable-large-balance` flag.
- Outputs the filtered accounts to a JSON file (`accounts.json`).
- Includes predefined names for known exchange accounts.

## Configuration

The application includes several constants that can be adjusted to modify its behavior and output. These constants are defined in the `main.go` file:

```go
const (
    FILE_PATH    = "accounts.json"   // Path to save the output JSON file
    MIN_TX_COUNT = 10_000            // Minimum transaction count for filtering accounts, default 10k.
    MIN_ICP_BAL  = 250_000           // Minimum ICP balance (in e8s) for filtering accounts, default 250k ICP.
    ACTIVE_THR   = 60 * 60 * 24 * 30 // Threshold for account activity, default is 30 days (in seconds).
)
```

## Flags

The script includes a flag to modify its behavior:

- `-enable-large-balance`: When this flag is set, the script will include accounts with large balances (greater than `MIN_ICP_BAL`) in the output. By default, this flag is disabled.

### Usage

To run the script without including accounts with large balances:

```sh
go run main.go
```

To enable the inclusion of accounts with large balances:

```sh
go run main.go -enable-large-balance
```

## Output

The script generates a JSON file (`accounts.json`) containing the filtered accounts. Each account includes the following fields:

- `active`: A boolean field indicating whether the account is considered active.
- `large_balance`: A boolean field indicating whether the account's balance exceeds the `MIN_ICP_BAL` threshold.
- `name`: The name of the account (if known).
- `account_identifier`: The unique identifier of the account.
- `balance`: The balance of the account in e8s.
- `transaction_count`: The total number of transactions for the account.
- `updated_at`: The timestamp of the last update.

## Known Exchange Accounts

The script includes a predefined list of known exchange accounts with their names. These accounts are automatically labeled in the output.
The list can be found in the `names` map in the `main.go` file.
