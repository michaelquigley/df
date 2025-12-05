package da

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/michaelquigley/df/dd"
	"gopkg.in/yaml.v3"
)

// Container is an application container that manages singletons and objects by type and (optionally) by name.
type Container struct {
	singletons    map[reflect.Type]any
	namedObjects  map[namedKey]any
	taggedObjects map[string][]any
}

// NewContainer creates and returns a new empty container.
func NewContainer() *Container {
	return &Container{
		singletons:    make(map[reflect.Type]any),
		namedObjects:  make(map[namedKey]any),
		taggedObjects: make(map[string][]any),
	}
}

// Visit calls the provided function for each object in the container.
// Objects that appear in multiple locations (e.g., both as singleton and tagged) are only visited once.
func (c *Container) Visit(f func(object any) error) error {
	// Track visited objects using pointer addresses for deduplication
	// This works for pointer types; value types are tracked if comparable
	visited := make(map[uintptr]bool)

	markVisited := func(obj any) bool {
		v := reflect.ValueOf(obj)
		// For pointer types, use the pointer address
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if !v.IsNil() {
				ptr := v.Pointer()
				if visited[ptr] {
					return true // already visited
				}
				visited[ptr] = true
			}
		}
		return false
	}

	// Visit singletons
	for _, object := range c.singletons {
		markVisited(object)
		if err := f(object); err != nil {
			return err
		}
	}

	// Visit named objects
	for _, object := range c.namedObjects {
		markVisited(object)
		if err := f(object); err != nil {
			return err
		}
	}

	// Visit tagged objects (skip if already visited)
	for _, objects := range c.taggedObjects {
		for _, object := range objects {
			if !markVisited(object) {
				if err := f(object); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Set registers a singleton object in the container by its type.
// If an object of the same type already exists, it will be replaced.
func Set(c *Container, object any) {
	c.singletons[reflect.TypeOf(object)] = object
}

// SetAs registers a singleton object in the container by the specified type.
// If an object of the same type already exists, it will be replaced.
func SetAs[T any](c *Container, object T) {
	var zero T
	targetType := reflect.TypeOf(zero)
	c.singletons[targetType] = object
}

// SetNamed registers a named object in the container by its type and name.
// If an object with the same type and name already exists, it will be replaced.
func SetNamed(c *Container, name string, object any) {
	key := namedKey{
		typ:  reflect.TypeOf(object),
		name: name,
	}
	c.namedObjects[key] = object
}

// AddTagged adds an object to a tagged collection in the container.
// The same object can be added to multiple tags.
func AddTagged(c *Container, tag string, object any) {
	c.taggedObjects[tag] = append(c.taggedObjects[tag], object)
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

// Tagged retrieves all objects with the specified tag.
func Tagged(c *Container, tag string) []any {
	objects, exists := c.taggedObjects[tag]
	if !exists {
		return nil
	}
	// Return a copy to prevent external modification
	result := make([]any, len(objects))
	copy(result, objects)
	return result
}

// TaggedOfType retrieves all objects with the specified tag that are of type T.
func TaggedOfType[T any](c *Container, tag string) []T {
	objects, exists := c.taggedObjects[tag]
	if !exists {
		return nil
	}
	var zero T
	targetType := reflect.TypeOf(zero)
	var results []T
	for _, obj := range objects {
		if reflect.TypeOf(obj) == targetType {
			if typed, ok := obj.(T); ok {
				results = append(results, typed)
			}
		}
	}
	return results
}

// TaggedAsType retrieves all objects with the specified tag that can be cast to type T.
// This enables finding tagged objects by interface regardless of their concrete type.
func TaggedAsType[T any](c *Container, tag string) []T {
	objects, exists := c.taggedObjects[tag]
	if !exists {
		return nil
	}
	var results []T
	for _, obj := range objects {
		if typed, ok := obj.(T); ok {
			results = append(results, typed)
		}
	}
	return results
}

// Has checks if an object of type T exists in the container.
func Has[T any](c *Container) bool {
	var zero T
	targetType := reflect.TypeOf(zero)
	_, exists := c.singletons[targetType]
	return exists
}

// HasNamed checks if a named object of type T with the given name exists in the container.
func HasNamed[T any](c *Container, name string) bool {
	var zero T
	key := namedKey{
		typ:  reflect.TypeOf(zero),
		name: name,
	}
	_, exists := c.namedObjects[key]
	return exists
}

// HasTagged checks if any objects exist with the specified tag.
func HasTagged(c *Container, tag string) bool {
	objects, exists := c.taggedObjects[tag]
	return exists && len(objects) > 0
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

// RemoveNamed removes a named object of type T with the given name from the container.
// Returns true if the object was found and removed, false if it didn't exist.
func RemoveNamed[T any](c *Container, name string) bool {
	var zero T
	key := namedKey{
		typ:  reflect.TypeOf(zero),
		name: name,
	}
	_, exists := c.namedObjects[key]
	if exists {
		delete(c.namedObjects, key)
		return true
	}
	return false
}

// RemoveTaggedFrom removes a specific object from a specific tag.
// Returns true if the object was found and removed, false otherwise.
func RemoveTaggedFrom(c *Container, tag string, object any) bool {
	objects, exists := c.taggedObjects[tag]
	if !exists {
		return false
	}
	for i, obj := range objects {
		if obj == object {
			c.taggedObjects[tag] = append(objects[:i], objects[i+1:]...)
			// Clean up empty tags
			if len(c.taggedObjects[tag]) == 0 {
				delete(c.taggedObjects, tag)
			}
			return true
		}
	}
	return false
}

// RemoveTagged removes a specific object from ALL tags.
// Returns the number of tags the object was removed from.
func RemoveTagged(c *Container, object any) int {
	count := 0
	for tag := range c.taggedObjects {
		if RemoveTaggedFrom(c, tag, object) {
			count++
		}
	}
	return count
}

// ClearTagged removes all objects with the specified tag.
// Returns the number of objects that were removed.
func ClearTagged(c *Container, tag string) int {
	objects, exists := c.taggedObjects[tag]
	if !exists {
		return 0
	}
	count := len(objects)
	delete(c.taggedObjects, tag)
	return count
}

// Clear removes all objects from the container.
func (c *Container) Clear() {
	c.singletons = make(map[reflect.Type]any)
	c.namedObjects = make(map[namedKey]any)
	c.taggedObjects = make(map[string][]any)
}

// Tags returns a slice of all tags in the container.
func (c *Container) Tags() []string {
	tags := make([]string, 0, len(c.taggedObjects))
	for tag := range c.taggedObjects {
		tags = append(tags, tag)
	}
	return tags
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

// OfType retrieves all objects of type T from the container (singleton, named, and tagged).
// Returns a slice containing the singleton (if exists), followed by named instances, then tagged instances.
// Duplicate objects (same pointer instance in multiple locations) are deduplicated.
func OfType[T any](c *Container) []T {
	var zero T
	targetType := reflect.TypeOf(zero)
	var results []T
	seen := make(map[uintptr]bool)

	// Helper to check if an object (pointer) has been seen
	markSeen := func(obj any) bool {
		v := reflect.ValueOf(obj)
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if !v.IsNil() {
				ptr := v.Pointer()
				if seen[ptr] {
					return true
				}
				seen[ptr] = true
			}
		}
		return false
	}

	// Add singleton if it exists
	if obj, exists := c.singletons[targetType]; exists {
		if typed, ok := obj.(T); ok {
			markSeen(obj)
			results = append(results, typed)
		}
	}

	// Add all named instances
	for key, obj := range c.namedObjects {
		if key.typ == targetType {
			if typed, ok := obj.(T); ok {
				markSeen(obj)
				results = append(results, typed)
			}
		}
	}

	// Add all tagged instances (deduplicated)
	for _, objects := range c.taggedObjects {
		for _, obj := range objects {
			if reflect.TypeOf(obj) == targetType && !markSeen(obj) {
				if typed, ok := obj.(T); ok {
					results = append(results, typed)
				}
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
	Tagged     int `json:"tagged" yaml:"tagged"`
}

// InspectObject represents a single object in the container for inspection.
type InspectObject struct {
	Type    string  `json:"type" yaml:"type"`
	Storage string  `json:"storage" yaml:"storage"`
	Name    *string `json:"name,omitempty" yaml:"name,omitempty"`
	Tag     *string `json:"tag,omitempty" yaml:"tag,omitempty"`
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
			Tag:     nil,
			Value:   fmt.Sprintf("%+v", obj),
		})
	}

	// collect named objects
	for key, obj := range c.namedObjects {
		objects = append(objects, InspectObject{
			Type:    key.typ.String(),
			Storage: "named",
			Name:    &key.name,
			Tag:     nil,
			Value:   fmt.Sprintf("%+v", obj),
		})
	}

	// collect tagged objects
	taggedCount := 0
	for tag, objs := range c.taggedObjects {
		tagCopy := tag // create a copy for the pointer
		for _, obj := range objs {
			objects = append(objects, InspectObject{
				Type:    reflect.TypeOf(obj).String(),
				Storage: "tagged",
				Name:    nil,
				Tag:     &tagCopy,
				Value:   fmt.Sprintf("%+v", obj),
			})
			taggedCount++
		}
	}

	summary := InspectSummary{
		Total:      len(c.singletons) + len(c.namedObjects) + taggedCount,
		Singletons: len(c.singletons),
		Named:      len(c.namedObjects),
		Tagged:     taggedCount,
	}

	return InspectData{
		Summary: summary,
		Objects: objects,
	}
}

func (c *Container) formatHuman(data InspectData) (string, error) {
	return dd.Inspect(data)
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
