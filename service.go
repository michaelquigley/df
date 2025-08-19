package df

import (
	"fmt"
	"path/filepath"
)

// Factory creates and registers objects in the service registry.
// Implementations should use SetAs[T]() to register created objects with the registry.
type Factory[C any] interface {
	Build(s *Service[C]) error
}

// Linkable defines objects that can establish connections to other registry objects
// during the linking phase after all objects have been created.
type Linkable interface {
	Link(*Registry) error
}

// Startable defines objects that require initialization after linking is complete.
type Startable interface {
	Start() error
}

// Stoppable defines objects that require cleanup during shutdown.
type Stoppable interface {
	Stop() error
}

// Service orchestrates the lifecycle of a dependency injection container with configuration.
// It manages object creation through factories, dependency linking, startup, and shutdown phases.
type Service[C any] struct {
	Cfg       C            // configuration object
	R         *Registry    // dependency injection registry
	Factories []Factory[C] // factories for creating and registering objects
}

// NewService creates a new service with the given configuration.
// The configuration object is automatically registered in the registry.
func NewService[C any](cfg C) *Service[C] {
	s := &Service[C]{
		Cfg: cfg,
		R:   NewRegistry(),
	}
	SetAs[C](s.R, cfg)
	return s
}

// WithFactory adds a factory to the service for fluent configuration.
// Returns the service to enable method chaining.
func WithFactory[C any](s *Service[C], f Factory[C]) *Service[C] {
	s.Factories = append(s.Factories, f)
	return s
}

// Initialize executes Configure, Build, and Link phases in sequence.
// Returns on first error without proceeding to subsequent phases.
func (s *Service[C]) Initialize(configPaths ...string) error {
	for _, path := range configPaths {
		if err := s.Configure(path); err != nil {
			return err
		}
	}

	if err := s.Build(); err != nil {
		return err
	}

	return s.Link()
}

// Configure loads additional configuration from a file and merges it with the existing configuration.
// Supports JSON and YAML file formats based on file extension.
func (s *Service[C]) Configure(path string) error {
	pathExt := filepath.Ext(path)
	if pathExt == ".yaml" || pathExt == ".yml" {
		if err := MergeFromYAML(s.Cfg, path); err != nil {
			return err
		}
	} else if pathExt == ".json" {
		if err := MergeFromJSON(s.Cfg, path); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported configuration file extension: %s", pathExt)
	}
	return nil
}

// Build executes all registered factories to create and register objects in the registry.
// Factories are responsible for calling SetAs[T]() to register their created objects.
func (s *Service[C]) Build() error {
	for _, f := range s.Factories {
		if err := f.Build(s); err != nil {
			return err
		}
	}
	return nil
}

// Link establishes dependencies between objects by calling Link() on all Linkable objects.
// This phase occurs after Build() to ensure all objects exist before dependency resolution.
// Returns the first error encountered, which stops the linking process.
func (s *Service[C]) Link() error {
	return s.R.Visit(func(object any) error {
		if l, ok := object.(Linkable); ok {
			return l.Link(s.R)
		}
		return nil
	})
}

// Start initializes all Startable objects after linking is complete.
// Returns the first error encountered, which stops the startup process.
func (s *Service[C]) Start() error {
	return s.R.Visit(func(object any) error {
		if startable, ok := object.(Startable); ok {
			return startable.Start()
		}
		return nil
	})
}

// Stop shuts down all Stoppable objects for graceful cleanup.
// Returns the first error encountered, but continues attempting to stop remaining objects.
func (s *Service[C]) Stop() error {
	var firstError error

	err := s.R.Visit(func(object any) error {
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
