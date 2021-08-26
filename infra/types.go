package infra

import (
	"context"
	"net"
)

// App is the interface exposed by application
type App interface {
	Deploy(ctx context.Context, config Config, target Target) error
}

// Env is the environment to deploy
type Env []App

// Deploy deploys app in environment to the target
func (e Env) Deploy(ctx context.Context, config Config, t Target) error {
	for _, app := range e {
		if err := app.Deploy(ctx, config, t); err != nil {
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

// Target represents target of deployment
type Target interface {
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
