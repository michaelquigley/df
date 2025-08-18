package df

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.NotNil(t, container.objects)
}

func TestContainer_Set_And_Get(t *testing.T) {
	container := NewContainer()
	
	// test setting and getting a service
	service := &containerTestService{name: "test service"}
	container.Set(service)
	
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
	container.Set(service1)
	
	// replace with new service of same type
	service2 := &containerTestService{name: "second service"}
	container.Set(service2)
	
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
	
	container.Set(service)
	container.Set(repo)
	
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
	container.Set(42)
	container.Set("hello world")
	
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
	container.Set(service)
	
	retrieved, found := Get[interface{}](container)
	assert.False(t, found)
	assert.Nil(t, retrieved)
	
	// but this will work - storing and retrieving by the same interface type
	container.Set(writer)
	retrievedWriter, foundWriter := Get[interface {
		Write([]byte) (int, error)
	}](container)
	assert.False(t, foundWriter) // this will also be false because writer is nil
	assert.Nil(t, retrievedWriter)
}