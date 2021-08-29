package main

import (
	"os"
	"strings"

	"github.com/ridge/must"
	"github.com/spf13/cobra"
	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/lib/run"
)

func main() {
	run.Tool("localnet", localnet.IoC, func(c *ioc.Container, configF *localnet.ConfigFactory, cmdF *localnet.CmdFactory) error {
		rootCmd := &cobra.Command{
			Short: "Creates preconfigured bash session for environment",
			RunE:  cmdF.Cmd(localnet.Activate),
		}
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("LOCALNET_ENV", "localnet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("LOCALNET_HOME", must.String(os.UserHomeDir())+"/.localnet"), "Directory where all files created automatically by localnet are stored")
		rootCmd.PersistentFlags().StringVar(&configF.Target, "target", defaultString("LOCALNET_TARGET", "tmux"), "Target of the deployment: "+strings.Join(c.Names((*infra.Target)(nil)), " | "))
		rootCmd.PersistentFlags().StringVar(&configF.BinDir, "bin-dir", defaultString("LOCALNET_BIN_DIR", must.String(os.UserHomeDir())+"/go/bin"), "Path to directory where executables exist")
		rootCmd.PersistentFlags().StringVar(&configF.TMuxNetwork, "tmux-network", defaultString("LOCALNET_TMUX_NETWORK", "127.1.0.0"), "Network where IPs for applications are taken from")
		rootCmd.PersistentFlags().BoolVarP(&configF.VerboseLogging, "verbose", "v", defaultBool("LOCALNET_VERBOSE", false), "Turns on verbose logging")
		rootCmd.Flags().StringVar(&configF.SetName, "set", defaultString("LOCALNET_SET", "dev"), "Application set to deploy: "+strings.Join(c.Names((*infra.Set)(nil)), " | "))

		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Starts dev environment",
			RunE:  cmdF.Cmd(localnet.Start),
		}
		startCmd.Flags().StringVar(&configF.SetName, "set", defaultString("LOCALNET_SET", "dev"), "Application set to deploy: "+strings.Join(c.Names((*infra.Set)(nil)), " | "))
		rootCmd.AddCommand(startCmd)

		rootCmd.AddCommand(&cobra.Command{
			Use:   "test",
			Short: "Runs integration tests",
			RunE:  cmdF.Cmd(localnet.Test),
		})

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

func defaultBool(env string, def bool) bool {
	switch os.Getenv(env) {
	case "1", "true", "True", "TRUE":
		return true
	case "0", "false", "False", "FALSE":
		return false
	default:
		return def
	}
}
