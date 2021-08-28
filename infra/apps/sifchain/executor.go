package sifchain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
)

// NewExecutor returns new executor
func NewExecutor(name, binPath, homeDir, keyName string) *Executor {
	tmpDir := homeDir + "/tmp"
	if err := os.RemoveAll(tmpDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.MkdirAll(tmpDir, 0o700))
	return &Executor{
		name:    name,
		binPath: binPath,
		tmpDir:  tmpDir,
		dataDir: homeDir + "/data",
		keyName: keyName,
	}
}

// Executor exposes methods for executing sifnoded binary
type Executor struct {
	name    string
	binPath string
	tmpDir  string
	dataDir string
	keyName string
}

// Name returns name of the chain
func (e *Executor) Name() string {
	return e.name
}

// Bin returns path to sifnode binary
func (e *Executor) Bin() string {
	return e.binPath
}

// Home returns path to home dir
func (e *Executor) Home() string {
	return e.dataDir
}

// AddKey adds key to the client
func (e *Executor) AddKey(ctx context.Context, name string) (addr, validatorAddr string, err error) {
	keyData := &bytes.Buffer{}
	addrBuf := &bytes.Buffer{}
	validatorAddrBuf := &bytes.Buffer{}
	err = exec.Run(ctx,
		e.sifnodedOut(keyData, e.tmpDir, "keys", "add", name, "--output", "json", "--keyring-backend", "test"),
		e.sifnodedOut(addrBuf, e.tmpDir, "keys", "show", name, "-a", "--keyring-backend", "test"),
		e.sifnodedOut(validatorAddrBuf, e.tmpDir, "keys", "show", name, "-a", "--bech", "val", "--keyring-backend", "test"),
	)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSuffix(addrBuf.String(), "\n"), strings.TrimSuffix(validatorAddrBuf.String(), "\n"), ioutil.WriteFile(e.tmpDir+"/"+name+".json", keyData.Bytes(), 0o600)
}

// PrepareNode prepares node to start
func (e *Executor) PrepareNode(ctx context.Context, genesis *Genesis) error {
	addr, valAddr, err := e.AddKey(ctx, e.keyName)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(e.dataDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.Rename(e.tmpDir, e.dataDir))

	cmds := []*osexec.Cmd{
		e.sifnoded(e.dataDir, "init", e.name, "--chain-id", e.name, "-o"),
		e.sifnoded(e.dataDir, "add-genesis-account", addr, "500000000000000000000000rowan,990000000000000000000000000stake", "--keyring-backend", "test"),
		e.sifnoded(e.dataDir, "add-genesis-validators", valAddr, "--keyring-backend", "test"),
	}
	for wallet, balances := range genesis.wallets {
		if len(balances) == 0 {
			continue
		}
		balancesStr := ""
		for _, balance := range balances {
			if balancesStr != "" {
				balancesStr += ","
			}
			balancesStr += balance.Amount.String() + balance.Denom
		}
		cmds = append(cmds, e.sifnoded(e.dataDir, "add-genesis-account", wallet.Address, balancesStr, "--keyring-backend", "test"))
	}
	cmds = append(cmds,
		e.sifnoded(e.dataDir, "gentx", e.keyName, "1000000000000000000000000stake", "--chain-id", e.name, "--keyring-backend", "test"),
		e.sifnoded(e.dataDir, "collect-gentxs"),
	)
	return exec.Run(ctx, cmds...)
}

// QBankBalances queries for bank balances owned by address
func (e *Executor) QBankBalances(ctx context.Context, address string, ip net.IP) ([]byte, error) {
	balances := &bytes.Buffer{}
	if err := exec.Run(ctx, e.sifnodedOut(balances, e.dataDir, "q", "bank", "balances", address, "--chain-id", e.name, "--node", fmt.Sprintf("tcp://%s:26657", ip), "--output", "json")); err != nil {
		return nil, err
	}
	return balances.Bytes(), nil
}

// TxBankSend sends tokens from one address to another
func (e *Executor) TxBankSend(ctx context.Context, sender, address string, balance Balance, ip net.IP) ([]byte, error) {
	tx := &bytes.Buffer{}
	if err := exec.Run(ctx, e.sifnodedOut(tx, e.dataDir, "tx", "bank", "send", sender, address, balance.Amount.String()+balance.Denom, "--yes", "--chain-id", e.name, "--node", fmt.Sprintf("tcp://%s:26657", ip), "--keyring-backend", "test", "--output", "json")); err != nil {
		return nil, err
	}
	return tx.Bytes(), nil
}

func (e *Executor) sifnoded(homeDir string, args ...string) *osexec.Cmd {
	return osexec.Command(e.binPath, append([]string{"--home", homeDir}, args...)...)
}

func (e *Executor) sifnodedOut(buf *bytes.Buffer, homeDir string, args ...string) *osexec.Cmd {
	cmd := e.sifnoded(homeDir, args...)
	cmd.Stdout = buf
	return cmd
}
