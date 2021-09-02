package infra

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// App is the interface exposed by application
type App interface {
	Deploy(ctx context.Context, target AppTarget) error
}

// Set is the environment to deploy
type Set []App

// Deploy deploys app in environment to the target
func (s Set) Deploy(ctx context.Context, t AppTarget) error {
	for _, app := range s {
		if err := app.Deploy(ctx, t); err != nil {
			return err
		}
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
	return &Spec{
		Target: config.Target,
		Set:    config.SetName,
		Env:    config.EnvName,
		Apps:   map[string]*AppDescription{},
	}
}

// Spec describes running environment
type Spec struct {
	// Target is the name of target being used to run apps
	Target string `json:"target"`

	// Set is the name of app set
	Set string `json:"set"`

	// Env is the name of env
	Env string `json:"env"`

	// Running indicates if apps are running
	Running bool `json:"running"`

	mu sync.Mutex

	// Apps is the description of running apps
	Apps map[string]*AppDescription `json:"apps"`
}

// DescribeApp adds description of running app
func (s *Spec) DescribeApp(appType string, name string) *AppDescription {
	s.mu.Lock()
	defer s.mu.Unlock()

	appDesc := &AppDescription{
		Type: appType,
	}
	s.Apps[name] = appDesc
	return appDesc
}

// AppDescription describes app running in environment
type AppDescription struct {
	// Type is the type of app
	Type string `json:"type"`

	// IP is the IP reserved for this application
	IP net.IP `json:"ip,omitempty"`

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

	if _, exists := a.Endpoints[name]; exists {
		panic(fmt.Sprintf("endpoint %s already exists", name))
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

	if _, exists := a.Params[name]; exists {
		panic(fmt.Sprintf("param %s already exists", name))
	}
	a.Params[name] = value
}
