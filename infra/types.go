package infra

import "context"

// App is the interface exposed by application
type App interface {
	Deploy(ctx context.Context, target Target) error
}

// Env is the environment to deploy
type Env []App

// Deploy deploys app in environment to the target
func (e Env) Deploy(ctx context.Context, t Target) error {
	for _, app := range e {
		if err := app.Deploy(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

// Target represents target of deployment
type Target interface {
	// DeployBinary deploys binary to the target
	DeployBinary(ctx context.Context, app Binary) error

	// DeployContainer deploys container to the target
	DeployContainer(ctx context.Context, app Container) error
}

// Binary represents binary file to be deployed
type Binary struct {
	// Name of the application
	Name string

	// Path is the path to binary file
	Path string

	// Args are args passed to binary
	Args []string
}

// Deploy deploys binary to the target
func (b Binary) Deploy(ctx context.Context, target Target) error {
	return target.DeployBinary(ctx, b)
}

// Container represents container to be deployed
type Container struct {
	// Name of the application
	Name string

	// Image is the url of the container image
	Image string

	// Tag is the tag of the image
	Tag string

	// Args are args passed to the container
	Args []string
}

// Deploy deploys container to the target
func (c Container) Deploy(ctx context.Context, target Target) error {
	return target.DeployContainer(ctx, c)
}
