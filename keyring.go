package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ignite/cli/v29/ignite/pkg/cosmosaccount"
)

const (
	// DefaultKeyringServiceName is the default service name for the keyring
	DefaultKeyringServiceName = "spamtx"

	// DefaultKeyringBackend is the default keyring backend
	DefaultKeyringBackend = cosmosaccount.KeyringTest
)

// initializeKeyring creates and configures a cosmos keyring for the specified chain
func initializeKeyring(chainName string) (cosmosaccount.Registry, string, error) {
	if chainName == "" {
		return cosmosaccount.Registry{}, "", fmt.Errorf("chain name cannot be empty")
	}

	// Get chain information to determine bech32 prefix
	_, bech32Prefix, err := getChainInfo(chainName)
	if err != nil {
		return cosmosaccount.Registry{}, "", fmt.Errorf("failed to get chain info: %w", err)
	}

	// Create keyring home directory
	homeDir, err := getKeyringHome()
	if err != nil {
		return cosmosaccount.Registry{}, "", fmt.Errorf("failed to get keyring home: %w", err)
	}

	// Create the keyring with chain-specific configuration
	registry, err := cosmosaccount.New(
		cosmosaccount.WithHome(homeDir),
		cosmosaccount.WithKeyringBackend(DefaultKeyringBackend),
		cosmosaccount.WithKeyringServiceName(DefaultKeyringServiceName),
		cosmosaccount.WithBech32Prefix(bech32Prefix),
	)
	if err != nil {
		return cosmosaccount.Registry{}, "", fmt.Errorf("failed to create keyring: %w", err)
	}

	return registry, bech32Prefix, nil
}

// getKeyringHome returns the home directory for the keyring
func getKeyringHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	keyringHome := filepath.Join(homeDir, ".spamtx", "keyring")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(keyringHome, 0755); err != nil {
		return "", fmt.Errorf("failed to create keyring directory: %w", err)
	}

	return keyringHome, nil
}

// getOrCreateAccount retrieves an existing account or creates a new one if it doesn't exist
func getOrCreateAccount(registry cosmosaccount.Registry, accountName string) (cosmosaccount.Account, bool, error) {
	if err := validateAccountName(accountName); err != nil {
		return cosmosaccount.Account{}, false, err
	}

	// Try to get existing account
	account, err := registry.GetByName(accountName)
	if err == nil {
		return account, false, nil
	}

	// If account doesn't exist, create it
	var accountDoesNotExistError *cosmosaccount.AccountDoesNotExistError
	if errors.As(err, &accountDoesNotExistError) {
		fmt.Printf("Account '%s' not found. Creating new account...\n", accountName)

		account, mnemonic, err := registry.Create(accountName)
		if err != nil {
			return cosmosaccount.Account{}, false, fmt.Errorf("failed to create account: %w", err)
		}

		fmt.Printf("✅ Created new account '%s'\n", accountName)
		fmt.Printf("🔑 Mnemonic: %s\n", mnemonic)
		fmt.Printf("⚠️ Please save this mnemonic in a secure location!\n")

		return account, true, nil
	}

	return cosmosaccount.Account{}, false, fmt.Errorf("failed to get account: %w", err)
}

// validateAccountName validates that the account name is acceptable
func validateAccountName(name string) error {
	if name == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("account name cannot contain spaces")
	}

	if len(name) < 2 {
		return fmt.Errorf("account name must be at least 2 characters long")
	}

	if len(name) > 50 {
		return fmt.Errorf("account name must be less than 50 characters long")
	}

	return nil
}

// listAccounts lists all accounts in the keyring
func listAccounts(registry cosmosaccount.Registry, bech32Prefix string) error {
	accounts, err := registry.List()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts found in keyring.")
		return nil
	}

	fmt.Printf("Found %d account(s) in keyring:\n", len(accounts))
	for i, account := range accounts {
		address, err := account.Address(bech32Prefix)
		if err != nil {
			fmt.Printf("%d. %s (error getting address: %v)\n", i+1, account.Name, err)
			continue
		}
		fmt.Printf("%d. %s (%s)\n", i+1, account.Name, address)
	}

	return nil
}

// importAccount imports an account from a mnemonic or private key
func importAccount(registry cosmosaccount.Registry, name, secret, passphrase, bech32prefix string) error {
	if err := validateAccountName(name); err != nil {
		return err
	}

	if secret == "" {
		return fmt.Errorf("secret (mnemonic or private key) cannot be empty")
	}

	account, err := registry.Import(name, secret, passphrase)
	if err != nil {
		return fmt.Errorf("failed to import account: %w", err)
	}

	address, err := account.Address(bech32prefix)
	if err != nil {
		return fmt.Errorf("failed to get account address: %w", err)
	}

	fmt.Printf("✅ Successfully imported account '%s' (%s)\n", name, address)
	return nil
}

// deleteAccount removes an account from the keyring
func deleteAccount(registry cosmosaccount.Registry, name string) error {
	if err := validateAccountName(name); err != nil {
		return err
	}

	// Check if account exists first
	_, err := registry.GetByName(name)
	if err != nil {
		var accountDoesNotExistError *cosmosaccount.AccountDoesNotExistError
		if errors.As(err, &accountDoesNotExistError) {
			return fmt.Errorf("account '%s' does not exist", name)
		}
		return fmt.Errorf("failed to check account existence: %w", err)
	}

	err = registry.DeleteByName(name)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	fmt.Printf("✅ Successfully deleted account '%s'\n", name)
	return nil
}
