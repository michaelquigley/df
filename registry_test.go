package df

import (
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
	assert.NotNil(t, registry.objects)
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
