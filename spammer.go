package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosclient"
)

// getChainInfo fetches chain information from the registry
func getChainInfo(chainName string) (string, string, error) {
	registry := NewChainRegistry()
	if err := registry.FetchChains(); err != nil {
		return "", "", fmt.Errorf("failed to fetch chains: %w", err)
	}

	chain, exists := registry.Chains[chainName]
	if !exists {
		return "", "", fmt.Errorf("chain '%s' not found in registry", chainName)
	}

	// Enrich the chain to get full details
	if err := EnrichChain(&chain); err != nil {
		return "", "", fmt.Errorf("failed to enrich chain '%s': %w", chainName, err)
	}

	// Get RPC endpoint
	if len(chain.APIs.RPC) == 0 {
		return "", "", fmt.Errorf("no RPC endpoints found for chain '%s'", chainName)
	}
	rpcEndpoint := chain.APIs.RPC[0].Address

	// Get bech32 prefix
	bech32Prefix := chain.Bech32Prefix
	if bech32Prefix == "" {
		return "", "", fmt.Errorf("no bech32 prefix found for chain '%s'", chainName)
	}

	return rpcEndpoint, bech32Prefix, nil
}

// spamTransactions starts the transaction spamming process
func spamTransactions(ctx context.Context, config Config) error {
	// Get chain information from registry
	rpcEndpoint, bech32Prefix, err := getChainInfo(config.Chain)
	if err != nil {
		return fmt.Errorf("failed to get chain info: %w", err)
	}

	// Get keyring home directory
	keyringDir, err := getKeyringHome()
	if err != nil {
		return fmt.Errorf("failed to get keyring home: %w", err)
	}

	// Initialize cosmos client with configuration
	client, err := cosmosclient.New(ctx,
		cosmosclient.WithNodeAddress(rpcEndpoint),
		cosmosclient.WithFees(config.Fees),
		cosmosclient.WithBech32Prefix(bech32Prefix),
		cosmosclient.WithKeyringDir(keyringDir),
		cosmosclient.WithKeyringBackend(DefaultKeyringBackend),
		cosmosclient.WithKeyringServiceName(DefaultKeyringServiceName),
	)
	if err != nil {
		return fmt.Errorf("failed to create cosmos client: %w", err)
	}

	// Get account from cosmos client's keyring
	account, err := client.Account(config.Account)
	if err != nil {
		return fmt.Errorf("failed to get account '%s' from cosmos client keyring: %w", config.Account, err)
	}

	// Get account address for verification
	accountAddr, err := account.Address(bech32Prefix)
	if err != nil {
		return fmt.Errorf("failed to get account '%s' from keyring: %w", config.Account, err)
	}

	// Check if account exists on the blockchain
	if err := verifyAccountExists(ctx, client, accountAddr); err != nil {
		return fmt.Errorf("account verification failed: %w", err)
	}

	// Parse the fees to get the amount for self-transfers
	amount, err := parseAmount(config.Fees)
	if err != nil {
		return fmt.Errorf("failed to parse fees as amount: %w", err)
	}

	// Create ticker for rate limiting
	interval := time.Second / time.Duration(config.TPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	txCount := 0
	for {
		select {
		case <-ticker.C:
			if err := sendTransaction(ctx, client, account, config, amount, txCount, bech32Prefix, config.Memo); err != nil {
				log.Printf("❌ Failed to send transaction: %v", err)
				continue
			}
			txCount++
			if txCount%config.TPS == 0 {
				fmt.Printf("✅ Sent %d transactions (Rate: %d TPS)\n", txCount, config.TPS)
			}
		case <-ctx.Done():
			fmt.Printf("Sent %d transactions total.\n", txCount)
			return nil
		}
	}
}

// sendTransaction sends a bank transfer transaction to self with a specified memo.
func sendTransaction(ctx context.Context, client cosmosclient.Client, account cosmosaccount.Account, config Config, amount sdk.Coins, txNum int, addressPrefix, memo string) error {
	txCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get account address for self-transfer using the chain's bech32 prefix
	accountAddr, err := account.Address(addressPrefix)
	if err != nil {
		return fmt.Errorf("failed to get account address: %w", err)
	}

	// Create and broadcast bank send transaction to self
	bankSendMsg := &banktypes.MsgSend{
		FromAddress: accountAddr,
		ToAddress:   accountAddr,
		Amount:      amount,
	}

	txService, err := client.CreateTxWithOptions(
		ctx,
		account,
		cosmosclient.TxOptions{
			Memo: memo,
			Fees: config.Fees,
		},
		bankSendMsg,
	)
	if err != nil {
		return fmt.Errorf("failed to create bank send transaction: %w", err)
	}

	// Broadcast the transaction
	response, err := txService.Broadcast(txCtx)
	if err != nil || response.Code != 0 {
		return fmt.Errorf("failed to broadcast transaction: %w, code: %d", err, response.Code)
	}

	// Log transaction details periodically
	if txNum%100 == 0 {
		log.Printf("🔗 Transaction #%d broadcasted with hash: %s, memo: %s", txNum, response.TxHash, config.Memo)
	}

	return nil
}

// parseAmount parses a string like "1000uatom" or "1000uatom,500stake" into sdk.Coins
func parseAmount(amountStr string) (sdk.Coins, error) {
	if amountStr == "" {
		return nil, fmt.Errorf("amount string cannot be empty")
	}

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coins: %w", err)
	}

	if coins.IsZero() {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	return coins, nil
}

// verifyAccountExists checks if an account exists on the blockchain
func verifyAccountExists(ctx context.Context, client cosmosclient.Client, address string) error {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Query the account to see if it exists
	queryClient := authtypes.NewQueryClient(client.Context())

	_, err := queryClient.Account(queryCtx, &authtypes.QueryAccountRequest{
		Address: address,
	})

	if err != nil {
		return fmt.Errorf("account %s not found on blockchain - please fund this account first: %w", address, err)
	}

	return nil
}
