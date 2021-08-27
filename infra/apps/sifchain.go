package apps

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net"
	"os"
	osexec "os/exec"
	"strings"
	"sync"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
)

// NewSifchain creates new sifchain app
func NewSifchain(name string) *Sifchain {
	return &Sifchain{
		name: name,
	}
}

// Sifchain represents sifchain
type Sifchain struct {
	name string

	mu sync.RWMutex
	ip net.IP
}

// ID returns chain ID
func (s *Sifchain) ID() string {
	return s.name
}

// IP returns IP chain listens on
func (s *Sifchain) IP() net.IP {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ip
}

// Deploy deploys sifchain app to the target
func (s *Sifchain) Deploy(ctx context.Context, config infra.Config, target infra.AppTarget) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bin := config.BinDir + "/sifnoded"

	sifchainHome := config.HomeDir + "/" + s.name
	sifnoded := func(args ...string) *osexec.Cmd {
		return osexec.Command(bin, append([]string{"--home", sifchainHome}, args...)...)
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

	deployment, err := target.DeployBinary(ctx, infra.Binary{
		Path: bin,
		AppBase: infra.AppBase{
			Name: s.name,
			Args: []string{
				"start",
				"--home", sifchainHome,
				"--rpc.laddr", "tcp://{{ .IP }}:26657",
				"--p2p.laddr", "tcp://{{ .IP }}:26656",
				"--grpc.address", "{{ .IP }}:9090",
				"--rpc.pprof_laddr", "{{ .IP }}:6060",
			},
			Copy: []string{
				bin,
				sifchainHome,
			},
			PreFunc: func(ctx context.Context) error {
				keyName := "master"
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
				return exec.Run(ctx,
					sifnoded("init", s.name, "--chain-id", s.name, "-o"),
					sifnoded("add-genesis-account", strings.TrimSuffix(accountAddrBuf.String(), "\n"), "500000000000000000000000rowan,990000000000000000000000000stake", "--keyring-backend", "test"),
					sifnoded("add-genesis-validators", strings.TrimSuffix(accountAddrBechBuf.String(), "\n"), "--keyring-backend", "test"),
					sifnoded("gentx", keyName, "1000000000000000000000000stake", "--chain-id", s.name, "--keyring-backend", "test"),
					sifnoded("collect-gentxs"),
				)
			},
		},
	})
	if err != nil {
		return err
	}
	s.ip = deployment.IP

	client := `#!/bin/sh
OPTS=""
if [ "$1" == "tx" ] || [ "$1" == "q" ]; then
	OPTS="$OPTS --chain-id ""` + s.name + `"" --node ""tcp://` + s.ip.String() + `:26657"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec ` + bin + ` --home "` + sifchainHome + `" "$@" $OPTS
`
	if err := os.MkdirAll(config.HomeDir+"/bin", 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	must.OK(ioutil.WriteFile(config.HomeDir+"/bin/"+s.name, []byte(client), 0o700))
	return nil
}
