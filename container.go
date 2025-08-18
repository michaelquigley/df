package df

import (
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
