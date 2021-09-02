package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ridge/must"
)

// App is the interface exposed by application
type App interface {
	// Name returns name of application
	Name() string

	// Deploy deploys app to the target
	Deploy(ctx context.Context, target AppTarget) error
}

// Set is the environment to deploy
type Set []App

// Deploy deploys app in environment to the target
func (s Set) Deploy(ctx context.Context, t AppTarget, spec *Spec) error {
	for _, app := range s {
		if appSpec, exists := spec.Apps[app.Name()]; exists && appSpec.Running {
			continue
		}
		if err := app.Deploy(ctx, t); err != nil {
			return err
		}
		spec.Apps[app.Name()].Running = true
	}
	return nil
}

// Deployment contains info about deployed application
type Deployment struct {
	// IP is the IP address assigned to application
	IP net.IP
}

// Target represents target of deployment from the perspective of localnet
type Target interface {
	// Deploy deploys environment to the target
	Deploy(ctx context.Context, env Set) error

	// Stop stops apps in the environment
	Stop(ctx context.Context) error
}

// AppTarget represents target of deployment from the perspective of application
type AppTarget interface {
	// DeployBinary deploys binary to the target
	DeployBinary(ctx context.Context, app Binary) error

	// DeployContainer deploys container to the target
	DeployContainer(ctx context.Context, app Container) error
}

// File represents file to be created for application
type File struct {
	// Path is the path to file
	Path string

	// Content represents content of file. Takes precedence over `ContentFunc`
	Content []byte

	// ContentFunc is a function returning content of the file.
	// It is called if `Content` is empty.
	// It is called before application's `PreFunc` but after verifying that all the apps specified in `Requires` are healthy.
	ContentFunc func() []byte

	// Preprocess tells if file should be preprocessed using data delivered by target
	Preprocess bool
}

// PreprocessFunc is the function called to preprocess app
type PreprocessFunc func(ctx context.Context) error

// PostprocessFunc is the function called after application is deployed
type PostprocessFunc func(ctx context.Context, deployment Deployment) error

// Prerequisites specifies list of other apps which have to be healthy because app may be started.
type Prerequisites struct {
	// Timeout tells how long we should wait for prerequisite to become healthy
	Timeout time.Duration

	// Dependencies specifies a list of health checks this app depends on
	Dependencies []HealthCheckCapable
}

// AppBase contain properties common to all types of app
type AppBase struct {
	// Name of the application
	Name string

	// Args are args passed to binary
	Args []string

	// Files are the files to be created for application
	Files []File

	// Copy lists all the files and dirs required by the application
	Copy []string

	// Requires is the list of health checks to be required before app can be deployed
	Requires Prerequisites

	// PreFunc is called to preprocess app
	PreFunc PreprocessFunc

	// PostFunc is called after app is deployed
	PostFunc PostprocessFunc
}

// Binary represents binary file to be deployed
type Binary struct {
	AppBase

	// Path is the path to binary file
	Path string

	// RequiresIP set to true means app requires an IP address
	RequiresIP bool
}

// Container represents container to be deployed
type Container struct {
	AppBase

	// Image is the url of the container image
	Image string

	// Tag is the tag of the image
	Tag string
}

// NewSpec returns new spec
func NewSpec(config Config) *Spec {
	specFile := config.HomeDir + "/spec.json"
	specRaw, err := ioutil.ReadFile(specFile)
	switch {
	case err == nil:
		spec := &Spec{
			specFile: specFile,
		}
		must.OK(json.Unmarshal(specRaw, spec))
		if spec.Target != config.Target {
			panic(fmt.Sprintf("target mismatch, spec: %s, config: %s", spec.Target, config.Target))
		}
		if spec.Env != config.EnvName {
			panic(fmt.Sprintf("env mismatch, spec: %s, config: %s", spec.Env, config.EnvName))
		}
		if spec.Set != config.SetName {
			panic(fmt.Sprintf("set mismatch, spec: %s, config: %s", spec.Set, config.SetName))
		}
		return spec
	case errors.Is(err, os.ErrNotExist):
	default:
		panic(err)
	}

	spec := &Spec{
		specFile: specFile,
		Target:   config.Target,
		Set:      config.SetName,
		Env:      config.EnvName,
		Apps:     map[string]*AppDescription{},
	}
	if config.Target == "direct" {
		spec.PGID = os.Getpid()
	}
	return spec
}

// Spec describes running environment
type Spec struct {
	specFile string

	// PGID stores process group ID used to run apps - used only by direct target
	PGID int `json:"pgid,omitempty"`

	// Target is the name of target being used to run apps
	Target string `json:"target"`

	// Set is the name of app set
	Set string `json:"set"`

	// Env is the name of env
	Env string `json:"env"`

	mu sync.Mutex

	// Apps is the description of running apps
	Apps map[string]*AppDescription `json:"apps"`
}

// DescribeApp adds description of running app
func (s *Spec) DescribeApp(appType string, name string) *AppDescription {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app, exists := s.Apps[name]; exists {
		if app.Type != appType {
			panic(fmt.Sprintf("app type doesn't match for application existing in spec: %s, expected: %s, got: %s", name, app.Type, appType))
		}
		return app
	}

	appDesc := &AppDescription{
		Type: appType,
	}
	s.Apps[name] = appDesc
	return appDesc
}

// String converts spec to json string
func (s *Spec) String() string {
	return string(must.Bytes(json.MarshalIndent(s, "", "  ")))
}

// Save saves spec into file
func (s *Spec) Save() error {
	return ioutil.WriteFile(s.specFile, []byte(s.String()), 0o600)
}

// Reset removes spec size
func (s *Spec) Reset() error {
	if err := os.Remove(s.specFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// AppDescription describes app running in environment
type AppDescription struct {
	// Type is the type of app
	Type string `json:"type"`

	// IP is the IP reserved for this application
	IP net.IP `json:"ip,omitempty"`

	// Running indicates if apps are running
	Running bool `json:"running"`

	mu sync.Mutex

	// Endpoints describe endpoints exposed by application
	Endpoints map[string]string `json:"endpoints,omitempty"`

	// Params is a space for any parameters declared by application
	Params map[string]string `json:"params,omitempty"`
}

// AddEndpoint adds endpoint to app description
func (a *AppDescription) AddEndpoint(name, endpoint string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Endpoints == nil {
		a.Endpoints = map[string]string{}
	}

	if value, exists := a.Endpoints[name]; exists && endpoint != value {
		panic(fmt.Sprintf("conflict with existing endpoint: %s, expected: %s, got: %s", name, value, endpoint))
	}
	a.Endpoints[name] = endpoint
}

// AddParam adds parameter to app description
func (a *AppDescription) AddParam(name, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Params == nil {
		a.Params = map[string]string{}
	}

	if oldValue, exists := a.Params[name]; exists && oldValue != value {
		panic(fmt.Sprintf("conflict with existing parameter: %s, expected: %s, got: %s", name, oldValue, value))
	}
	a.Params[name] = value
}
