package df

import (
	"reflect"
)

// Registry is an application container that manages singleton objects by type.
type Registry struct {
	objects map[reflect.Type]any
}

// NewRegistry creates and returns a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		objects: make(map[reflect.Type]any),
	}
}

// Set registers a singleton object in the registry by its type.
// If an object of the same type already exists, it will be replaced.
func (r *Registry) Set(object any) {
	r.objects[reflect.TypeOf(object)] = object
}

// Get retrieves an object of type T from the registry.
// Returns the object and true if found, or zero value and false if not found.
func Get[T any](c *Registry) (T, bool) {
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
