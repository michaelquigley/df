package df

import (
	"fmt"
	"reflect"
)

// Container is an application container that manages singleton objects by type.
type Container struct {
	objects map[reflect.Type]any
}

// NewContainer creates and returns a new empty container.
func NewContainer() *Container {
	return &Container{
		objects: make(map[reflect.Type]any),
	}
}

// Set registers a singleton object in the container by its type.
// If an object of the same type already exists, it will be replaced.
func (c *Container) Set(object any) {
	c.objects[reflect.TypeOf(object)] = object
}

// Get retrieves an object of type T from the container.
// Returns the object and true if found, or zero value and false if not found.
func Get[T any](c *Container) (T, bool) {
	var zero T
	targetType := reflect.TypeOf(zero)

	obj, exists := c.objects[targetType]
	if !exists {
		return zero, false
	}

	typed, ok := obj.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// ContainerFactory is a function that creates an object using the container
type ContainerFactory[T any] func(*Container) (T, error)

// ContainerLinker is a function that wires dependencies between objects in the container
type ContainerLinker func(*Container) error

// ContainerBuilder provides a fluent API for building containers in phases
type ContainerBuilder struct {
	container *Container
	factories map[reflect.Type]func(*Container) (any, error)
	linkers   []ContainerLinker
}

// NewBuilder creates a new container builder
func NewBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		container: NewContainer(),
		factories: make(map[reflect.Type]func(*Container) (any, error)),
		linkers:   make([]ContainerLinker, 0),
	}
}

// Bind adds a configuration object or pre-existing object to the container
func (b *ContainerBuilder) Bind(object any) *ContainerBuilder {
	b.container.Set(object)
	return b
}

// BindFrom loads and binds configuration from a data source
func (b *ContainerBuilder) BindFrom(data map[string]any, target any) (*ContainerBuilder, error) {
	if err := Bind(target, data); err != nil {
		return b, fmt.Errorf("failed to bind configuration: %w", err)
	}
	b.container.Set(target)
	return b, nil
}

// Factory registers a factory function for creating objects of type T
func Factory[T any](b *ContainerBuilder, factory ContainerFactory[T]) *ContainerBuilder {
	var zero T
	targetType := reflect.TypeOf(zero)
	
	b.factories[targetType] = func(c *Container) (any, error) {
		return factory(c)
	}
	return b
}

// Link registers a linker function for wiring dependencies
func (b *ContainerBuilder) Link(linker ContainerLinker) *ContainerBuilder {
	b.linkers = append(b.linkers, linker)
	return b
}

// Create executes all registered factories to create objects
func (b *ContainerBuilder) Create() (*ContainerBuilder, error) {
	for objectType, factory := range b.factories {
		obj, err := factory(b.container)
		if err != nil {
			return b, fmt.Errorf("factory for type %v failed: %w", objectType, err)
		}
		b.container.objects[objectType] = obj
	}
	return b, nil
}

// Wire executes all registered linkers to wire dependencies
func (b *ContainerBuilder) Wire() (*ContainerBuilder, error) {
	for i, linker := range b.linkers {
		if err := linker(b.container); err != nil {
			return b, fmt.Errorf("linker %d failed: %w", i, err)
		}
	}
	return b, nil
}

// Build returns the final configured container
func (b *ContainerBuilder) Build() *Container {
	return b.container
}
