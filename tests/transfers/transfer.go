package transfers

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/require"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/infra/apps/sifchain"
	"github.com/wojciech-sif/localnet/infra/testing"
	"github.com/wojciech-sif/localnet/lib/logger"
)

// Successful is the transfer test which succeeds
func Successful(chain *apps.Sifchain) (testing.PrepareFunc, testing.RunFunc) {
	var wallet sifchain.Wallet
	return func(ctx context.Context) error {
			var err error
			wallet, err = chain.Genesis().AddWallet(ctx, sifchain.Balance{Denom: "rowan", Amount: big.NewInt(100)})
			return err
		}, func(ctx context.Context, t *testing.T) {
			logger.Get(ctx).Info(wallet.Address)
		}
}

// Failing is the transfer test which fails
func Failing(chain *apps.Sifchain) (testing.PrepareFunc, testing.RunFunc) {
	var wallet sifchain.Wallet
	return func(ctx context.Context) error {
			var err error
			wallet, err = chain.Genesis().AddWallet(ctx, sifchain.Balance{Denom: "rowan", Amount: big.NewInt(100)})
			return err
		}, func(ctx context.Context, t *testing.T) {
			logger.Get(ctx).Info(wallet.Address)
			require.Equal(t, 1, 2)
		}
}
