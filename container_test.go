package df

import (
	"fmt"
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

// test types for builder pattern
type appConfig struct {
	DatabaseURL string `df:"database_url"`
	Port        int    `df:"port"`
}

type databaseService struct {
	url string
}

type webServer struct {
	port int
	db   *databaseService
}

func TestContainerBuilder_BasicUsage(t *testing.T) {
	// configuration data that would come from JSON/YAML
	configData := map[string]any{
		"database_url": "postgres://localhost/myapp",
		"port":         8080,
	}
	
	// build container using fluent API
	builder, err := NewBuilder().BindFrom(configData, &appConfig{})
	assert.NoError(t, err)
	container := builder.Build()
	
	// verify configuration was bound
	config, found := Get[*appConfig](container)
	assert.True(t, found)
	assert.Equal(t, "postgres://localhost/myapp", config.DatabaseURL)
	assert.Equal(t, 8080, config.Port)
}

func TestContainerBuilder_WithFactories(t *testing.T) {
	configData := map[string]any{
		"database_url": "postgres://localhost/myapp",
		"port":         8080,
	}
	
	builder := NewBuilder()
	
	// bind configuration
	builder, err := builder.BindFrom(configData, &appConfig{})
	assert.NoError(t, err)
	
	// register factories
	Factory(builder, func(c *Container) (*databaseService, error) {
		config, found := Get[*appConfig](c)
		if !found {
			return nil, fmt.Errorf("app config not found")
		}
		return &databaseService{url: config.DatabaseURL}, nil
	})
	
	Factory(builder, func(c *Container) (*webServer, error) {
		config, found := Get[*appConfig](c)
		if !found {
			return nil, fmt.Errorf("app config not found")
		}
		
		db, found := Get[*databaseService](c)
		if !found {
			return nil, fmt.Errorf("database service not found")
		}
		
		return &webServer{port: config.Port, db: db}, nil
	})
	
	// create all objects
	builder, err = builder.Create()
	assert.NoError(t, err)
	
	container := builder.Build()
	
	// verify objects were created
	db, found := Get[*databaseService](container)
	assert.True(t, found)
	assert.Equal(t, "postgres://localhost/myapp", db.url)
	
	server, found := Get[*webServer](container)
	assert.True(t, found)
	assert.Equal(t, 8080, server.port)
	assert.Equal(t, db, server.db)
}

func TestContainerBuilder_WithLinking(t *testing.T) {
	configData := map[string]any{
		"database_url": "postgres://localhost/myapp",
		"port":         8080,
	}
	
	// create objects without dependencies first
	dbService := &databaseService{}
	webSrv := &webServer{}
	
	builder := NewBuilder()
	builder, err := builder.BindFrom(configData, &appConfig{})
	assert.NoError(t, err)
	
	container, err := builder.
		Bind(dbService).
		Bind(webSrv).
		Link(func(c *Container) error {
			// wire database service
			config, found := Get[*appConfig](c)
			if !found {
				return fmt.Errorf("config not found")
			}
			
			db, found := Get[*databaseService](c)
			if !found {
				return fmt.Errorf("database service not found")
			}
			
			db.url = config.DatabaseURL
			return nil
		}).
		Link(func(c *Container) error {
			// wire web server
			config, found := Get[*appConfig](c)
			if !found {
				return fmt.Errorf("config not found")
			}
			
			server, found := Get[*webServer](c)
			if !found {
				return fmt.Errorf("web server not found")
			}
			
			db, found := Get[*databaseService](c)
			if !found {
				return fmt.Errorf("database service not found")
			}
			
			server.port = config.Port
			server.db = db
			return nil
		}).
		Wire()
	
	assert.NoError(t, err)
	
	finalContainer := container.Build()
	
	// verify linking worked
	db, found := Get[*databaseService](finalContainer)
	assert.True(t, found)
	assert.Equal(t, "postgres://localhost/myapp", db.url)
	
	server, found := Get[*webServer](finalContainer)
	assert.True(t, found)
	assert.Equal(t, 8080, server.port)
	assert.Equal(t, db, server.db)
}

func TestContainerBuilder_FullWorkflow(t *testing.T) {
	configData := map[string]any{
		"database_url": "postgres://localhost/myapp",
		"port":         8080,
	}
	
	builder := NewBuilder()
	
	// phase 1: bind configuration
	builder, err := builder.BindFrom(configData, &appConfig{})
	assert.NoError(t, err)
	
	// phase 2: register factories
	Factory(builder, func(c *Container) (*databaseService, error) {
		return &databaseService{}, nil // created empty, will be wired later
	})
	
	Factory(builder, func(c *Container) (*webServer, error) {
		return &webServer{}, nil // created empty, will be wired later
	})
	
	// phase 3: create objects
	builder, err = builder.Create()
	assert.NoError(t, err)
	
	// phase 4: register linkers
	builder.Link(func(c *Container) error {
		config, _ := Get[*appConfig](c)
		db, _ := Get[*databaseService](c)
		db.url = config.DatabaseURL
		return nil
	}).Link(func(c *Container) error {
		config, _ := Get[*appConfig](c)
		server, _ := Get[*webServer](c)
		db, _ := Get[*databaseService](c)
		
		server.port = config.Port
		server.db = db
		return nil
	})
	
	// phase 5: wire dependencies
	builder, err = builder.Wire()
	assert.NoError(t, err)
	
	// final container
	container := builder.Build()
	
	// verify everything is wired correctly
	config, found := Get[*appConfig](container)
	assert.True(t, found)
	assert.Equal(t, "postgres://localhost/myapp", config.DatabaseURL)
	
	db, found := Get[*databaseService](container)
	assert.True(t, found)
	assert.Equal(t, "postgres://localhost/myapp", db.url)
	
	server, found := Get[*webServer](container)
	assert.True(t, found)
	assert.Equal(t, 8080, server.port)
	assert.Equal(t, db, server.db)
}