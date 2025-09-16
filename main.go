package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n🛑 Received interrupt signal. Shutting down gracefully...")
		cancel()
	}()

	_ = fang.Execute(ctx, rootCmd()) // errors are handled by command
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spamtx",
		Short: "Spam txs to a Cosmos SDK based blockchain",
		Long:  "A tool that performs self bank sends with memo fields to spam transactions at a controlled rate",
	}

	// Add subcommands
	cmd.AddCommand(spamCmd())
	cmd.AddCommand(keyringCmd())

	// Hide the completion command
	cmd.CompletionOptions.HiddenDefaultCmd = true

	return cmd
}

func spamCmd() *cobra.Command {
	var config Config

	cmd := &cobra.Command{
		Use:   "spam [chain]",
		Args:  cobra.ExactArgs(1),
		Short: "Start spamming transactions",
		Long:  "Start spamming self bank send transactions at a controlled rate",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Chain = args[0]
			if err := validateConfig(config); err != nil {
				return err
			}

			return spamTransactions(cmd.Context(), config)
		},
	}

	cmd.Flags().StringVar(&config.Account, flagFrom, "", "Account name from keyring")
	cmd.Flags().StringVar(&config.Fees, flagFees, "", "Transaction fees")
	cmd.Flags().StringVar(&config.Memo, flagMemo, "", "Transaction memo")
	cmd.Flags().IntVar(&config.TPS, flagTPS, 10, "Transactions per second")

	_ = cmd.MarkFlagRequired(flagFrom)
	_ = cmd.MarkFlagRequired(flagFees)
	_ = cmd.MarkFlagRequired(flagMemo)

	return cmd
}

func keyringCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keyring",
		Short: "Manage keyring accounts",
		Long:  "Create, list, import, and delete accounts in the keyring",
	}

	cmd.AddCommand(keyringCreateCmd())
	cmd.AddCommand(keyringListCmd())
	cmd.AddCommand(keyringImportCmd())
	cmd.AddCommand(keyringDeleteCmd())

	return cmd
}

func keyringCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [chain] [account-name]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a new account in the keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainName := args[0]
			accountName := args[1]

			registry, _, err := initializeKeyring(chainName)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			_, _, err = getOrCreateAccount(registry, accountName)
			return err
		},
	}
}

func keyringListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [chain]",
		Args:  cobra.ExactArgs(1),
		Short: "List all accounts in the keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainName := args[0]

			registry, bech32Prefix, err := initializeKeyring(chainName)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			return listAccounts(registry, bech32Prefix)
		},
	}
}

func keyringImportCmd() *cobra.Command {
	var passphrase string

	cmd := &cobra.Command{
		Use:   "import [chain] [account-name] [mnemonic-or-key]",
		Args:  cobra.ExactArgs(3),
		Short: "Import an account from mnemonic or private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainName := args[0]
			accountName := args[1]
			secret := args[2]

			registry, bech32Prefix, err := initializeKeyring(chainName)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			return importAccount(registry, accountName, secret, passphrase, bech32Prefix)
		},
	}

	cmd.Flags().StringVar(&passphrase, "passphrase", "", "Passphrase for encrypted private key")

	return cmd
}

func keyringDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [chain] [account-name]",
		Args:  cobra.ExactArgs(2),
		Short: "Delete an account from the keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainName := args[0]
			accountName := args[1]

			registry, _, err := initializeKeyring(chainName)
			if err != nil {
				return fmt.Errorf("failed to initialize keyring: %w", err)
			}

			return deleteAccount(registry, accountName)
		},
	}
}
