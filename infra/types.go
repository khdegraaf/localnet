package infra

import (
	"context"
	"fmt"
	"net"
	"sync"
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
	DeployBinary(ctx context.Context, app Binary) (Deployment, error)

	// DeployContainer deploys container to the target
	DeployContainer(ctx context.Context, app Container) (Deployment, error)
}

// File represents file to be created for application
type File struct {
	// Path is the path to file
	Path string

	// Content represents content of file
	Content []byte

	// Preprocess tells if file should be preprocessed using data delivered by target
	Preprocess bool
}

// PreprocessFunc is the function called to preprocess app
type PreprocessFunc func(ctx context.Context) error

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

	// Func is called to preprocess app
	PreFunc PreprocessFunc
}

// Binary represents binary file to be deployed
type Binary struct {
	AppBase

	// Path is the path to binary file
	Path string
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
func NewSpec() *Spec {
	return &Spec{
		Apps: map[string]*AppDescription{},
	}
}

// Spec describes running environment
type Spec struct {
	mu   sync.Mutex
	Apps map[string]*AppDescription `json:"apps"`
}

// DescribeApp adds description of running app
func (s *Spec) DescribeApp(appType string, name string) *AppDescription {
	s.mu.Lock()
	defer s.mu.Unlock()

	appDesc := &AppDescription{
		Type:      appType,
		Endpoints: map[string]string{},
	}
	s.Apps[name] = appDesc
	return appDesc
}

// AppDescription describes app running in environment
type AppDescription struct {
	// Type is the type of app
	Type string `json:"type"`

	// Endpoints describe endpoints exposed by application
	Endpoints map[string]string `json:"endpoints"`
}

// DescribeEndpoint adds endpoint to app description
func (a *AppDescription) DescribeEndpoint(name, endpoint string) {
	if _, exists := a.Endpoints[name]; exists {
		panic(fmt.Sprintf("endpoint %s already exists", name))
	}
	a.Endpoints[name] = endpoint
}
