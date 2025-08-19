package df

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type containerTestService struct {
	name string
}

type containerTestRepository struct {
	database string
}

func TestNewContainer(t *testing.T) {
	container := NewContainer()
	assert.NotNil(t, container)
	assert.NotNil(t, container.singletons)
	assert.NotNil(t, container.namedObjects)
}

func TestContainer_Set_And_Get(t *testing.T) {
	container := NewContainer()

	// test setting and getting a service
	service := &containerTestService{name: "test service"}
	Set(container, service)

	retrieved, found := Get[*containerTestService](container)
	assert.True(t, found)
	assert.Equal(t, service, retrieved)
	assert.Equal(t, "test service", retrieved.name)
}

func TestContainer_Get_NotFound(t *testing.T) {
	container := NewContainer()

	// try to get an object that was never set
	retrieved, found := Get[*containerTestService](container)
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestContainer_Set_Replace(t *testing.T) {
	container := NewContainer()

	// set initial service
	service1 := &containerTestService{name: "first service"}
	Set(container, service1)

	// replace with new service of same type
	service2 := &containerTestService{name: "second service"}
	Set(container, service2)

	// should get the second service
	retrieved, found := Get[*containerTestService](container)
	assert.True(t, found)
	assert.Equal(t, service2, retrieved)
	assert.Equal(t, "second service", retrieved.name)
}

func TestContainer_Multiple_Types(t *testing.T) {
	container := NewContainer()

	// set objects of different types
	service := &containerTestService{name: "my service"}
	repo := &containerTestRepository{database: "my database"}

	Set(container, service)
	Set(container, repo)

	// retrieve both objects
	retrievedService, foundService := Get[*containerTestService](container)
	assert.True(t, foundService)
	assert.Equal(t, service, retrievedService)

	retrievedRepo, foundRepo := Get[*containerTestRepository](container)
	assert.True(t, foundRepo)
	assert.Equal(t, repo, retrievedRepo)
}

func TestContainer_Value_Types(t *testing.T) {
	container := NewContainer()

	// test with value types
	Set(container, 42)
	Set(container, "hello world")

	retrievedInt, foundInt := Get[int](container)
	assert.True(t, foundInt)
	assert.Equal(t, 42, retrievedInt)

	retrievedString, foundString := Get[string](container)
	assert.True(t, foundString)
	assert.Equal(t, "hello world", retrievedString)
}

func TestContainer_Interface_Types(t *testing.T) {
	container := NewContainer()

	// test with interface
	var writer interface {
		Write([]byte) (int, error)
	}

	// this won't work as expected because we're storing the concrete type
	// but trying to retrieve by interface type
	service := &containerTestService{name: "service"}
	Set(container, service)

	retrieved, found := Get[interface{}](container)
	assert.False(t, found)
	assert.Nil(t, retrieved)

	// but this will work - storing and retrieving by the same interface type
	Set(container, writer)
	retrievedWriter, foundWriter := Get[interface {
		Write([]byte) (int, error)
	}](container)
	assert.False(t, foundWriter) // this will also be false because writer is nil
	assert.Nil(t, retrievedWriter)
}

func TestContainer_Has(t *testing.T) {
	container := NewContainer()

	// test has with empty container
	assert.False(t, Has[*containerTestService](container))
	assert.False(t, Has[int](container))

	// set a service and test has
	service := &containerTestService{name: "test service"}
	Set(container, service)

	assert.True(t, Has[*containerTestService](container))
	assert.False(t, Has[*containerTestRepository](container))
	assert.False(t, Has[int](container))

	// set different types
	Set(container, 42)
	Set(container, "hello")

	assert.True(t, Has[*containerTestService](container))
	assert.True(t, Has[int](container))
	assert.True(t, Has[string](container))
	assert.False(t, Has[*containerTestRepository](container))
}

func TestContainer_Remove(t *testing.T) {
	container := NewContainer()

	// test remove from empty container
	removed := Remove[*containerTestService](container)
	assert.False(t, removed)

	// set a service and remove it
	service := &containerTestService{name: "test service"}
	Set(container, service)

	assert.True(t, Has[*containerTestService](container))

	removed = Remove[*containerTestService](container)
	assert.True(t, removed)
	assert.False(t, Has[*containerTestService](container))

	// try to remove again
	removed = Remove[*containerTestService](container)
	assert.False(t, removed)

	// test remove with multiple types
	service2 := &containerTestService{name: "service2"}
	repo := &containerTestRepository{database: "db"}
	Set(container, service2)
	Set(container, repo)
	Set(container, 42)

	assert.True(t, Has[*containerTestService](container))
	assert.True(t, Has[*containerTestRepository](container))
	assert.True(t, Has[int](container))

	// remove service, others should remain
	removed = Remove[*containerTestService](container)
	assert.True(t, removed)
	assert.False(t, Has[*containerTestService](container))
	assert.True(t, Has[*containerTestRepository](container))
	assert.True(t, Has[int](container))
}

func TestContainer_Clear(t *testing.T) {
	container := NewContainer()

	// test clear on empty container
	container.Clear()
	assert.False(t, Has[*containerTestService](container))

	// add multiple objects
	service := &containerTestService{name: "service"}
	repo := &containerTestRepository{database: "db"}
	Set(container, service)
	Set(container, repo)
	Set(container, 42)
	Set(container, "hello")

	// verify objects exist
	assert.True(t, Has[*containerTestService](container))
	assert.True(t, Has[*containerTestRepository](container))
	assert.True(t, Has[int](container))
	assert.True(t, Has[string](container))

	// clear container
	container.Clear()

	// verify all objects are gone
	assert.False(t, Has[*containerTestService](container))
	assert.False(t, Has[*containerTestRepository](container))
	assert.False(t, Has[int](container))
	assert.False(t, Has[string](container))

	// verify we can add objects after clear
	Set(container, &containerTestService{name: "new service"})
	assert.True(t, Has[*containerTestService](container))
}

func TestContainer_Types(t *testing.T) {
	container := NewContainer()

	// test types on empty container
	types := container.Types()
	assert.Empty(t, types)

	// add single object
	service := &containerTestService{name: "service"}
	Set(container, service)

	types = container.Types()
	assert.Len(t, types, 1)
	assert.Contains(t, types, reflect.TypeOf(service))

	// add multiple objects
	repo := &containerTestRepository{database: "db"}
	Set(container, repo)
	Set(container, 42)
	Set(container, "hello")

	types = container.Types()
	assert.Len(t, types, 4)
	assert.Contains(t, types, reflect.TypeOf(service))
	assert.Contains(t, types, reflect.TypeOf(repo))
	assert.Contains(t, types, reflect.TypeOf(42))
	assert.Contains(t, types, reflect.TypeOf("hello"))

	// remove an object
	Remove[*containerTestService](container)
	types = container.Types()
	assert.Len(t, types, 3)
	assert.NotContains(t, types, reflect.TypeOf(service))

	// clear and verify empty
	container.Clear()
	types = container.Types()
	assert.Empty(t, types)
}

func TestContainer_SetNamed_And_GetNamed(t *testing.T) {
	container := NewContainer()

	// test setting and getting named objects
	service1 := &containerTestService{name: "primary service"}
	service2 := &containerTestService{name: "secondary service"}

	container.SetNamed("primary", service1)
	container.SetNamed("secondary", service2)

	// retrieve named objects
	retrieved1, found1 := GetNamed[*containerTestService](container, "primary")
	assert.True(t, found1)
	assert.Equal(t, service1, retrieved1)
	assert.Equal(t, "primary service", retrieved1.name)

	retrieved2, found2 := GetNamed[*containerTestService](container, "secondary")
	assert.True(t, found2)
	assert.Equal(t, service2, retrieved2)
	assert.Equal(t, "secondary service", retrieved2.name)
}

func TestContainer_GetNamed_NotFound(t *testing.T) {
	container := NewContainer()

	// try to get a named object that was never set
	retrieved, found := GetNamed[*containerTestService](container, "nonexistent")
	assert.False(t, found)
	assert.Nil(t, retrieved)
}

func TestContainer_SetNamed_Replace(t *testing.T) {
	container := NewContainer()

	// set initial named service
	service1 := &containerTestService{name: "first service"}
	container.SetNamed("test", service1)

	// replace with new service of same type and name
	service2 := &containerTestService{name: "second service"}
	container.SetNamed("test", service2)

	// should get the second service
	retrieved, found := GetNamed[*containerTestService](container, "test")
	assert.True(t, found)
	assert.Equal(t, service2, retrieved)
	assert.Equal(t, "second service", retrieved.name)
}

func TestContainer_Named_And_Singleton_Coexistence(t *testing.T) {
	container := NewContainer()

	// set singleton and named objects of same type
	singleton := &containerTestService{name: "singleton"}
	named1 := &containerTestService{name: "named1"}
	named2 := &containerTestService{name: "named2"}

	Set(container, singleton)
	container.SetNamed("first", named1)
	container.SetNamed("second", named2)

	// verify all can be retrieved independently
	retrievedSingleton, foundSingleton := Get[*containerTestService](container)
	assert.True(t, foundSingleton)
	assert.Equal(t, singleton, retrievedSingleton)

	retrievedNamed1, foundNamed1 := GetNamed[*containerTestService](container, "first")
	assert.True(t, foundNamed1)
	assert.Equal(t, named1, retrievedNamed1)

	retrievedNamed2, foundNamed2 := GetNamed[*containerTestService](container, "second")
	assert.True(t, foundNamed2)
	assert.Equal(t, named2, retrievedNamed2)
}

func TestContainer_OfType(t *testing.T) {
	container := NewContainer()

	// test OfType with empty container
	results := OfType[*containerTestService](container)
	assert.Empty(t, results)

	// add singleton only
	singleton := &containerTestService{name: "singleton"}
	Set(container, singleton)

	results = OfType[*containerTestService](container)
	assert.Len(t, results, 1)
	assert.Contains(t, results, singleton)

	// add named objects
	named1 := &containerTestService{name: "named1"}
	named2 := &containerTestService{name: "named2"}
	container.SetNamed("first", named1)
	container.SetNamed("second", named2)

	results = OfType[*containerTestService](container)
	assert.Len(t, results, 3)
	assert.Contains(t, results, singleton)
	assert.Contains(t, results, named1)
	assert.Contains(t, results, named2)

	// test with different type should be empty
	repoResults := OfType[*containerTestRepository](container)
	assert.Empty(t, repoResults)

	// add different type
	repo := &containerTestRepository{database: "test db"}
	Set(container, repo)

	repoResults = OfType[*containerTestRepository](container)
	assert.Len(t, repoResults, 1)
	assert.Contains(t, repoResults, repo)

	// original type results should be unchanged
	results = OfType[*containerTestService](container)
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

func TestContainer_AsType(t *testing.T) {
	container := NewContainer()

	// test AsType with empty container
	results := AsType[testInterface](container)
	assert.Empty(t, results)

	// add objects that implement the interface
	impl1 := &testImplementer1{value: "test1"}
	impl2 := &testImplementer2{number: 42}
	nonImpl := &containerTestService{name: "not an implementer"}

	Set(container, impl1)
	container.SetNamed("impl2", impl2)
	Set(container, nonImpl)

	// AsType should return only the implementers
	results = AsType[testInterface](container)
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
	concreteResults := AsType[*testImplementer1](container)
	assert.Len(t, concreteResults, 1)
	assert.Equal(t, impl1, concreteResults[0])
}

func TestContainer_Visit_WithNamedObjects(t *testing.T) {
	container := NewContainer()

	// add singleton and named objects
	singleton := &containerTestService{name: "singleton"}
	named1 := &containerTestService{name: "named1"}
	named2 := &containerTestRepository{database: "named db"}

	Set(container, singleton)
	container.SetNamed("service1", named1)
	container.SetNamed("repo1", named2)

	// collect all visited objects
	var visited []any
	err := container.Visit(func(object any) error {
		visited = append(visited, object)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, visited, 3)
	assert.Contains(t, visited, singleton)
	assert.Contains(t, visited, named1)
	assert.Contains(t, visited, named2)
}

func TestContainer_Clear_WithNamedObjects(t *testing.T) {
	container := NewContainer()

	// add singleton and named objects
	Set(container, &containerTestService{name: "singleton"})
	container.SetNamed("named", &containerTestService{name: "named"})

	// verify objects exist
	assert.True(t, Has[*containerTestService](container))
	_, found := GetNamed[*containerTestService](container, "named")
	assert.True(t, found)

	// clear container
	container.Clear()

	// verify all objects are gone
	assert.False(t, Has[*containerTestService](container))
	_, found = GetNamed[*containerTestService](container, "named")
	assert.False(t, found)
}

func TestContainer_Types_WithNamedObjects(t *testing.T) {
	container := NewContainer()

	// add singleton and named objects of same and different types
	Set(container, &containerTestService{name: "singleton"})
	container.SetNamed("named1", &containerTestService{name: "named1"})
	container.SetNamed("named2", &containerTestRepository{database: "named"})

	types := container.Types()

	// should have both types represented, but no duplicates
	assert.Len(t, types, 2)
	assert.Contains(t, types, reflect.TypeOf(&containerTestService{}))
	assert.Contains(t, types, reflect.TypeOf(&containerTestRepository{}))
}

func TestContainer_Inspect_Human(t *testing.T) {
	container := NewContainer()

	// add test objects
	service := &containerTestService{name: "test service"}
	repo := &containerTestRepository{database: "test db"}

	Set(container, service)
	container.SetNamed("primary", repo)
	Set(container, 42)

	output, err := container.Inspect(InspectHuman)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	fmt.Println(output)

	// verify human format uses the df.Inspect function
	// which should produce structured output similar to what we'd expect
	// from the InspectData structure
	assert.Contains(t, output, "containerTestService")
	assert.Contains(t, output, "containerTestRepository")
	assert.Contains(t, output, "test service")
	assert.Contains(t, output, "test db")
	assert.Contains(t, output, "42")

	// verify it contains summary information (lowercase field names from df.Inspect)
	assert.Contains(t, output, "summary")
	assert.Contains(t, output, "total")
	assert.Contains(t, output, "3")
	assert.Contains(t, output, "singletons")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "named")
	assert.Contains(t, output, "1")
}

func TestContainer_Inspect_JSON(t *testing.T) {
	container := NewContainer()

	// add test objects
	service := &containerTestService{name: "test service"}
	Set(container, service)
	container.SetNamed("primary", &containerTestRepository{database: "test db"})

	output, err := container.Inspect(InspectJSON)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// verify JSON structure
	assert.Contains(t, output, "\"summary\"")
	assert.Contains(t, output, "\"objects\"")
	assert.Contains(t, output, "\"total\": 2")
	assert.Contains(t, output, "\"singletons\": 1")
	assert.Contains(t, output, "\"named\": 1")
	assert.Contains(t, output, "\"storage\": \"singleton\"")
	assert.Contains(t, output, "\"storage\": \"named\"")
	assert.Contains(t, output, "\"name\": \"primary\"")

	// verify it's valid JSON
	var data InspectData
	err = json.Unmarshal([]byte(output), &data)
	assert.NoError(t, err)
	assert.Equal(t, 2, data.Summary.Total)
	assert.Equal(t, 1, data.Summary.Singletons)
	assert.Equal(t, 1, data.Summary.Named)
}

func TestContainer_Inspect_YAML(t *testing.T) {
	container := NewContainer()

	// add test objects
	Set(container, &containerTestService{name: "test service"})
	container.SetNamed("cache", &containerTestRepository{database: "cache db"})

	output, err := container.Inspect(InspectYAML)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// verify YAML structure
	assert.Contains(t, output, "summary:")
	assert.Contains(t, output, "objects:")
	assert.Contains(t, output, "total: 2")
	assert.Contains(t, output, "singletons: 1")
	assert.Contains(t, output, "named: 1")
	assert.Contains(t, output, "storage: singleton")
	assert.Contains(t, output, "storage: named")
	assert.Contains(t, output, "name: cache")

	// verify it's valid YAML
	var data InspectData
	err = yaml.Unmarshal([]byte(output), &data)
	assert.NoError(t, err)
	assert.Equal(t, 2, data.Summary.Total)
}

func TestContainer_Inspect_Empty(t *testing.T) {
	container := NewContainer()

	// test all formats with empty container
	humanOutput, err := container.Inspect(InspectHuman)
	assert.NoError(t, err)
	assert.Contains(t, humanOutput, "0")

	jsonOutput, err := container.Inspect(InspectJSON)
	assert.NoError(t, err)
	assert.Contains(t, jsonOutput, "\"total\": 0")

	yamlOutput, err := container.Inspect(InspectYAML)
	assert.NoError(t, err)
	assert.Contains(t, yamlOutput, "total: 0")
}

func TestContainer_Inspect_InvalidFormat(t *testing.T) {
	container := NewContainer()

	output, err := container.Inspect("invalid")
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Contains(t, err.Error(), "unsupported inspect format")
}
