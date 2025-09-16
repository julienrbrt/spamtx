# spamtx

Go tool to spam txs to a Cosmos SDK based blockchain.

> Currently, this tool only does self bank send with a memo field to save gas.

## Installation

```sh
go build -o spamtx .
```

## Usage

```sh
spamtx <chain> --from <address> --fees <amount> --memo <message> --tps <speed>
```

### Parameters

- `--from`: Your account address (must exist in keyring)
- `--fees`: Transaction fees (e.g., "1000uatom")
- `--memo`: Message to include in each transaction
- `--tps`: Transactions per second rate limit

### Example

```sh
./spamtx \
  cosmoshub \
  --from alice \
  --fees 1000uatom \
  --memo "spam test" \
  --tps 5
```

## Stack

- [cosmosclient](https://pkg.go.dev/github.com/ignite/cli/ignite/pkg/cosmosclient)
- [cobra](https://pkg.go.dev/github.com/spf13/cobra) for CLI
