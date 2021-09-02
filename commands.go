package localnet

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ridge/must"
	"github.com/wojciech-malota-wojcik/ioc"
	"github.com/wojciech-sif/localnet/exec"
	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/infra/apps"
	"github.com/wojciech-sif/localnet/infra/testing"
	"github.com/wojciech-sif/localnet/tests"
)

// Activate starts preconfigured bash environment
func Activate(ctx context.Context, configF *ConfigFactory) error {
	tty := exec.Tty()
	defer tty.Close()

	config := configF.Config()

	exe := must.String(filepath.EvalSymlinks(must.String(os.Executable())))
	var path string
	for _, p := range strings.Split(os.Getenv("PATH"), ":") {
		if !strings.HasPrefix(p, configF.HomeDir) {
			if path != "" {
				path += ":"
			}
			path += p
		}
	}
	path = config.WrapperDir + ":" + path

	must.OK(ioutil.WriteFile(config.WrapperDir+"/l", []byte(fmt.Sprintf("#!/bin/bash\nexec %s \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/start", []byte(fmt.Sprintf("#!/bin/bash\nexec %s start \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/stop", []byte(fmt.Sprintf("#!/bin/bash\nexec %s stop \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/tests", []byte(fmt.Sprintf("#!/bin/bash\nexec %s tests \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/spec", []byte(fmt.Sprintf("#!/bin/bash\nexec %s spec \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/logs", []byte(fmt.Sprintf("#!/bin/bash\nexec tail -f -n +0 \"%s/$1.log\"", config.LogDir)), 0o700))

	bash := osexec.Command("bash")
	bash.Env = append(os.Environ(),
		fmt.Sprintf("PS1=%s", "("+configF.EnvName+") "+regexp.MustCompile(`^\(.*?\) *`).ReplaceAllString(os.Getenv("PS1"), "")),
		fmt.Sprintf("PATH=%s", path),
		fmt.Sprintf("LOCALNET_ENV=%s", configF.EnvName),
		fmt.Sprintf("LOCALNET_SET=%s", configF.SetName),
		fmt.Sprintf("LOCALNET_HOME=%s", configF.HomeDir),
		fmt.Sprintf("LOCALNET_TARGET=%s", configF.Target),
		fmt.Sprintf("LOCALNET_BIN_DIR=%s", configF.BinDir),
		fmt.Sprintf("LOCALNET_NETWORK=%s", configF.Network),
		fmt.Sprintf("LOCALNET_FILTERS=%s", strings.Join(configF.TestFilters, ",")),
		fmt.Sprintf("LOCALNET_VERBOSE=%t", configF.VerboseLogging),
	)
	bash.Dir = config.LogDir
	bash.Stdin = tty
	bash.Stdout = tty
	bash.Stderr = tty
	return exec.Run(ctx, bash)
}

// Start starts environment
func Start(ctx context.Context, target infra.Target, set infra.Set, spec *infra.Spec) (retErr error) {
	defer func() {
		if err := spec.Save(); retErr == nil {
			retErr = err
		}
	}()
	return target.Deploy(ctx, set)
}

// Stop stops environment
func Stop(ctx context.Context, target infra.Target, _ infra.Set, spec *infra.Spec) (retErr error) {
	defer func() {
		if err := spec.Reset(); retErr == nil {
			retErr = err
		}
	}()
	return target.Stop(ctx)
}

// Tests runs integration tests
func Tests(c *ioc.Container, configF *ConfigFactory) error {
	configF.TestingMode = true
	configF.SetName = "tests"
	var err error
	c.Call(func(ctx context.Context, config infra.Config, target infra.Target, appF *apps.Factory, spec *infra.Spec) (retErr error) {
		defer func() {
			if err := spec.Save(); retErr == nil {
				retErr = err
			}
		}()

		env, tests := tests.Tests(appF)
		return testing.Run(ctx, target, env, tests, config.TestFilters)
	}, &err)
	return err
}

// Spec print specification of running environment
func Spec(spec *infra.Spec, _ infra.Set) error {
	fmt.Println(spec)
	return nil
}
