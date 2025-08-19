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

// Visit calls the provided function for each object in the registry.
func (r *Registry) Visit(f func(object any) error) error {
	for _, object := range r.objects {
		if err := f(object); err != nil {
			return err
		}
	}
	return nil
}

// Set registers a singleton object in the registry by its type.
// If an object of the same type already exists, it will be replaced.
func (r *Registry) Set(object any) {
	r.objects[reflect.TypeOf(object)] = object
}

// SetAs registers a singleton object in the registry by the specified type.
// If an object of the same type already exists, it will be replaced.
func SetAs[T any](c *Registry, object T) {
	var zero T
	targetType := reflect.TypeOf(zero)
	c.objects[targetType] = object
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

// Has checks if an object of type T exists in the registry.
func Has[T any](c *Registry) bool {
	var zero T
	targetType := reflect.TypeOf(zero)
	_, exists := c.objects[targetType]
	return exists
}

// Remove removes an object of type T from the registry.
// Returns true if the object was found and removed, false if it didn't exist.
func Remove[T any](c *Registry) bool {
	var zero T
	targetType := reflect.TypeOf(zero)
	_, exists := c.objects[targetType]
	if exists {
		delete(c.objects, targetType)
		return true
	}
	return false
}

// Clear removes all objects from the registry.
func (r *Registry) Clear() {
	r.objects = make(map[reflect.Type]any)
}

// Types returns a slice of all registered types in the registry.
// Useful for debugging and introspection.
func (r *Registry) Types() []reflect.Type {
	types := make([]reflect.Type, 0, len(r.objects))
	for t := range r.objects {
		types = append(types, t)
	}
	return types
}
