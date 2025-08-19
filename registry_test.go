package df

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type registryTestService struct {
	name string
}

type registryTestRepository struct {
	database string
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.singletons)
	assert.NotNil(t, registry.namedObjects)
}

func TestRegistry_Set_And_Get(t *testing.T) {
	registry := NewRegistry()

	// test setting and getting a service
	service := &registryTestService{name: "test service"}
	registry.Set(service)

	retrieved, found := Get[*registryTestService](registry)
	assert.True(t, found)
	assert.Equal(t, service, retrieved)
	assert.Equal(t, "test service", retrieved.name)
}

func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry()

	// try to get an object that was never set
	retrieved, found := Get[*registryTestService](registry)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestRegistry_Set_Replace(t *testing.T) {
	registry := NewRegistry()

	// set initial service
	service1 := &registryTestService{name: "first service"}
	registry.Set(service1)

	// replace with new service of same type
	service2 := &registryTestService{name: "second service"}
	registry.Set(service2)

	// should get the second service
	retrieved, found := Get[*registryTestService](registry)
	assert.True(t, found)
	assert.Equal(t, service2, retrieved)
	assert.Equal(t, "second service", retrieved.name)
}

func TestRegistry_Multiple_Types(t *testing.T) {
	registry := NewRegistry()

	// set objects of different types
	service := &registryTestService{name: "my service"}
	repo := &registryTestRepository{database: "my database"}

	registry.Set(service)
	registry.Set(repo)

	// retrieve both objects
	retrievedService, foundService := Get[*registryTestService](registry)
	assert.True(t, foundService)
	assert.Equal(t, service, retrievedService)

	retrievedRepo, foundRepo := Get[*registryTestRepository](registry)
	assert.True(t, foundRepo)
	assert.Equal(t, repo, retrievedRepo)
}

func TestRegistry_Value_Types(t *testing.T) {
	registry := NewRegistry()

	// test with value types
	registry.Set(42)
	registry.Set("hello world")

	retrievedInt, foundInt := Get[int](registry)
	assert.True(t, foundInt)
	assert.Equal(t, 42, retrievedInt)

	retrievedString, foundString := Get[string](registry)
	assert.True(t, foundString)
	assert.Equal(t, "hello world", retrievedString)
}

func TestRegistry_Interface_Types(t *testing.T) {
	registry := NewRegistry()

	// test with interface
	var writer interface {
		Write([]byte) (int, error)
	}

	// this won't work as expected because we're storing the concrete type
	// but trying to retrieve by interface type
	service := &registryTestService{name: "service"}
	registry.Set(service)

	retrieved, found := Get[interface{}](registry)
	assert.False(t, found)
	assert.Nil(t, retrieved)

	// but this will work - storing and retrieving by the same interface type
	registry.Set(writer)
	retrievedWriter, foundWriter := Get[interface {
		Write([]byte) (int, error)
	}](registry)
	assert.False(t, foundWriter) // this will also be false because writer is nil
	assert.Nil(t, retrievedWriter)
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()

	// test has with empty registry
	assert.False(t, Has[*registryTestService](registry))
	assert.False(t, Has[int](registry))

	// set a service and test has
	service := &registryTestService{name: "test service"}
	registry.Set(service)

	assert.True(t, Has[*registryTestService](registry))
	assert.False(t, Has[*registryTestRepository](registry))
	assert.False(t, Has[int](registry))

	// set different types
	registry.Set(42)
	registry.Set("hello")

	assert.True(t, Has[*registryTestService](registry))
	assert.True(t, Has[int](registry))
	assert.True(t, Has[string](registry))
	assert.False(t, Has[*registryTestRepository](registry))
}

func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry()

	// test remove from empty registry
	removed := Remove[*registryTestService](registry)
	assert.False(t, removed)

	// set a service and remove it
	service := &registryTestService{name: "test service"}
	registry.Set(service)

	assert.True(t, Has[*registryTestService](registry))
	
	removed = Remove[*registryTestService](registry)
	assert.True(t, removed)
	assert.False(t, Has[*registryTestService](registry))

	// try to remove again
	removed = Remove[*registryTestService](registry)
	assert.False(t, removed)

	// test remove with multiple types
	service2 := &registryTestService{name: "service2"}
	repo := &registryTestRepository{database: "db"}
	registry.Set(service2)
	registry.Set(repo)
	registry.Set(42)

	assert.True(t, Has[*registryTestService](registry))
	assert.True(t, Has[*registryTestRepository](registry))
	assert.True(t, Has[int](registry))

	// remove service, others should remain
	removed = Remove[*registryTestService](registry)
	assert.True(t, removed)
	assert.False(t, Has[*registryTestService](registry))
	assert.True(t, Has[*registryTestRepository](registry))
	assert.True(t, Has[int](registry))
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	// test clear on empty registry
	registry.Clear()
	assert.False(t, Has[*registryTestService](registry))

	// add multiple objects
	service := &registryTestService{name: "service"}
	repo := &registryTestRepository{database: "db"}
	registry.Set(service)
	registry.Set(repo)
	registry.Set(42)
	registry.Set("hello")

	// verify objects exist
	assert.True(t, Has[*registryTestService](registry))
	assert.True(t, Has[*registryTestRepository](registry))
	assert.True(t, Has[int](registry))
	assert.True(t, Has[string](registry))

	// clear registry
	registry.Clear()

	// verify all objects are gone
	assert.False(t, Has[*registryTestService](registry))
	assert.False(t, Has[*registryTestRepository](registry))
	assert.False(t, Has[int](registry))
	assert.False(t, Has[string](registry))

	// verify we can add objects after clear
	registry.Set(&registryTestService{name: "new service"})
	assert.True(t, Has[*registryTestService](registry))
}

func TestRegistry_Types(t *testing.T) {
	registry := NewRegistry()

	// test types on empty registry
	types := registry.Types()
	assert.Empty(t, types)

	// add single object
	service := &registryTestService{name: "service"}
	registry.Set(service)

	types = registry.Types()
	assert.Len(t, types, 1)
	assert.Contains(t, types, reflect.TypeOf(service))

	// add multiple objects
	repo := &registryTestRepository{database: "db"}
	registry.Set(repo)
	registry.Set(42)
	registry.Set("hello")

	types = registry.Types()
	assert.Len(t, types, 4)
	assert.Contains(t, types, reflect.TypeOf(service))
	assert.Contains(t, types, reflect.TypeOf(repo))
	assert.Contains(t, types, reflect.TypeOf(42))
	assert.Contains(t, types, reflect.TypeOf("hello"))

	// remove an object
	Remove[*registryTestService](registry)
	types = registry.Types()
	assert.Len(t, types, 3)
	assert.NotContains(t, types, reflect.TypeOf(service))

	// clear and verify empty
	registry.Clear()
	types = registry.Types()
	assert.Empty(t, types)
}

func TestRegistry_SetNamed_And_GetNamed(t *testing.T) {
	registry := NewRegistry()

	// test setting and getting named objects
	service1 := &registryTestService{name: "primary service"}
	service2 := &registryTestService{name: "secondary service"}
	
	registry.SetNamed("primary", service1)
	registry.SetNamed("secondary", service2)

	// retrieve named objects
	retrieved1, found1 := GetNamed[*registryTestService](registry, "primary")
	assert.True(t, found1)
	assert.Equal(t, service1, retrieved1)
	assert.Equal(t, "primary service", retrieved1.name)

	retrieved2, found2 := GetNamed[*registryTestService](registry, "secondary")
	assert.True(t, found2)
	assert.Equal(t, service2, retrieved2)
	assert.Equal(t, "secondary service", retrieved2.name)
}

func TestRegistry_GetNamed_NotFound(t *testing.T) {
	registry := NewRegistry()

	// try to get a named object that was never set
	retrieved, found := GetNamed[*registryTestService](registry, "nonexistent")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestRegistry_SetNamed_Replace(t *testing.T) {
	registry := NewRegistry()

	// set initial named service
	service1 := &registryTestService{name: "first service"}
	registry.SetNamed("test", service1)

	// replace with new service of same type and name
	service2 := &registryTestService{name: "second service"}
	registry.SetNamed("test", service2)

	// should get the second service
	retrieved, found := GetNamed[*registryTestService](registry, "test")
	assert.True(t, found)
	assert.Equal(t, service2, retrieved)
	assert.Equal(t, "second service", retrieved.name)
}

func TestRegistry_Named_And_Singleton_Coexistence(t *testing.T) {
	registry := NewRegistry()

	// set singleton and named objects of same type
	singleton := &registryTestService{name: "singleton"}
	named1 := &registryTestService{name: "named1"}
	named2 := &registryTestService{name: "named2"}

	registry.Set(singleton)
	registry.SetNamed("first", named1)
	registry.SetNamed("second", named2)

	// verify all can be retrieved independently
	retrievedSingleton, foundSingleton := Get[*registryTestService](registry)
	assert.True(t, foundSingleton)
	assert.Equal(t, singleton, retrievedSingleton)

	retrievedNamed1, foundNamed1 := GetNamed[*registryTestService](registry, "first")
	assert.True(t, foundNamed1)
	assert.Equal(t, named1, retrievedNamed1)

	retrievedNamed2, foundNamed2 := GetNamed[*registryTestService](registry, "second")
	assert.True(t, foundNamed2)
	assert.Equal(t, named2, retrievedNamed2)
}

func TestRegistry_OfType(t *testing.T) {
	registry := NewRegistry()

	// test OfType with empty registry
	results := OfType[*registryTestService](registry)
	assert.Empty(t, results)

	// add singleton only
	singleton := &registryTestService{name: "singleton"}
	registry.Set(singleton)

	results = OfType[*registryTestService](registry)
	assert.Len(t, results, 1)
	assert.Contains(t, results, singleton)

	// add named objects
	named1 := &registryTestService{name: "named1"}
	named2 := &registryTestService{name: "named2"}
	registry.SetNamed("first", named1)
	registry.SetNamed("second", named2)

	results = OfType[*registryTestService](registry)
	assert.Len(t, results, 3)
	assert.Contains(t, results, singleton)
	assert.Contains(t, results, named1)
	assert.Contains(t, results, named2)

	// test with different type should be empty
	repoResults := OfType[*registryTestRepository](registry)
	assert.Empty(t, repoResults)

	// add different type
	repo := &registryTestRepository{database: "test db"}
	registry.Set(repo)

	repoResults = OfType[*registryTestRepository](registry)
	assert.Len(t, repoResults, 1)
	assert.Contains(t, repoResults, repo)

	// original type results should be unchanged
	results = OfType[*registryTestService](registry)
	assert.Len(t, results, 3)
}

type testInterface interface {
	TestMethod() string
}

type testImplementer1 struct {
	value string
}

func (t *testImplementer1) TestMethod() string {
	return t.value
}

type testImplementer2 struct {
	number int
}

func (t *testImplementer2) TestMethod() string {
	return fmt.Sprintf("number: %d", t.number)
}

func TestRegistry_AsType(t *testing.T) {
	registry := NewRegistry()

	// test AsType with empty registry
	results := AsType[testInterface](registry)
	assert.Empty(t, results)

	// add objects that implement the interface
	impl1 := &testImplementer1{value: "test1"}
	impl2 := &testImplementer2{number: 42}
	nonImpl := &registryTestService{name: "not an implementer"}

	registry.Set(impl1)
	registry.SetNamed("impl2", impl2)
	registry.Set(nonImpl)

	// AsType should return only the implementers
	results = AsType[testInterface](registry)
	assert.Len(t, results, 2)

	// verify both implementers are present
	var foundImpl1, foundImpl2 bool
	for _, result := range results {
		if result.TestMethod() == "test1" {
			foundImpl1 = true
		}
		if result.TestMethod() == "number: 42" {
			foundImpl2 = true
		}
	}
	assert.True(t, foundImpl1)
	assert.True(t, foundImpl2)

	// test AsType with concrete type should work like OfType
	concreteResults := AsType[*testImplementer1](registry)
	assert.Len(t, concreteResults, 1)
	assert.Equal(t, impl1, concreteResults[0])
}

func TestRegistry_Visit_WithNamedObjects(t *testing.T) {
	registry := NewRegistry()
	
	// add singleton and named objects
	singleton := &registryTestService{name: "singleton"}
	named1 := &registryTestService{name: "named1"}
	named2 := &registryTestRepository{database: "named db"}
	
	registry.Set(singleton)
	registry.SetNamed("service1", named1)
	registry.SetNamed("repo1", named2)
	
	// collect all visited objects
	var visited []any
	err := registry.Visit(func(object any) error {
		visited = append(visited, object)
		return nil
	})
	
	assert.NoError(t, err)
	assert.Len(t, visited, 3)
	assert.Contains(t, visited, singleton)
	assert.Contains(t, visited, named1)
	assert.Contains(t, visited, named2)
}

func TestRegistry_Clear_WithNamedObjects(t *testing.T) {
	registry := NewRegistry()
	
	// add singleton and named objects
	registry.Set(&registryTestService{name: "singleton"})
	registry.SetNamed("named", &registryTestService{name: "named"})
	
	// verify objects exist
	assert.True(t, Has[*registryTestService](registry))
	_, found := GetNamed[*registryTestService](registry, "named")
	assert.True(t, found)
	
	// clear registry
	registry.Clear()
	
	// verify all objects are gone
	assert.False(t, Has[*registryTestService](registry))
	_, found = GetNamed[*registryTestService](registry, "named")
	assert.False(t, found)
}

func TestRegistry_Types_WithNamedObjects(t *testing.T) {
	registry := NewRegistry()
	
	// add singleton and named objects of same and different types
	registry.Set(&registryTestService{name: "singleton"})
	registry.SetNamed("named1", &registryTestService{name: "named1"})
	registry.SetNamed("named2", &registryTestRepository{database: "named"})
	
	types := registry.Types()
	
	// should have both types represented, but no duplicates
	assert.Len(t, types, 2)
	assert.Contains(t, types, reflect.TypeOf(&registryTestService{}))
	assert.Contains(t, types, reflect.TypeOf(&registryTestRepository{}))
}
