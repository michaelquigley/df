package df

import (
	"reflect"
)

// Registry is an application container that manages singleton objects by type.
type Registry struct {
	singletons   map[reflect.Type]any
	namedObjects map[namedKey]any
}

// NewRegistry creates and returns a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		singletons:   make(map[reflect.Type]any),
		namedObjects: make(map[namedKey]any),
	}
}

// Visit calls the provided function for each object in the registry.
func (r *Registry) Visit(f func(object any) error) error {
	for _, object := range r.singletons {
		if err := f(object); err != nil {
			return err
		}
	}
	for _, object := range r.namedObjects {
		if err := f(object); err != nil {
			return err
		}
	}
	return nil
}

// Set registers a singleton object in the registry by its type.
// If an object of the same type already exists, it will be replaced.
func (r *Registry) Set(object any) {
	r.singletons[reflect.TypeOf(object)] = object
}

// SetAs registers a singleton object in the registry by the specified type.
// If an object of the same type already exists, it will be replaced.
func SetAs[T any](c *Registry, object T) {
	var zero T
	targetType := reflect.TypeOf(zero)
	c.singletons[targetType] = object
}

// Get retrieves an object of type T from the registry.
// Returns the object and true if found, or zero value and false if not found.
func Get[T any](c *Registry) (T, bool) {
	var zero T
	targetType := reflect.TypeOf(zero)

	obj, exists := c.singletons[targetType]
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
	_, exists := c.singletons[targetType]
	return exists
}

// Remove removes an object of type T from the registry.
// Returns true if the object was found and removed, false if it didn't exist.
func Remove[T any](c *Registry) bool {
	var zero T
	targetType := reflect.TypeOf(zero)
	_, exists := c.singletons[targetType]
	if exists {
		delete(c.singletons, targetType)
		return true
	}
	return false
}

// Clear removes all objects from the registry.
func (r *Registry) Clear() {
	r.singletons = make(map[reflect.Type]any)
	r.namedObjects = make(map[namedKey]any)
}

// Types returns a slice of all registered types in the registry.
// Useful for debugging and introspection.
func (r *Registry) Types() []reflect.Type {
	types := make([]reflect.Type, 0, len(r.singletons))
	for t := range r.singletons {
		types = append(types, t)
	}
	for key := range r.namedObjects {
		found := false
		for _, existing := range types {
			if existing == key.typ {
				found = true
				break
			}
		}
		if !found {
			types = append(types, key.typ)
		}
	}
	return types
}

// SetNamed registers a named object in the registry by its type and name.
// If an object with the same type and name already exists, it will be replaced.
func (r *Registry) SetNamed(name string, object any) {
	key := namedKey{
		typ:  reflect.TypeOf(object),
		name: name,
	}
	r.namedObjects[key] = object
}

// GetNamed retrieves a named object of type T from the registry.
// Returns the object and true if found, or zero value and false if not found.
func GetNamed[T any](r *Registry, name string) (T, bool) {
	var zero T
	key := namedKey{
		typ:  reflect.TypeOf(zero),
		name: name,
	}

	obj, exists := r.namedObjects[key]
	if !exists {
		return zero, false
	}

	typed, ok := obj.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// Query retrieves all objects of type T from the registry (both singleton and named).
// Returns a slice containing the singleton (if exists) followed by all named instances.
func Query[T any](r *Registry) []T {
	var zero T
	targetType := reflect.TypeOf(zero)
	var results []T

	// Add singleton if it exists
	if obj, exists := r.singletons[targetType]; exists {
		if typed, ok := obj.(T); ok {
			results = append(results, typed)
		}
	}

	// Add all named instances
	for key, obj := range r.namedObjects {
		if key.typ == targetType {
			if typed, ok := obj.(T); ok {
				results = append(results, typed)
			}
		}
	}

	return results
}

// namedKey represents a composite key for named object storage.
type namedKey struct {
	typ  reflect.Type
	name string
}
