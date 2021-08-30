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
	"github.com/wojciech-sif/localnet/lib/logger"
	"go.uber.org/zap"
)

// VerifyInitialBalance checks that initial balance is set by genesis block
func VerifyInitialBalance(chain *apps.Sifchain) (testing.PrepareFunc, testing.RunFunc) {
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
			// Wait until chain is healthy
			testing.WaitUntilHealthy(ctx, t, 20*time.Second, chain)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Query for current balance available on the wallet
			balances, err := client.QBankBalances(ctx, wallet)
			require.NoError(t, err)

			// Test that wallet owns expected balance
			assert.Equal(t, "100", balances["rowan"].Amount.String())
		}
}

// TransferRowan checks that rowan is transferred correctly between wallets
func TransferRowan(chain *apps.Sifchain) (testing.PrepareFunc, testing.RunFunc) {
	var sender, receiver sifchain.Wallet

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			var err error

			// Create two random wallets with predefined amounts of rowans
			sender, err = chain.Genesis().AddWallet(ctx, sifchain.Balance{Denom: "rowan", Amount: big.NewInt(100)})
			if err != nil {
				return err
			}
			receiver, err = chain.Genesis().AddWallet(ctx, sifchain.Balance{Denom: "rowan", Amount: big.NewInt(10)})
			return err
		},

		// Second function runs test
		func(ctx context.Context, t *testing.T) {
			// Wait until chain is healthy
			testing.WaitUntilHealthy(ctx, t, 20*time.Second, chain)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Transfer 10 rowans from sender to receiver
			txHash, err := client.TxBankSend(ctx, sender, receiver, sifchain.Balance{Denom: "rowan", Amount: big.NewInt(10)})
			require.NoError(t, err)

			logger.Get(ctx).Info("Transfer executed", zap.String("txHash", txHash))

			// Query wallets for current balance
			balancesSender, err := client.QBankBalances(ctx, sender)
			require.NoError(t, err)

			balancesReceiver, err := client.QBankBalances(ctx, receiver)
			require.NoError(t, err)

			// Test that tokens disappeared from sender's wallet
			assert.Equal(t, "90", balancesSender["rowan"].Amount.String())

			// Test that tokens reached receiver's wallet
			assert.Equal(t, "20", balancesReceiver["rowan"].Amount.String())
		}
}
