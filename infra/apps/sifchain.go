package apps

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
)

// NewSifchain creates new sifchain app
func NewSifchain(name, ip string) *Sifchain {
	return &Sifchain{
		name: name,
		ip:   ip,
	}
}

// Sifchain represents sifchain
type Sifchain struct {
	name string
	ip   string
}

// ID returns chain ID
func (s *Sifchain) ID() string {
	return s.name
}

// IP returns IP chain listens on
func (s *Sifchain) IP() string {
	return s.ip
}

// Deploy deploys sifchain app to the target
func (s *Sifchain) Deploy(ctx context.Context, config infra.Config, target infra.Target) error {
	sifchainHome := config.HomeDir + "/" + s.name
	sifnoded := func(args ...string) *osexec.Cmd {
		return osexec.Command("sifnoded", append([]string{"--home", sifchainHome}, args...)...)
	}
	sifnodedOut := func(buf *bytes.Buffer, args ...string) *osexec.Cmd {
		cmd := sifnoded(args...)
		cmd.Stdout = buf
		return cmd
	}

	if err := os.RemoveAll(sifchainHome); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.MkdirAll(sifchainHome, 0o700))

	keyName := s.name
	keyData := &bytes.Buffer{}
	accountAddrBuf := &bytes.Buffer{}
	accountAddrBechBuf := &bytes.Buffer{}
	err := exec.Run(ctx,
		sifnodedOut(keyData, "keys", "add", keyName, "--output", "json", "--keyring-backend", "test"),
		sifnodedOut(accountAddrBuf, "keys", "show", keyName, "-a", "--keyring-backend", "test"),
		sifnodedOut(accountAddrBechBuf, "keys", "show", keyName, "-a", "--bech", "val", "--keyring-backend", "test"),
	)
	if err != nil {
		return err
	}

	must.OK(ioutil.WriteFile(config.HomeDir+"/"+s.name+".json", keyData.Bytes(), 0o600))

	// FIXME (wojciech): create genesis file manually
	err = exec.Run(ctx,
		sifnoded("init", s.name, "--chain-id", s.name, "-o"),
		sifnoded("add-genesis-account", strings.TrimSuffix(accountAddrBuf.String(), "\n"), "500000000000000000000000rowan,990000000000000000000000000stake", "--keyring-backend", "test"),
		sifnoded("add-genesis-validators", strings.TrimSuffix(accountAddrBechBuf.String(), "\n"), "--keyring-backend", "test"),
		sifnoded("gentx", keyName, "1000000000000000000000000stake", "--chain-id", s.name, "--keyring-backend", "test"),
		sifnoded("collect-gentxs"),
	)
	if err != nil {
		return err
	}
	return target.DeployBinary(ctx, infra.Binary{
		Name: s.name,
		Path: "sifnoded",
		Args: []string{
			"start",
			"--home", sifchainHome,
			"--rpc.laddr", "tcp://" + s.ip + ":26657",
			"--p2p.laddr", "tcp://" + s.ip + ":26656",
			"--grpc.address", s.ip + ":9090",
			"--rpc.pprof_laddr", s.ip + ":6060",
		},
	})
}
