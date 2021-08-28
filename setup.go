package localnet

import (
	"errors"
	"net"
	"os"
	"path/filepath"

	"github.com/ridge/must"
	"github.com/spf13/cobra"
	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/infra/targets"
)

// IoC configures IoC container
func IoC(c *ioc.Container) {
	c.Singleton(NewCmdFactory)
	c.Singleton(func() *ConfigFactory {
		cf := NewConfigFactory()
		cf.TestingMode = len(os.Args) > 1 && os.Args[1] == "test"
		return cf
	})
	c.Transient(func(configF *ConfigFactory) infra.Config {
		return configF.Config()
	})
	c.Transient(apps.NewFactory)
	c.TransientNamed("dev", DevSet)
	c.TransientNamed("single", SingleChainSet)
	c.Transient(func(c *ioc.Container, config infra.Config) infra.Set {
		var set infra.Set
		c.ResolveNamed(config.SetName, &set)
		return set
	})
	c.TransientNamed("tmux", targets.NewTMux)
	c.TransientNamed("docker", targets.NewDocker)
	c.Transient(func(c *ioc.Container, config infra.Config) infra.Target {
		var target infra.Target
		c.ResolveNamed(config.Target, &target)
		return target
	})
}

// NewCmdFactory returns new CmdFactory
func NewCmdFactory(c *ioc.Container) *CmdFactory {
	return &CmdFactory{
		c: c,
	}
}

// CmdFactory is a wrapper around cobra RunE
type CmdFactory struct {
	c *ioc.Container
}

// Cmd returns function compatible with RunE
func (f *CmdFactory) Cmd(cmdFunc interface{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		f.c.Call(cmdFunc, &err)
		return err
	}
}

// NewConfigFactory creates new ConfigFactory
func NewConfigFactory() *ConfigFactory {
	return &ConfigFactory{}
}

// ConfigFactory collects config from CLI and produces real config
type ConfigFactory struct {
	// EnvName is the name of created environment
	EnvName string

	// SetName is the name of set
	SetName string

	// Target is the deployment target
	Target string

	// HomeDir is the path where all the files are kept
	HomeDir string

	// BinDir is the path where all binaries are present
	BinDir string

	// TMuxNetwork is the IP network for processes executed directly in tmux
	TMuxNetwork string

	// TestingMode means we are in testing mode and deployment should not block execution
	TestingMode bool
}

// Config produces final config
func (cf *ConfigFactory) Config() infra.Config {
	if err := os.MkdirAll(cf.HomeDir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	homeDir := must.String(filepath.Abs(must.String(filepath.EvalSymlinks(cf.HomeDir)))) + "/" + cf.EnvName
	if err := os.Mkdir(homeDir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	return infra.Config{
		EnvName:     cf.EnvName,
		SetName:     cf.SetName,
		Target:      cf.Target,
		HomeDir:     homeDir,
		BinDir:      must.String(filepath.Abs(must.String(filepath.EvalSymlinks(cf.BinDir)))),
		TMuxNetwork: net.ParseIP(cf.TMuxNetwork),
		TestingMode: cf.TestingMode,
	}
}
