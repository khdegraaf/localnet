package main

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ridge/must"
	"github.com/spf13/cobra"
	"github.com/wojciech-sif/localnet"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/lib/run"
)

func main() {
	run.Tool("localnet", localnet.IoC, func(configF *localnet.ConfigFactory, cmdF *localnet.CmdFactory) error {
		rootCmd := &cobra.Command{
			RunE: cmdF.Cmd(func(ctx context.Context, configF *localnet.ConfigFactory) error {
				tty := exec.Tty()
				defer tty.Close()

				config := configF.Config()
				homeBin := config.HomeDir + "/bin"
				exeDir := filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))
				path := os.Getenv("PATH")
				if !strings.Contains(path, homeBin) {
					path += ":" + homeBin
				}
				if !strings.Contains(path, exeDir) {
					path += ":" + exeDir
				}

				bash := osexec.Command("bash")
				bash.Env = append(os.Environ(),
					fmt.Sprintf("PS1=%s", "("+configF.EnvName+") "+regexp.MustCompile(`^\(.*?\) *`).ReplaceAllString(os.Getenv("PS1"), "")),
					fmt.Sprintf("PATH=%s", path),
					fmt.Sprintf("LOCALNET_ENV=%s", configF.EnvName),
					fmt.Sprintf("LOCALNET_HOME=%s", configF.HomeDir),
					fmt.Sprintf("LOCALNET_TARGET=%s", configF.Target),
					fmt.Sprintf("LOCALNET_TMUX_NETWORK=%s", configF.TMuxNetwork),
				)
				bash.Stdin = tty
				bash.Stdout = tty
				bash.Stderr = tty
				return exec.Run(ctx, bash)
			}),
		}
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("LOCALNET_ENV", "localnet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("LOCALNET_HOME", must.String(os.UserHomeDir())+"/.localnet"), "Directory where all files created automatically by localnet are stored")
		rootCmd.PersistentFlags().StringVar(&configF.Target, "target", defaultString("LOCALNET_TARGET", "tmux"), "Target of the deployment (tmux | docker)")
		rootCmd.Flags().StringVar(&configF.TMuxNetwork, "tmux-network", defaultString("LOCALNET_TMUX_NETWORK", "127.1.0.0"), "Network where IPs for applications are taken from")

		startCmd := &cobra.Command{
			Use: "start",
			RunE: cmdF.Cmd(func(ctx context.Context, target infra.Target) error {
				return target.Deploy(ctx, localnet.DevEnv())
			}),
		}
		startCmd.Flags().StringVar(&configF.TMuxNetwork, "tmux-network", defaultString("LOCALNET_TMUX_NETWORK", "127.1.0.0"), "Network where IPs for applications are taken from")
		rootCmd.AddCommand(startCmd)

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
