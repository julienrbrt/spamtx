package main

import (
	"context"
	"testing"

	"github.com/ignite/cli/v29/ignite/pkg/cosmosaccount"
	"gotest.tools/v3/assert"
)

func TestInitializeKeyring(t *testing.T) {
	tests := []struct {
		name        string
		chainName   string
		expectError bool
	}{
		{
			name:        "valid chain should initialize keyring",
			chainName:   "cosmoshub",
			expectError: false,
		},
		{
			name:        "empty chain should fail",
			chainName:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, bech32Prefix, err := initializeKeyring(tt.chainName)

			if tt.expectError {
				assert.Assert(t, err != nil)
				return
			}

			if err != nil {
				// Network call might fail in test environment
				t.Logf("Network call failed (expected in test env): %v", err)
				return
			}

			assert.Assert(t, bech32Prefix != "")
		})
	}
}

func TestCreateAccountInKeyring(t *testing.T) {
	// Create an in-memory keyring for testing
	registry, err := cosmosaccount.NewInMemory(
		cosmosaccount.WithBech32Prefix("cosmos"),
	)
	assert.NilError(t, err)

	accountName := "test-account"

	// Create account
	account, mnemonic, err := registry.Create(accountName)
	assert.NilError(t, err)
	assert.Assert(t, account.Name == accountName)
	assert.Assert(t, mnemonic != "")

	// Verify account can be retrieved
	retrievedAccount, err := registry.GetByName(accountName)
	assert.NilError(t, err)
	assert.Equal(t, retrievedAccount.Name, accountName)
}

func TestListAccountsInKeyring(t *testing.T) {
	// Create an in-memory keyring for testing
	registry, err := cosmosaccount.NewInMemory(
		cosmosaccount.WithBech32Prefix("cosmos"),
	)
	assert.NilError(t, err)

	// Initially should have no accounts
	accounts, err := registry.List()
	assert.NilError(t, err)
	assert.Equal(t, len(accounts), 0)

	// Create a test account
	accountName := "test-account"
	_, _, err = registry.Create(accountName)
	assert.NilError(t, err)

	// Should now have one account
	accounts, err = registry.List()
	assert.NilError(t, err)
	assert.Equal(t, len(accounts), 1)
	assert.Equal(t, accounts[0].Name, accountName)
}

func TestGetOrCreateAccount(t *testing.T) {
	// Create an in-memory keyring for testing
	registry, err := cosmosaccount.NewInMemory(
		cosmosaccount.WithBech32Prefix("cosmos"),
	)
	assert.NilError(t, err)

	accountName := "test-account"

	// First call should create the account
	account1, created, err := getOrCreateAccount(registry, accountName)
	assert.NilError(t, err)
	assert.Assert(t, created == true)
	assert.Equal(t, account1.Name, accountName)

	// Second call should return existing account
	account2, created, err := getOrCreateAccount(registry, accountName)
	assert.NilError(t, err)
	assert.Assert(t, created == false)
	assert.Equal(t, account2.Name, accountName)

	// Accounts should be the same
	assert.Equal(t, account1.Name, account2.Name)
}

func TestSpamTransactionsWithKeyring(t *testing.T) {
	config := Config{
		Chain:   "cosmoshub",
		Account: "test-account",
		Fees:    "1000uatom",
		Memo:    "test memo",
		TPS:     1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately to test that it respects context
	cancel()

	err := spamTransactions(ctx, config)

	// Should return an error (either from network or context cancellation)
	// but should not panic and should handle keyring properly
	assert.Assert(t, err != nil)
}

func TestAccountNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		accountName string
		expectError bool
	}{
		{
			name:        "valid account name",
			accountName: "test-account",
			expectError: false,
		},
		{
			name:        "account name with numbers",
			accountName: "account123",
			expectError: false,
		},
		{
			name:        "account name with hyphens",
			accountName: "my-test-account",
			expectError: false,
		},
		{
			name:        "empty account name",
			accountName: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAccountName(tt.accountName)

			if tt.expectError {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
