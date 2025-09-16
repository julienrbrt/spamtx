package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Chain:   "cosmoshub",
				Account: "cosmos1abc123",
				Fees:    "1000uatom",
				Memo:    "test memo",
				TPS:     10,
			},
			wantErr: false,
		},
		{
			name: "valid config with custom RPC",
			config: Config{
				Chain:   "cosmoshub",
				Account: "cosmos1abc123",
				Fees:    "1000uatom",
				Memo:    "test memo",
				TPS:     10,
				RPC:     "http://localhost:26657",
			},
			wantErr: false,
		},
		{
			name: "empty chain",
			config: Config{
				Chain:   "",
				Account: "cosmos1abc123",
				Fees:    "1000uatom",
				Memo:    "test memo",
				TPS:     10,
			},
			wantErr: true,
		},
		{
			name: "empty account",
			config: Config{
				Chain:   "cosmoshub",
				Account: "",
				Fees:    "1000uatom",
				Memo:    "test memo",
				TPS:     10,
			},
			wantErr: true,
		},
		{
			name: "empty fees",
			config: Config{
				Chain:   "cosmoshub",
				Account: "cosmos1abc123",
				Fees:    "",
				Memo:    "test memo",
				TPS:     10,
			},
			wantErr: true,
		},
		{
			name: "empty memo",
			config: Config{
				Chain:   "cosmoshub",
				Account: "cosmos1abc123",
				Fees:    "1000uatom",
				Memo:    "",
				TPS:     10,
			},
			wantErr: true,
		},
		{
			name: "zero tps",
			config: Config{
				Chain:   "cosmoshub",
				Account: "cosmos1abc123",
				Fees:    "1000uatom",
				Memo:    "test memo",
				TPS:     0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
