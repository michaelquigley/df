---
title: Dependency Injection
description: Master the df container for object management, service discovery, and dependency injection patterns.
---

The df container provides a modern dependency injection system that manages object lifecycles, enables service discovery, and supports sophisticated dependency resolution patterns.

## Container Basics

The `Container` type is the central registry for objects in your application:

```go
import "github.com/michaelquigley/df"

// Create a new container
container := df.NewContainer()
```

## Object Registration

### Singleton Registration
Register objects by their type (one instance per type):

```go
type Database struct {
    URL string
    connected bool
}

// Register singleton
db := &Database{URL: "localhost:5432"}
container.Set(db)

// Later retrieve it
database, found := df.Get[*Database](container)
if found {
    fmt.Printf("Database URL: %s\n", database.URL)
}
```

### Named Registration
Register multiple instances of the same type:

```go
// Register named instances
primary := &Database{URL: "primary-db:5432"}
cache := &Database{URL: "cache-db:6379"}

container.SetNamed("primary", primary)
container.SetNamed("cache", cache)

// Retrieve by name
primaryDB, found := df.GetNamed[*Database](container, "primary")
cacheDB, found := df.GetNamed[*Database](container, "cache")
```

### Generic Registration
Use type parameters for cleaner registration:

```go
// Alternative registration syntax
df.SetAs[*Database](container, &Database{URL: "localhost:5432"})
df.SetNamedAs[*Database](container, "backup", &Database{URL: "backup:5432"})
```

## Service Discovery

### Type-Based Queries
Find objects by exact type:

```go
// Get singleton instance
db, found := df.Get[*Database](container)

// Get all instances of a type (singleton + all named instances)
allDatabases := df.OfType[*Database](container)
fmt.Printf("Found %d database instances\n", len(allDatabases))
```

### Interface-Based Queries
Find objects that implement specific interfaces:

```go
type Startable interface {
    Start() error
}

type Database struct {
    URL string
}

func (d *Database) Start() error {
    fmt.Printf("Starting database: %s\n", d.URL)
    return nil
}

// Register services
container.Set(&Database{URL: "localhost:5432"})

// Find all objects implementing Startable
startables := df.AsType[Startable](container)
for _, service := range startables {
    service.Start()
}
```

### Named Queries
Work with named instances:

```go
// Check if named instance exists
if container.HasNamed("cache") {
    cache, _ := df.GetNamed[*Database](container, "cache")
    // Use cache database
}

// Get all names for a type
names := df.NamesOfType[*Database](container)
for _, name := range names {
    db, _ := df.GetNamed[*Database](container, name)
    fmt.Printf("Database %s: %s\n", name, db.URL)
}
```

## Dependency Patterns

### Constructor Injection
Use factories to inject dependencies during construction:

```go
type UserService struct {
    db *Database
    logger *Logger
}

func NewUserService(db *Database, logger *Logger) *UserService {
    return &UserService{
        db: db,
        logger: logger,
    }
}

// In your factory
type UserServiceFactory struct{}

func (f *UserServiceFactory) Build(app *df.Application[Config]) error {
    db, _ := df.Get[*Database](app.C)
    logger, _ := df.Get[*Logger](app.C)
    
    service := NewUserService(db, logger)
    df.SetAs[*UserService](app.C, service)
    
    return nil
}
```

### Setter Injection
Use the `Linkable` interface for dependency injection after construction:

```go
type OrderService struct {
    db *Database
    userService *UserService
    logger *Logger
}

// Linkable allows dependency injection after object creation
func (s *OrderService) Link(c *df.Container) error {
    var found bool
    
    s.db, found = df.Get[*Database](c)
    if !found {
        return errors.New("database not found")
    }
    
    s.userService, found = df.Get[*UserService](c)
    if !found {
        return errors.New("user service not found")
    }
    
    s.logger, found = df.Get[*Logger](c)
    if !found {
        return errors.New("logger not found")
    }
    
    return nil
}

// Register the service
container.Set(&OrderService{})

// Link dependencies (usually done by Application)
orderService, _ := df.Get[*OrderService](container)
err := orderService.Link(container)
```

### Interface Segregation
Use multiple small interfaces for flexible dependency management:

```go
type DatabaseReader interface {
    Query(query string) ([]Row, error)
}

type DatabaseWriter interface {
    Execute(query string) error
}

type DatabaseReaderWriter interface {
    DatabaseReader
    DatabaseWriter
}

type Database struct {
    URL string
}

func (d *Database) Query(query string) ([]Row, error) { /* ... */ }
func (d *Database) Execute(query string) error { /* ... */ }

// Services can depend on specific interfaces
type ReportService struct {
    db DatabaseReader // Only needs read access
}

type UserService struct {
    db DatabaseReaderWriter // Needs full access
}

// Register with interface types
container.Set(&Database{URL: "localhost:5432"})

// Services find what they need
db, _ := df.Get[*Database](container)
reportDB := DatabaseReader(db)
userDB := DatabaseReaderWriter(db)
```

## Container Introspection

### Human-Readable Output
Inspect container contents for debugging:

```go
container.Set(&Database{URL: "localhost:5432"})
container.SetNamed("cache", &Database{URL: "cache:6379"})
container.Set(&Logger{Level: "info"})

output, err := container.Inspect(df.InspectHuman)
fmt.Println(output)
```

Output:
```
Container Contents:
  *Database (singleton): &{localhost:5432}
  *Database["cache"]: &{cache:6379}
  *Logger (singleton): &{info}
```

### Machine-Readable Formats
Export container state as JSON or YAML:

```go
// JSON format
jsonOutput, err := container.Inspect(df.InspectJSON)

// YAML format  
yamlOutput, err := container.Inspect(df.InspectYAML)
```

## Advanced Patterns

### Service Locator Pattern
Create a centralized service locator:

```go
type ServiceLocator struct {
    container *df.Container
}

func NewServiceLocator() *ServiceLocator {
    return &ServiceLocator{
        container: df.NewContainer(),
    }
}

func (sl *ServiceLocator) GetDatabase() *Database {
    db, _ := df.Get[*Database](sl.container)
    return db
}

func (sl *ServiceLocator) GetLogger() *Logger {
    logger, _ := df.Get[*Logger](sl.container)
    return logger
}

func (sl *ServiceLocator) GetUserService() *UserService {
    service, _ := df.Get[*UserService](sl.container)
    return service
}
```

### Plugin Architecture
Use the container for dynamic plugin loading:

```go
type Plugin interface {
    Name() string
    Initialize() error
    Process(data any) (any, error)
}

type PluginManager struct {
    container *df.Container
}

func (pm *PluginManager) LoadPlugin(name string, plugin Plugin) {
    pm.container.SetNamed(name, plugin)
}

func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
    return df.GetNamed[Plugin](pm.container, name)
}

func (pm *PluginManager) GetAllPlugins() []Plugin {
    return df.AsType[Plugin](pm.container)
}

func (pm *PluginManager) ProcessWithAllPlugins(data any) {
    plugins := pm.GetAllPlugins()
    for _, plugin := range plugins {
        result, err := plugin.Process(data)
        if err != nil {
            fmt.Printf("Plugin %s failed: %v\n", plugin.Name(), err)
        } else {
            fmt.Printf("Plugin %s result: %v\n", plugin.Name(), result)
        }
    }
}
```

### Conditional Registration
Register services based on configuration:

```go
type Config struct {
    DatabaseType string `df:"database_type"`
    CacheEnabled bool   `df:"cache_enabled"`
    LogLevel     string `df:"log_level"`
}

func RegisterServices(container *df.Container, config Config) error {
    // Register database based on type
    switch config.DatabaseType {
    case "postgres":
        container.Set(&PostgresDatabase{})
    case "mysql":
        container.Set(&MySQLDatabase{})
    default:
        container.Set(&InMemoryDatabase{})
    }
    
    // Conditionally register cache
    if config.CacheEnabled {
        container.Set(&RedisCache{})
    }
    
    // Register logger with appropriate level
    logger := &Logger{Level: config.LogLevel}
    container.Set(logger)
    
    return nil
}
```

## Best Practices

### Interface-First Design
Design with interfaces to maximize flexibility:

```go
// Good - depend on interfaces
type UserService struct {
    db     DatabaseInterface
    cache  CacheInterface
    logger LoggerInterface
}

// Less flexible - depend on concrete types
type UserService struct {
    db     *PostgresDatabase
    cache  *RedisCache
    logger *FileLogger
}
```

### Single Responsibility
Keep services focused on single concerns:

```go
// Good - focused responsibilities
type UserRepository struct {
    db DatabaseReader
}

type UserValidator struct {
    rules []ValidationRule
}

type UserService struct {
    repo      *UserRepository
    validator *UserValidator
    logger    Logger
}

// Poor - too many responsibilities
type UserService struct {
    db            *Database
    validationRules []Rule
    emailSender     *EmailSender
    logger          *Logger
    cache          *Cache
    // ... many more dependencies
}
```

### Lifecycle Management
Use lifecycle interfaces consistently:

```go
type Service struct {
    db     *Database
    logger *Logger
}

func (s *Service) Link(c *df.Container) error {
    // Resolve dependencies
    s.db, _ = df.Get[*Database](c)
    s.logger, _ = df.Get[*Logger](c)
    return nil
}

func (s *Service) Start() error {
    s.logger.Info("service starting")
    return s.db.Connect()
}

func (s *Service) Stop() error {
    s.logger.Info("service stopping")
    return s.db.Disconnect()
}
```

### Error Handling
Handle missing dependencies gracefully:

```go
func (s *OrderService) Link(c *df.Container) error {
    var err error
    
    s.db, found := df.Get[*Database](c)
    if !found {
        return fmt.Errorf("required dependency Database not found")
    }
    
    s.logger, found = df.Get[*Logger](c)
    if !found {
        // Use default logger if not provided
        s.logger = &DefaultLogger{}
    }
    
    return nil
}
```

## Testing with Containers

### Mock Dependencies
Replace real services with mocks for testing:

```go
func TestUserService(t *testing.T) {
    container := df.NewContainer()
    
    // Register mocks
    mockDB := &MockDatabase{}
    mockLogger := &MockLogger{}
    
    container.Set(mockDB)
    container.Set(mockLogger)
    
    // Create and test service
    service := &UserService{}
    service.Link(container)
    
    // Test service behavior
    result, err := service.GetUser("123")
    assert.NoError(t, err)
    assert.Equal(t, "John Doe", result.Name)
}
```

## Next Steps

Now that you understand dependency injection, learn about:

- **[Application Lifecycle](/guides/application-lifecycle/)** - Orchestrate complex applications with factories and lifecycle management
- **[Advanced Features](/guides/advanced-features/)** - Explore Dynamic types, object references, and custom converters
- **[Getting Started](/guides/getting-started/)** - Review the basics if needed