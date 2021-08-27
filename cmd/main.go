package main

import (
	"context"
	"os"

	"github.com/ridge/must"
	"github.com/spf13/cobra"
	"github.com/wojciech-sif/localnet"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/lib/run"
)

func main() {
	run.Tool("localnet", localnet.IoC, func(configF *localnet.ConfigFactory, cmdF *localnet.CmdFactory) error {
		rootCmd := &cobra.Command{
			RunE: cmdF.Cmd(func(ctx context.Context, target infra.Target) error {
				return target.Deploy(ctx, localnet.DevEnv())
			}),
		}
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("LOCALNET_ENV", "localnet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("LOCALNET_HOME", must.String(os.UserHomeDir())+"/.localnet"), "Directory where all files created automatically by localnet are stored")
		rootCmd.PersistentFlags().StringVar(&configF.Target, "target", defaultString("LOCALNET_TARGET", "tmux"), "Target of the deployment (tmux | docker)")
		rootCmd.Flags().StringVar(&configF.TMuxNetwork, "tmux-network", defaultString("LOCALNET_TMUX_NETWORK", "127.1.0.0"), "Network where IPs for applications are taken from")
		return rootCmd.Execute()
	})
}

func defaultString(env, def string) string {
	val := os.Getenv(env)
	if val == "" {
		val = def
	}
	return val
}
