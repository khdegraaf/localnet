package localnet

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ridge/must"
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

	homeRoot := filepath.Dir(config.HomeDir)
	exeDir := filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))

	var path string
	for _, p := range strings.Split(os.Getenv("PATH"), ":") {
		if !strings.HasPrefix(p, homeRoot) {
			if path != "" {
				path += ":"
			}
			path += p
		}
	}
	path = config.WrapperDir + ":" + path
	if !strings.Contains(path, exeDir) {
		path = exeDir + ":" + path
	}

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
		fmt.Sprintf("LOCALNET_VERBOSE=%t", configF.VerboseLogging),
	)
	bash.Dir = config.LogDir
	bash.Stdin = tty
	bash.Stdout = tty
	bash.Stderr = tty
	return exec.Run(ctx, bash)
}

// Start starts dev environment
func Start(ctx context.Context, config infra.Config, target infra.Target, set infra.Set, spec *infra.Spec) error {
	if err := target.Deploy(ctx, set); err != nil {
		return err
	}
	return saveSpec(config.HomeDir, spec)
}

// Test runs integration tests
func Test(ctx context.Context, config infra.Config, target infra.Target, appF *apps.Factory, spec *infra.Spec) error {
	env, tests := tests.Tests(appF)
	if err := testing.Run(ctx, target, env, tests); err != nil {
		return err
	}
	return saveSpec(config.HomeDir, spec)
}

// Spec print specification of running environment
func Spec(config infra.Config) error {
	spec, err := ioutil.ReadFile(config.HomeDir + "/spec.json")
	if err != nil {
		return err
	}
	fmt.Println(string(spec))
	return nil
}

func saveSpec(homeDir string, spec *infra.Spec) error {
	return ioutil.WriteFile(homeDir+"/spec.json", must.Bytes(json.MarshalIndent(spec, "", "  ")), 0o600)
}
