package da

import (
	"fmt"
	"path/filepath"

	"github.com/michaelquigley/df/dd"
)

// Factory creates and registers objects in the application container.
// Implementations should use SetAs[T]() to register created objects with the container.
type Factory[C any] interface {
	Build(a *Application[C]) error
}

// FactoryFunc is a function type that implements Factory[C].
// It allows using raw functions as factories without defining separate types.
type FactoryFunc[C any] func(a *Application[C]) error

// Build implements the Factory interface for FactoryFunc.
func (f FactoryFunc[C]) Build(a *Application[C]) error {
	return f(a)
}

// Linkable defines objects that can establish connections to other container objects
// during the linking phase after all objects have been created.
type Linkable interface {
	Link(*Container) error
}

// Startable defines objects that require initialization after linking is complete.
type Startable interface {
	Start() error
}

// Stoppable defines objects that require cleanup during shutdown.
type Stoppable interface {
	Stop() error
}

// Application orchestrates the lifecycle of a dependency injection container with configuration.
// It manages object creation through factories, dependency linking, startup, and shutdown phases.
type Application[C any] struct {
	Cfg       C            // configuration object
	C         *Container   // dependency injection container
	Factories []Factory[C] // factories for creating and registering objects
}

// NewApplication creates a new application with the given configuration.
// The configuration object is automatically registered in the container.
func NewApplication[C any](cfg C) *Application[C] {
	a := &Application[C]{
		Cfg: cfg,
		C:   NewContainer(),
	}
	SetAs[C](a.C, cfg)
	return a
}

// WithFactory adds a factory to the application for fluent configuration.
// Returns the application to enable method chaining.
func WithFactory[C any](a *Application[C], f Factory[C]) *Application[C] {
	a.Factories = append(a.Factories, f)
	return a
}

// WithFactoryFunc adds a function factory to the application for fluent configuration.
// Returns the application to enable method chaining.
func WithFactoryFunc[C any](a *Application[C], f func(a *Application[C]) error) *Application[C] {
	return WithFactory(a, FactoryFunc[C](f))
}

// Initialize executes Configure, Build, and Link phases in sequence.
// Returns on first error without proceeding to subsequent phases.
func (a *Application[C]) Initialize(configPaths ...string) error {
	return a.InitializeWithOptions(nil, configPaths...)
}

// InitializeWithOptions executes Configure, Build, and Link phases in sequence with custom options.
// Returns on first error without proceeding to subsequent phases.
func (a *Application[C]) InitializeWithOptions(opts *dd.Options, configPaths ...string) error {
	for _, path := range configPaths {
		if err := a.Configure(path, opts); err != nil {
			return err
		}
	}

	if err := a.Build(); err != nil {
		return err
	}

	return a.Link()
}

// Configure loads additional configuration from a file and merges it with the existing configuration.
// Supports JSON and YAML file formats based on file extension.
func (a *Application[C]) Configure(path string, opts ...*dd.Options) error {
	pathExt := filepath.Ext(path)
	if pathExt == ".yaml" || pathExt == ".yml" {
		if err := dd.MergeFromYAML(a.Cfg, path, opts...); err != nil {
			return err
		}
	} else if pathExt == ".json" {
		if err := dd.MergeFromJSON(a.Cfg, path, opts...); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported configuration file extension: %s", pathExt)
	}
	return nil
}

// Build executes all registered factories to create and register objects in the container.
// Factories are responsible for calling SetAs[T]() to register their created objects.
func (a *Application[C]) Build() error {
	for _, f := range a.Factories {
		if err := f.Build(a); err != nil {
			return err
		}
	}
	return nil
}

// Link establishes dependencies between objects by calling Link() on all Linkable objects.
// This phase occurs after Build() to ensure all objects exist before dependency resolution.
// Returns the first error encountered, which stops the linking process.
func (a *Application[C]) Link() error {
	return a.C.Visit(func(object any) error {
		if l, ok := object.(Linkable); ok {
			return l.Link(a.C)
		}
		return nil
	})
}

// Start initializes all Startable objects after linking is complete.
// Returns the first error encountered, which stops the startup process.
func (a *Application[C]) Start() error {
	return a.C.Visit(func(object any) error {
		if startable, ok := object.(Startable); ok {
			return startable.Start()
		}
		return nil
	})
}

// Stop shuts down all Stoppable objects for graceful cleanup.
// Returns the first error encountered, but continues attempting to stop remaining objects.
func (a *Application[C]) Stop() error {
	var firstError error

	err := a.C.Visit(func(object any) error {
		if stoppable, ok := object.(Stoppable); ok {
			if err := stoppable.Stop(); err != nil && firstError == nil {
				firstError = err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return firstError
}
