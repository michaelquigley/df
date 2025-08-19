# df_13_container - dependency injection and application management

this example demonstrates the `Container` and `Application` types for building scalable application containers with dependency injection, lifecycle management, and application discovery.

## what you'll learn

- **dependency injection**: automatic object creation and registration through factories
- **application lifecycle**: initialize → start → use → stop pattern
- **object registration**: singleton and named object storage
- **type queries**: find objects by exact type or interface compatibility
- **container introspection**: inspect container contents for debugging

## key concepts

### container
the `Container` is an application container that manages objects:

```go
// singleton objects (one per type)
container.Set(database)
logger, found := df.Get[*Logger](container)

// named objects (multiple per type)
container.SetNamed("primary", primaryDB)
container.SetNamed("cache", cacheDB)

// type queries
allDatabases := df.OfType[*Database](container)     // exact type matches
allStartables := df.AsType[df.Startable](container) // interface matches
```

### application
the `Application` orchestrates object creation and lifecycle:

```go
// create application with configuration
app := df.NewApplication(config)
df.WithFactory(app, &DatabaseFactory{})

// initialize: build + link dependencies
app.Initialize()

// start all startable objects
app.Start()

// clean shutdown
app.Stop()
```

### factories
factories create and register objects in the container:

```go
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(s *df.Application[Config]) error {
    cfg, _ := df.Get[Config](s.R)
    
    db := &Database{URL: cfg.DatabaseURL}
    df.SetAs[*Database](s.R, db)
    return nil
}
```

### lifecycle interfaces
objects can implement lifecycle interfaces for automatic management:

```go
type Database struct {
    URL string
    Connected bool
}

// df.Startable - called during application.Start()
func (d *Database) Start() error {
    return d.Connect()
}

// df.Stoppable - called during application.Stop()
func (d *Database) Stop() error {
    d.Connected = false
    return nil
}
```

## running the example

```bash
cd df_13_container
go run main.go
```

## example output

```
connecting to database: postgres://localhost:5432/mydb
starting logger with level: info

=== container contents ===
InspectData {
  summary       : InspectSummary {
    total       : 3
    singletons  : 3
    named       : 0
  }
  objects       : [
    [0]: InspectObject {
      type      : "main.Config"
      storage   : "singleton"
      name      : <nil>
      value     : "{AppName:example app DatabaseURL:postgres://localhost:5432/mydb LogLevel:info}"
    }
    [1]: InspectObject {
      type      : "*main.Database"
      storage   : "singleton"
      name      : <nil>
      value     : "&{URL:postgres://localhost:5432/mydb Connected:true}"
    }
    [2]: InspectObject {
      type      : "*main.Logger"
      storage   : "singleton"
      name      : <nil>
      value     : "&{Level:info}"
    }
  ]
}
[INFO] application started successfully
database connected: true

found 2 loggers:
  [0] level: info
  [1] level: debug

found 4 startable applications
```

## real-world applications

this pattern is essential for:

- **microapplications**: manage databases, message queues, http servers
- **plugin architectures**: dynamically load and manage plugin instances
- **testing**: easily mock dependencies by registering test doubles
- **configuration**: different environments can register different implementations
- **monitoring**: inspect container state for debugging and health checks

## best practices

1. **use factories**: keep object creation logic separate and testable
2. **implement lifecycle interfaces**: enable clean startup and shutdown
3. **leverage type queries**: find related applications (all caches, all loggers)
4. **use named objects**: multiple instances of the same type
5. **inspect for debugging**: use `container.Inspect()` to understand container state

## next steps

- integrate with your application's configuration loading
- implement health check endpoints using container introspection
- build plugin systems with dynamic object registration
- create test utilities that mock dependencies in the container