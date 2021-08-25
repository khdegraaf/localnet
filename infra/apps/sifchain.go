package apps

import (
	"bytes"
	"context"
	"errors"
	"os"
	osexec "os/exec"
	"strings"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
)

var home string

func init() {
	home = must.String(os.UserHomeDir()) + "/.localnet"
}

// NewSifchain creates new sifchain app
func NewSifchain(name, ip string) *Sifchain {
	return &Sifchain{
		name: name,
		ip:   ip,
	}
}

// Sifchain represents
type Sifchain struct {
	name string
	ip   string
}

// Deploy deploys sifchain app to the target
func (s *Sifchain) Deploy(ctx context.Context, target infra.Target) error {
	home := home + "/" + s.name
	sifnoded := func(args ...string) *osexec.Cmd {
		return osexec.Command("sifnoded", append([]string{"--home", home}, args...)...)
	}
	sifnodedOut := func(buf *bytes.Buffer, args ...string) *osexec.Cmd {
		cmd := sifnoded(args...)
		cmd.Stdout = buf
		return cmd
	}

	if err := os.RemoveAll(home); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.MkdirAll(home, 0o700))

	keyName := s.name
	accountAddrBuf := &bytes.Buffer{}
	accountAddrBechBuf := &bytes.Buffer{}
	err := exec.Run(ctx,
		sifnoded("keys", "add", keyName, "--no-backup", "--keyring-backend", "test"),
		sifnodedOut(accountAddrBuf, "keys", "show", keyName, "-a", "--keyring-backend", "test"),
		sifnodedOut(accountAddrBechBuf, "keys", "show", keyName, "-a", "--bech", "val", "--keyring-backend", "test"),
	)
	if err != nil {
		return err
	}
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
			"--home", home,
			"--rpc.laddr", "tcp://" + s.ip + ":26657",
			"--p2p.laddr", "tcp://" + s.ip + ":26656",
			"--grpc.address", s.ip + ":9090",
			"--rpc.pprof_laddr", s.ip + ":6060",
		},
	})
}
