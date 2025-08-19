package df

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Container is an application container that manages singletons and objects by type and (optionally) by name.
type Container struct {
	singletons   map[reflect.Type]any
	namedObjects map[namedKey]any
}

// NewContainer creates and returns a new empty container.
func NewContainer() *Container {
	return &Container{
		singletons:   make(map[reflect.Type]any),
		namedObjects: make(map[namedKey]any),
	}
}

// Visit calls the provided function for each object in the container.
func (c *Container) Visit(f func(object any) error) error {
	for _, object := range c.singletons {
		if err := f(object); err != nil {
			return err
		}
	}
	for _, object := range c.namedObjects {
		if err := f(object); err != nil {
			return err
		}
	}
	return nil
}

// Set registers a singleton object in the container by its type.
// If an object of the same type already exists, it will be replaced.
func (c *Container) Set(object any) {
	c.singletons[reflect.TypeOf(object)] = object
}

// SetAs registers a singleton object in the container by the specified type.
// If an object of the same type already exists, it will be replaced.
func SetAs[T any](c *Container, object T) {
	var zero T
	targetType := reflect.TypeOf(zero)
	c.singletons[targetType] = object
}

// Get retrieves an object of type T from the container.
// Returns the object and true if found, or zero value and false if not found.
func Get[T any](c *Container) (T, bool) {
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

// Has checks if an object of type T exists in the container.
func Has[T any](c *Container) bool {
	var zero T
	targetType := reflect.TypeOf(zero)
	_, exists := c.singletons[targetType]
	return exists
}

// Remove removes an object of type T from the container.
// Returns true if the object was found and removed, false if it didn't exist.
func Remove[T any](c *Container) bool {
	var zero T
	targetType := reflect.TypeOf(zero)
	_, exists := c.singletons[targetType]
	if exists {
		delete(c.singletons, targetType)
		return true
	}
	return false
}

// Clear removes all objects from the container.
func (c *Container) Clear() {
	c.singletons = make(map[reflect.Type]any)
	c.namedObjects = make(map[namedKey]any)
}

// Types returns a slice of all registered types in the container.
// Useful for debugging and introspection.
func (c *Container) Types() []reflect.Type {
	types := make([]reflect.Type, 0, len(c.singletons))
	for t := range c.singletons {
		types = append(types, t)
	}
	for key := range c.namedObjects {
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

// SetNamed registers a named object in the container by its type and name.
// If an object with the same type and name already exists, it will be replaced.
func (c *Container) SetNamed(name string, object any) {
	key := namedKey{
		typ:  reflect.TypeOf(object),
		name: name,
	}
	c.namedObjects[key] = object
}

// GetNamed retrieves a named object of type T from the container.
// Returns the object and true if found, or zero value and false if not found.
func GetNamed[T any](c *Container, name string) (T, bool) {
	var zero T
	key := namedKey{
		typ:  reflect.TypeOf(zero),
		name: name,
	}

	obj, exists := c.namedObjects[key]
	if !exists {
		return zero, false
	}

	typed, ok := obj.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// OfType retrieves all objects of type T from the container (both singleton and named).
// Returns a slice containing the singleton (if exists) followed by all named instances.
func OfType[T any](c *Container) []T {
	var zero T
	targetType := reflect.TypeOf(zero)
	var results []T

	// Add singleton if it exists
	if obj, exists := c.singletons[targetType]; exists {
		if typed, ok := obj.(T); ok {
			results = append(results, typed)
		}
	}

	// Add all named instances
	for key, obj := range c.namedObjects {
		if key.typ == targetType {
			if typed, ok := obj.(T); ok {
				results = append(results, typed)
			}
		}
	}

	return results
}

// AsType visits all objects in the container and returns any that can be cast to type T.
// This enables finding objects by interface or supertype regardless of their registration type.
func AsType[T any](c *Container) []T {
	var results []T

	err := c.Visit(func(object any) error {
		if typed, ok := object.(T); ok {
			results = append(results, typed)
		}
		return nil
	})

	// Visit should never return an error in this context
	_ = err

	return results
}

// namedKey represents a composite key for named object storage.
type namedKey struct {
	typ  reflect.Type
	name string
}

// InspectFormat defines the output format for container inspection.
type InspectFormat string

const (
	InspectHuman InspectFormat = "human"
	InspectJSON  InspectFormat = "json"
	InspectYAML  InspectFormat = "yaml"
)

// InspectData represents the structured data for container inspection.
type InspectData struct {
	Summary InspectSummary  `json:"summary" yaml:"summary"`
	Objects []InspectObject `json:"objects" yaml:"objects"`
}

// InspectSummary provides aggregate statistics about the container.
type InspectSummary struct {
	Total      int `json:"total" yaml:"total"`
	Singletons int `json:"singletons" yaml:"singletons"`
	Named      int `json:"named" yaml:"named"`
}

// InspectObject represents a single object in the container for inspection.
type InspectObject struct {
	Type    string  `json:"type" yaml:"type"`
	Storage string  `json:"storage" yaml:"storage"`
	Name    *string `json:"name" yaml:"name"`
	Value   string  `json:"value" yaml:"value"`
}

// Inspect returns a formatted representation of the container contents.
// Supports table, JSON, and YAML formats for human and machine consumption.
func (c *Container) Inspect(format InspectFormat) (string, error) {
	data := c.gatherInspectData()

	switch format {
	case InspectHuman:
		return c.formatHuman(data)
	case InspectJSON:
		return c.formatJSON(data)
	case InspectYAML:
		return c.formatYAML(data)
	default:
		return "", fmt.Errorf("unsupported inspect format: %s", format)
	}
}

// gatherInspectData collects all container objects into structured data.
func (c *Container) gatherInspectData() InspectData {
	var objects []InspectObject

	// collect singletons
	for typ, obj := range c.singletons {
		objects = append(objects, InspectObject{
			Type:    typ.String(),
			Storage: "singleton",
			Name:    nil,
			Value:   fmt.Sprintf("%+v", obj),
		})
	}

	// collect named objects
	for key, obj := range c.namedObjects {
		objects = append(objects, InspectObject{
			Type:    key.typ.String(),
			Storage: "named",
			Name:    &key.name,
			Value:   fmt.Sprintf("%+v", obj),
		})
	}

	summary := InspectSummary{
		Total:      len(c.singletons) + len(c.namedObjects),
		Singletons: len(c.singletons),
		Named:      len(c.namedObjects),
	}

	return InspectData{
		Summary: summary,
		Objects: objects,
	}
}

func (c *Container) formatHuman(data InspectData) (string, error) {
	return Inspect(data)
}

// formatJSON creates JSON output.
func (c *Container) formatJSON(data InspectData) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// formatYAML creates YAML output.
func (c *Container) formatYAML(data InspectData) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(bytes), nil
}
