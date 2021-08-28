package transfers

import (
	"context"
	"math/big"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/infra/apps/sifchain"
	"github.com/wojciech-sif/localnet/infra/testing"
)

// Successful is the transfer test which succeeds
func Successful(chain *apps.Sifchain) (testing.PrepareFunc, testing.RunFunc) {
	var wallet sifchain.Wallet

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			var err error

			// Create new random wallet with predefined balance added to genesis block
			wallet, err = chain.Genesis().AddWallet(ctx, sifchain.Balance{Denom: "rowan", Amount: big.NewInt(100)})

			return err
		},

		// Second function runs test
		func(ctx context.Context, t *testing.T) {
			// FIXME (wojciech): implement healthcheck loop
			time.Sleep(10 * time.Second)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Query for current balance available on the wallet
			balances, err := client.QBankBalances(ctx, wallet)
			require.NoError(t, err)

			// Test that wallet owns expected balance
			assert.Equal(t, "100", balances["rowan"].Amount.String())
		}
}
