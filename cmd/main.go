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
		rootCmd.PersistentFlags().BoolVarP(&configF.VerboseLogging, "verbose", "v", defaultBool("LOCALNET_VERBOSE", false), "Turns on verbose logging")
		addFlags(rootCmd, c, configF)
		addSetFlag(rootCmd, c, configF)

		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Starts dev environment",
			RunE:  cmdF.Cmd(localnet.Start),
		}
		addFlags(startCmd, c, configF)
		addSetFlag(startCmd, c, configF)
		rootCmd.AddCommand(startCmd)

		testCmd := &cobra.Command{
			Use:   "test",
			Short: "Runs integration tests",
			RunE:  cmdF.Cmd(localnet.Test),
		}
		addFlags(testCmd, c, configF)
		rootCmd.AddCommand(testCmd)

		rootCmd.AddCommand(&cobra.Command{
			Use:   "spec",
			Short: "Prints specification of running environment",
			RunE:  cmdF.Cmd(localnet.Spec),
		})

		return rootCmd.Execute()
	})
}

func addFlags(cmd *cobra.Command, c *ioc.Container, configF *localnet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.Target, "target", defaultString("LOCALNET_TARGET", "tmux"), "Target of the deployment: "+strings.Join(c.Names((*infra.Target)(nil)), " | "))
	cmd.Flags().StringVar(&configF.BinDir, "bin-dir", defaultString("LOCALNET_BIN_DIR", must.String(os.UserHomeDir())+"/go/bin"), "Path to directory where executables exist")
	cmd.Flags().StringVar(&configF.Network, "network", defaultString("LOCALNET_NETWORK", "127.1.0.0"), "Network where IPs for applications are taken from (related to 'tmux' and 'direct' targets only)")
}

func addSetFlag(cmd *cobra.Command, c *ioc.Container, configF *localnet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.SetName, "set", defaultString("LOCALNET_SET", "dev"), "Application set to deploy: "+strings.Join(c.Names((*infra.Set)(nil)), " | "))
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
