package main

import (
	"errors"
)

var (
	flagFrom = "from"
	flagFees = "fees"
	flagMemo = "memo"
	flagTPS  = "tps"
)

// Config holds the command line configuration
type Config struct {
	Chain   string
	Account string
	Fees    string
	Memo    string
	TPS     int
}

// validateConfig validates the configuration parameters
func validateConfig(config Config) error {
	if config.Chain == "" {
		return errors.New("chain name is required")
	}
	if config.Account == "" {
		return errors.New("account address is required")
	}
	if config.Fees == "" {
		return errors.New("fees are required")
	}
	if config.Memo == "" {
		return errors.New("memo is required")
	}
	if config.TPS <= 0 {
		return errors.New("tps must be greater than 0")
	}

	return nil
}
