package da

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
	assert.NotNil(t, container.taggedObjects)
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

	SetNamed(container, "primary", service1)
	SetNamed(container, "secondary", service2)

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
	SetNamed(container, "test", service1)

	// replace with new service of same type and name
	service2 := &containerTestService{name: "second service"}
	SetNamed(container, "test", service2)

	// should get the second service
	retrieved, found := GetNamed[*containerTestService](container, "test")
	assert.True(t, found)
	assert.Equal(t, service2, retrieved)
	assert.Equal(t, "second service", retrieved.name)
}

func TestContainer_HasNamed(t *testing.T) {
	container := NewContainer()

	// test HasNamed with empty container
	assert.False(t, HasNamed[*containerTestService](container, "nonexistent"))

	// set named objects
	service := &containerTestService{name: "test service"}
	repo := &containerTestRepository{database: "test db"}
	SetNamed(container, "service", service)
	SetNamed(container, "repo", repo)

	// test HasNamed for existing objects
	assert.True(t, HasNamed[*containerTestService](container, "service"))
	assert.True(t, HasNamed[*containerTestRepository](container, "repo"))

	// test HasNamed for non-existing objects
	assert.False(t, HasNamed[*containerTestService](container, "repo"))        // wrong type
	assert.False(t, HasNamed[*containerTestRepository](container, "service"))  // wrong type
	assert.False(t, HasNamed[*containerTestService](container, "nonexistent")) // wrong name
	assert.False(t, HasNamed[int](container, "service"))                       // completely different type
}

func TestContainer_RemoveNamed(t *testing.T) {
	container := NewContainer()

	// test RemoveNamed from empty container
	removed := RemoveNamed[*containerTestService](container, "nonexistent")
	assert.False(t, removed)

	// set named objects
	service := &containerTestService{name: "test service"}
	repo := &containerTestRepository{database: "test db"}
	SetNamed(container, "service", service)
	SetNamed(container, "repo", repo)

	// verify objects exist
	assert.True(t, HasNamed[*containerTestService](container, "service"))
	assert.True(t, HasNamed[*containerTestRepository](container, "repo"))

	// remove service
	removed = RemoveNamed[*containerTestService](container, "service")
	assert.True(t, removed)
	assert.False(t, HasNamed[*containerTestService](container, "service"))

	// repo should still exist
	assert.True(t, HasNamed[*containerTestRepository](container, "repo"))

	// try to remove service again
	removed = RemoveNamed[*containerTestService](container, "service")
	assert.False(t, removed)

	// try to remove with wrong type
	removed = RemoveNamed[*containerTestService](container, "repo")
	assert.False(t, removed)
	assert.True(t, HasNamed[*containerTestRepository](container, "repo")) // should still exist

	// remove repo with correct type
	removed = RemoveNamed[*containerTestRepository](container, "repo")
	assert.True(t, removed)
	assert.False(t, HasNamed[*containerTestRepository](container, "repo"))
}

func TestContainer_SemanticCompatibility_Has_vs_HasNamed(t *testing.T) {
	container := NewContainer()

	// test 1: both should return false for non-existent objects
	assert.False(t, Has[string](container))
	assert.False(t, HasNamed[string](container, "test"))

	// test 2: set singleton, verify Has returns true but HasNamed remains false
	Set(container, "singleton value")
	assert.True(t, Has[string](container))
	assert.False(t, HasNamed[string](container, "test"))

	// test 3: set named object, verify HasNamed returns true but Has still works for singleton
	SetNamed(container, "test", "named value")
	assert.True(t, Has[string](container))
	assert.True(t, HasNamed[string](container, "test"))

	// test 4: different names should not interfere
	assert.True(t, HasNamed[string](container, "test"))
	assert.False(t, HasNamed[string](container, "different"))

	// test 5: type safety should work consistently
	assert.False(t, Has[int](container))
	assert.False(t, HasNamed[int](container, "test"))
}

func TestContainer_SemanticCompatibility_Remove_vs_RemoveNamed(t *testing.T) {
	container := NewContainer()

	// test 1: both should return false for non-existent objects
	assert.False(t, Remove[string](container))
	assert.False(t, RemoveNamed[string](container, "test"))

	// test 2: set objects and verify removal behavior
	Set(container, "singleton")
	SetNamed(container, "test", "named")

	// verify objects exist
	assert.True(t, Has[string](container))
	assert.True(t, HasNamed[string](container, "test"))

	// test 3: remove singleton, named should remain
	removed := Remove[string](container)
	assert.True(t, removed)
	assert.False(t, Has[string](container))
	assert.True(t, HasNamed[string](container, "test"))

	// test 4: attempt to remove singleton again should return false
	removed = Remove[string](container)
	assert.False(t, removed)

	// test 5: remove named object
	removed = RemoveNamed[string](container, "test")
	assert.True(t, removed)
	assert.False(t, HasNamed[string](container, "test"))

	// test 6: attempt to remove named again should return false
	removed = RemoveNamed[string](container, "test")
	assert.False(t, removed)

	// test 7: removing with wrong name should return false
	SetNamed(container, "correct", "value")
	assert.False(t, RemoveNamed[string](container, "wrong"))
	assert.True(t, HasNamed[string](container, "correct")) // should still exist
}

func TestContainer_SemanticCompatibility_CrossOperations(t *testing.T) {
	container := NewContainer()

	// set both singleton and named objects of same type
	Set(container, "singleton")
	SetNamed(container, "named", "named value")

	// test 1: Has/Remove work only on singletons, not named objects
	assert.True(t, Has[string](container))
	assert.True(t, Remove[string](container))
	assert.False(t, Has[string](container))
	assert.True(t, HasNamed[string](container, "named")) // named should remain

	// test 2: HasNamed/RemoveNamed work only on named objects, not singletons
	Set(container, "singleton again")
	assert.True(t, Has[string](container))
	assert.True(t, HasNamed[string](container, "named"))

	assert.True(t, RemoveNamed[string](container, "named"))
	assert.False(t, HasNamed[string](container, "named"))
	assert.True(t, Has[string](container)) // singleton should remain

	// test 3: type safety works consistently across operations
	Remove[string](container) // remove the string singleton from test 2
	Set(container, 42)
	SetNamed(container, "int", 100)

	assert.True(t, Has[int](container))
	assert.True(t, HasNamed[int](container, "int"))
	assert.False(t, Has[string](container))             // removed earlier
	assert.False(t, HasNamed[string](container, "int")) // wrong type

	assert.True(t, Remove[int](container))
	assert.True(t, RemoveNamed[int](container, "int"))
}

func TestContainer_Named_And_Singleton_Coexistence(t *testing.T) {
	container := NewContainer()

	// set singleton and named objects of same type
	singleton := &containerTestService{name: "singleton"}
	named1 := &containerTestService{name: "named1"}
	named2 := &containerTestService{name: "named2"}

	Set(container, singleton)
	SetNamed(container, "first", named1)
	SetNamed(container, "second", named2)

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
	SetNamed(container, "first", named1)
	SetNamed(container, "second", named2)

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
	SetNamed(container, "impl2", impl2)
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
	SetNamed(container, "service1", named1)
	SetNamed(container, "repo1", named2)

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
	SetNamed(container, "named", &containerTestService{name: "named"})

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
	SetNamed(container, "named1", &containerTestService{name: "named1"})
	SetNamed(container, "named2", &containerTestRepository{database: "named"})

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
	SetNamed(container, "primary", repo)
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
	SetNamed(container, "primary", &containerTestRepository{database: "test db"})

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
	SetNamed(container, "cache", &containerTestRepository{database: "cache db"})

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

// Tagged object tests

func TestContainer_AddTagged_And_Tagged(t *testing.T) {
	container := NewContainer()

	// add objects with a tag
	service1 := &containerTestService{name: "service1"}
	service2 := &containerTestService{name: "service2"}
	AddTagged(container, "http-handlers", service1)
	AddTagged(container, "http-handlers", service2)

	// retrieve all objects with the tag
	objects := Tagged(container, "http-handlers")
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, service1)
	assert.Contains(t, objects, service2)
}

func TestContainer_Tagged_NotFound(t *testing.T) {
	container := NewContainer()

	// try to get objects from a non-existent tag
	objects := Tagged(container, "non-existent")
	assert.Nil(t, objects)
}

func TestContainer_TaggedOfType(t *testing.T) {
	container := NewContainer()

	// add objects of different types with the same tag
	service := &containerTestService{name: "service"}
	repo := &containerTestRepository{database: "db"}
	AddTagged(container, "components", service)
	AddTagged(container, "components", repo)

	// retrieve only services
	services := TaggedOfType[*containerTestService](container, "components")
	assert.Len(t, services, 1)
	assert.Equal(t, service, services[0])

	// retrieve only repositories
	repos := TaggedOfType[*containerTestRepository](container, "components")
	assert.Len(t, repos, 1)
	assert.Equal(t, repo, repos[0])
}

func TestContainer_TaggedAsType(t *testing.T) {
	container := NewContainer()

	// add objects implementing an interface
	service := &containerTestService{name: "service"}
	repo := &containerTestRepository{database: "db"}
	AddTagged(container, "all", service)
	AddTagged(container, "all", repo)

	// use fmt.Stringer as a test - neither implements it, so should return empty
	stringers := TaggedAsType[fmt.Stringer](container, "all")
	assert.Len(t, stringers, 0)

	// retrieve all as any
	all := Tagged(container, "all")
	assert.Len(t, all, 2)
}

func TestContainer_HasTagged(t *testing.T) {
	container := NewContainer()

	// should not have any tags initially
	assert.False(t, HasTagged(container, "test-tag"))

	// add an object with a tag
	service := &containerTestService{name: "service"}
	AddTagged(container, "test-tag", service)

	// should have the tag now
	assert.True(t, HasTagged(container, "test-tag"))
	assert.False(t, HasTagged(container, "other-tag"))
}

func TestContainer_RemoveTaggedFrom(t *testing.T) {
	container := NewContainer()

	service1 := &containerTestService{name: "service1"}
	service2 := &containerTestService{name: "service2"}
	AddTagged(container, "handlers", service1)
	AddTagged(container, "handlers", service2)

	// remove one object
	removed := RemoveTaggedFrom(container, "handlers", service1)
	assert.True(t, removed)

	// only service2 should remain
	objects := Tagged(container, "handlers")
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, service2)

	// try to remove again - should return false
	removed = RemoveTaggedFrom(container, "handlers", service1)
	assert.False(t, removed)
}

func TestContainer_RemoveTaggedFrom_CleansUpEmptyTags(t *testing.T) {
	container := NewContainer()

	service := &containerTestService{name: "service"}
	AddTagged(container, "handlers", service)

	// remove the only object
	RemoveTaggedFrom(container, "handlers", service)

	// tag should be removed entirely
	assert.False(t, HasTagged(container, "handlers"))
	assert.Len(t, container.Tags(), 0)
}

func TestContainer_RemoveTagged(t *testing.T) {
	container := NewContainer()

	service := &containerTestService{name: "service"}
	AddTagged(container, "tag1", service)
	AddTagged(container, "tag2", service)
	AddTagged(container, "tag3", service)

	// remove from all tags
	count := RemoveTagged(container, service)
	assert.Equal(t, 3, count)

	// all tags should be gone
	assert.False(t, HasTagged(container, "tag1"))
	assert.False(t, HasTagged(container, "tag2"))
	assert.False(t, HasTagged(container, "tag3"))
}

func TestContainer_ClearTagged(t *testing.T) {
	container := NewContainer()

	service1 := &containerTestService{name: "service1"}
	service2 := &containerTestService{name: "service2"}
	AddTagged(container, "handlers", service1)
	AddTagged(container, "handlers", service2)

	// clear all objects with the tag
	count := ClearTagged(container, "handlers")
	assert.Equal(t, 2, count)

	// tag should be gone
	assert.False(t, HasTagged(container, "handlers"))
}

func TestContainer_Tags(t *testing.T) {
	container := NewContainer()

	// initially empty
	assert.Len(t, container.Tags(), 0)

	// add objects with different tags
	service := &containerTestService{name: "service"}
	AddTagged(container, "handlers", service)
	AddTagged(container, "middleware", service)
	AddTagged(container, "filters", service)

	tags := container.Tags()
	assert.Len(t, tags, 3)
	assert.Contains(t, tags, "handlers")
	assert.Contains(t, tags, "middleware")
	assert.Contains(t, tags, "filters")
}

func TestContainer_MultipleTagsPerObject(t *testing.T) {
	container := NewContainer()

	service := &containerTestService{name: "service"}
	AddTagged(container, "tag1", service)
	AddTagged(container, "tag2", service)

	// should appear in both tags
	objects1 := Tagged(container, "tag1")
	assert.Len(t, objects1, 1)
	assert.Equal(t, service, objects1[0])

	objects2 := Tagged(container, "tag2")
	assert.Len(t, objects2, 1)
	assert.Equal(t, service, objects2[0])
}

func TestContainer_Visit_WithTaggedObjects(t *testing.T) {
	container := NewContainer()

	// add singleton, named, and tagged objects
	singleton := &containerTestService{name: "singleton"}
	named := &containerTestService{name: "named"}
	tagged := &containerTestService{name: "tagged"}
	Set(container, singleton)
	SetNamed(container, "named", named)
	AddTagged(container, "tag", tagged)

	var visited []*containerTestService
	err := container.Visit(func(obj any) error {
		if s, ok := obj.(*containerTestService); ok {
			visited = append(visited, s)
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, visited, 3)
	assert.Contains(t, visited, singleton)
	assert.Contains(t, visited, named)
	assert.Contains(t, visited, tagged)
}

func TestContainer_Visit_DeduplicatesTaggedObjects(t *testing.T) {
	container := NewContainer()

	// add same object to multiple tags
	service := &containerTestService{name: "service"}
	AddTagged(container, "tag1", service)
	AddTagged(container, "tag2", service)
	AddTagged(container, "tag3", service)

	visitCount := 0
	err := container.Visit(func(obj any) error {
		visitCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, visitCount) // should only visit once
}

func TestContainer_Visit_DeduplicatesSingletonAndTagged(t *testing.T) {
	container := NewContainer()

	// add same object as singleton and tagged
	service := &containerTestService{name: "service"}
	Set(container, service)
	AddTagged(container, "tag", service)

	visitCount := 0
	err := container.Visit(func(obj any) error {
		visitCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, visitCount) // should only visit once
}

func TestContainer_Clear_IncludesTagged(t *testing.T) {
	container := NewContainer()

	service := &containerTestService{name: "service"}
	Set(container, &containerTestRepository{database: "db"})
	SetNamed(container, "named", service)
	AddTagged(container, "tag", service)

	container.Clear()

	assert.Len(t, container.Tags(), 0)
	assert.False(t, HasTagged(container, "tag"))
	assert.False(t, Has[*containerTestRepository](container))
	assert.False(t, HasNamed[*containerTestService](container, "named"))
}

func TestContainer_OfType_IncludesTagged(t *testing.T) {
	container := NewContainer()

	singleton := &containerTestService{name: "singleton"}
	named := &containerTestService{name: "named"}
	tagged := &containerTestService{name: "tagged"}

	Set(container, singleton)
	SetNamed(container, "named", named)
	AddTagged(container, "tag", tagged)

	all := OfType[*containerTestService](container)
	assert.Len(t, all, 3)
	assert.Contains(t, all, singleton)
	assert.Contains(t, all, named)
	assert.Contains(t, all, tagged)
}

func TestContainer_OfType_DeduplicatesTagged(t *testing.T) {
	container := NewContainer()

	// add same object as singleton and tagged
	service := &containerTestService{name: "service"}
	Set(container, service)
	AddTagged(container, "tag1", service)
	AddTagged(container, "tag2", service)

	all := OfType[*containerTestService](container)
	assert.Len(t, all, 1) // should only appear once
	assert.Equal(t, service, all[0])
}

func TestContainer_Inspect_WithTagged(t *testing.T) {
	container := NewContainer()

	Set(container, &containerTestService{name: "singleton"})
	SetNamed(container, "named", &containerTestRepository{database: "db"})
	AddTagged(container, "handlers", &containerTestService{name: "tagged"})

	output, err := container.Inspect(InspectJSON)
	assert.NoError(t, err)

	var data InspectData
	err = json.Unmarshal([]byte(output), &data)
	assert.NoError(t, err)

	assert.Equal(t, 3, data.Summary.Total)
	assert.Equal(t, 1, data.Summary.Singletons)
	assert.Equal(t, 1, data.Summary.Named)
	assert.Equal(t, 1, data.Summary.Tagged)

	// verify tagged object is in the list
	foundTagged := false
	for _, obj := range data.Objects {
		if obj.Storage == "tagged" {
			foundTagged = true
			assert.NotNil(t, obj.Tag)
			assert.Equal(t, "handlers", *obj.Tag)
		}
	}
	assert.True(t, foundTagged)
}
