package apps

import (
	"context"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"time"

	"github.com/ridge/must"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps/hermes"
)

// NewHermes creates new hermes app
func NewHermes(config infra.Config, name string, spec *infra.Spec, chainA, chainB hermes.Peer) *Hermes {
	appDesc := spec.DescribeApp("hermes", name)
	appDesc.AddParam("chainA", chainA.Name())
	appDesc.AddParam("chainB", chainB.Name())
	return &Hermes{
		config:  config,
		appDesc: appDesc,
		name:    name,
		chainA:  chainA,
		chainB:  chainB,
	}
}

// Hermes represents hermes relayer
type Hermes struct {
	config  infra.Config
	appDesc *infra.AppDescription
	name    string
	chainA  hermes.Peer
	chainB  hermes.Peer
}

// Name returns name of app
func (h *Hermes) Name() string {
	return h.name
}

// Deploy deploys sifchain app to the target
func (h *Hermes) Deploy(ctx context.Context, target infra.AppTarget) error {
	waitCtx, waitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer waitCancel()
	if err := infra.WaitUntilHealthy(waitCtx, h.chainA); err != nil {
		return err
	}
	if err := infra.WaitUntilHealthy(waitCtx, h.chainB); err != nil {
		return err
	}

	bin := h.config.BinDir + "/hermes"

	hermesHome := h.config.AppDir + "/" + h.name
	configFile := hermesHome + "/config.toml"
	hermes := func(args ...string) *osexec.Cmd {
		return osexec.Command(bin, append([]string{"--config", configFile}, args...)...)
	}

	cfg := `[global]
strategy = 'packets'
filter = false
log_level = 'info'
clear_packets_interval = 100

[telemetry]
enabled = true
host = '{{ .IP }}'
port = 3001

[[chains]]
id = '` + h.chainA.ID() + `'
rpc_addr = 'http://` + h.chainA.IP().String() + `:26657'
grpc_addr = 'http://` + h.chainA.IP().String() + `:9090'
websocket_addr = 'ws://` + h.chainA.IP().String() + `:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'sif'
key_name = '` + h.config.EnvName + "-" + h.chainA.ID() + `'
store_prefix = 'ibc'
max_gas = 3000000
gas_price = { price = 0.001, denom = 'stake' }
gas_adjustment = 0.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = '` + h.chainB.ID() + `'
rpc_addr = 'http://` + h.chainB.IP().String() + `:26657'
grpc_addr = 'http://` + h.chainB.IP().String() + `:9090'
websocket_addr = 'ws://` + h.chainB.IP().String() + `:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'sif'
key_name = '` + h.config.EnvName + "-" + h.chainB.ID() + `'
store_prefix = 'ibc'
max_gas = 3000000
gas_price = { price = 0.001, denom = 'stake' }
gas_adjustment = 0.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
`
	if err := os.RemoveAll(hermesHome); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
	must.OK(os.MkdirAll(hermesHome, 0o700))

	deployment, err := target.DeployBinary(ctx, infra.Binary{
		Path: bin,
		AppBase: infra.AppBase{
			Name: h.name,
			Args: []string{
				"--config", configFile,
				"start",
			},
			Files: []infra.File{
				{
					Path:       configFile,
					Content:    []byte(cfg),
					Preprocess: true,
				},
			},
			Copy: []string{
				bin,
				hermesHome,
			},
			PreFunc: func(ctx context.Context) error {
				return exec.Run(ctx,
					hermes("keys", "add", h.chainA.ID(), "--file", h.config.AppDir+"/"+h.chainA.ID()+"/data/master.json"),
					hermes("keys", "add", h.chainB.ID(), "--file", h.config.AppDir+"/"+h.chainB.ID()+"/data/master.json"),
					hermes("create", "channel", h.chainA.ID(), h.chainB.ID(), "--port-a", "transfer", "--port-b", "transfer"),
				)
			},
		},
	})
	if err != nil {
		return err
	}

	h.appDesc.AddEndpoint("telemetry", fmt.Sprintf("%s:3001", deployment.IP))

	return nil
}
