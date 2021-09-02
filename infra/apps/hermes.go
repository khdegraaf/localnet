package apps

import (
	"context"
	"fmt"
	osexec "os/exec"
	"time"

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
	bin := h.config.BinDir + "/hermes"
	hermesHome := h.config.AppDir + "/" + h.name
	configFile := hermesHome + "/config.toml"
	hermes := func(args ...string) *osexec.Cmd {
		return osexec.Command(bin, append([]string{"--config", configFile}, args...)...)
	}

	return target.DeployBinary(ctx, infra.Binary{
		Path:       bin,
		RequiresIP: true,
		AppBase: infra.AppBase{
			Name: h.name,
			Args: []string{
				"--config", configFile,
				"start",
			},
			Files: []infra.File{
				{
					Path:        configFile,
					ContentFunc: h.generateConfig,
					Preprocess:  true,
				},
			},
			Copy: []string{
				bin,
				hermesHome,
			},
			Requires: infra.Prerequisites{
				Timeout: 10 * time.Second,
				Dependencies: []infra.HealthCheckCapable{
					h.chainA,
					h.chainB,
				},
			},
			PreFunc: func(ctx context.Context) error {
				return exec.Run(ctx,
					hermes("keys", "add", h.chainA.ID(), "--file", h.config.AppDir+"/"+h.chainA.ID()+"/master.json"),
					hermes("keys", "add", h.chainB.ID(), "--file", h.config.AppDir+"/"+h.chainB.ID()+"/master.json"),
					hermes("create", "channel", h.chainA.ID(), h.chainB.ID(), "--port-a", "transfer", "--port-b", "transfer"),
				)
			},
			PostFunc: func(ctx context.Context, deployment infra.Deployment) error {
				h.appDesc.IP = deployment.IP
				h.appDesc.AddEndpoint("telemetry", fmt.Sprintf("%s:3001", deployment.IP))
				return nil
			},
		},
	})
}

func (h *Hermes) generateConfig() []byte {
	return []byte(`[global]
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
`)
}
