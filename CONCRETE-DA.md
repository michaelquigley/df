# Reconceptualize `da` Framework - Concrete Container Approach

## Overview

Replace the reflection-heavy `da.Container` with a pattern where developers define their own **concrete container struct** with explicit types. The `da` package provides **lifecycle utilities** rather than a DI container.

## Design Decisions

- **Generic `Linkable[C]`** - Type-safe: components know their container type at compile time
- **Reflection traversal** - Automatic discovery of components via struct field traversal, controlled by tags
- **Functions only** - Minimal API: `LoadConfig()`, `Link()`, `Start()`, `Stop()`, `Run()`

---

## API Design

### 1. Core Interfaces

```go
package da

// Linkable receives the concrete container for wiring dependencies.
// Components implement this with their specific container type.
type Linkable[C any] interface {
    Link(c *C) error
}

// Startable components can be initialized after linking.
type Startable interface {
    Start() error
}

// Stoppable components can be cleaned up during shutdown.
type Stoppable interface {
    Stop() error
}
```

### 2. Configuration Loading

```go
package da

// Loader defines how configuration is loaded from a source.
type Loader interface {
    Load(dest any) error
}

// FileLoader loads config from JSON/YAML files (determined by extension).
// Returns error if file doesn't exist or can't be parsed.
func FileLoader(paths ...string) Loader

// OptionalFileLoader loads config from files, skipping missing ones silently.
func OptionalFileLoader(paths ...string) Loader

// ChainLoader combines multiple loaders, applying them in sequence.
func ChainLoader(loaders ...Loader) Loader

// LoadConfig populates a config struct using the provided loaders.
func LoadConfig[C any](cfg *C, loaders ...Loader) error
```

Note: EnvLoader for environment variables can be added later if needed.

### 3. Lifecycle Functions

```go
package da

// Link calls Link(c) on all components implementing Linkable[C].
// Traverses struct fields recursively, processing pointers, slices, and maps.
// Respects `da:"order=N"` tags for deterministic ordering.
// Fields with `da:"-"` are skipped.
func Link[C any](c *C) error

// Start calls Start() on all Startable components.
// Processes in order specified by `da:"order=N"` tags.
func Start[C any](c *C) error

// Stop calls Stop() on all Stoppable components.
// Processes in reverse order. Continues on error, returns first error.
func Stop[C any](c *C) error

// Run is a convenience that: Link -> Start -> wait for signal -> Stop.
func Run[C any](c *C) error

// WaitForSignal blocks until SIGINT or SIGTERM is received.
func WaitForSignal()
```

### 4. Struct Tags

```go
type App struct {
    Config   *Config    `da:"-"`           // Skip - not a component
    Database *Database  `da:"order=1"`     // Process first
    Cache    *Cache     `da:"order=2"`     // Process second

    Services struct {
        Auth  *AuthService  `da:"order=10"`
        Users *UserService  `da:"order=20"`
        API   *APIServer    `da:"order=100"` // Process last
    }
}
```

**Tag syntax:**
- `da:"-"` - Skip this field entirely
- `da:"order=N"` - Process in order N (lower first, default 0)

---

## Implementation

### File Structure

```
da/
├── interfaces.go    # Linkable[C], Startable, Stoppable interfaces
├── config.go        # Loader interface, FileLoader, EnvLoader, ChainLoader, LoadConfig
├── lifecycle.go     # Link, Start, Stop, Run, WaitForSignal functions
├── traverse.go      # Internal: field traversal, tag parsing, ordering
└── da_test.go       # Tests for new API
```

### Files to Delete

- `da/container.go` - Entire file (reflection-based container)
- `da/container_test.go` - Container tests

### Files to Rewrite

- `da/application.go` → becomes `da/interfaces.go` (keep interfaces, remove Application/Factory)

---

## Detailed Implementation Plan

### Step 1: Create `da/interfaces.go`

Keep and modify interfaces:

```go
package da

// Linkable receives the concrete container for wiring.
type Linkable[C any] interface {
    Link(c *C) error
}

// Startable defines components that require initialization.
type Startable interface {
    Start() error
}

// Stoppable defines components that require cleanup.
type Stoppable interface {
    Stop() error
}
```

### Step 2: Create `da/config.go`

Configuration loading with pluggable loaders:

```go
package da

import (
    "errors"
    "fmt"
    "path/filepath"

    "github.com/michaelquigley/df/dd"
)

// Loader defines how configuration is loaded from a source.
type Loader interface {
    Load(dest any) error
}

// fileLoader implements Loader for JSON/YAML files
type fileLoader struct {
    paths    []string
    optional bool
}

// FileLoader creates a loader for required config files.
func FileLoader(paths ...string) Loader {
    return &fileLoader{paths: paths, optional: false}
}

// OptionalFileLoader creates a loader that skips missing files.
func OptionalFileLoader(paths ...string) Loader {
    return &fileLoader{paths: paths, optional: true}
}

func (l *fileLoader) Load(dest any) error {
    for _, path := range l.paths {
        ext := filepath.Ext(path)
        var err error
        switch ext {
        case ".yaml", ".yml":
            err = dd.MergeYAMLFile(dest, path)
        case ".json":
            err = dd.MergeJSONFile(dest, path)
        default:
            return fmt.Errorf("unsupported config extension: %s", ext)
        }
        if err != nil {
            // Check if it's a not-found error and optional
            var fileErr *dd.FileError
            if l.optional && errors.As(err, &fileErr) && fileErr.IsNotFound() {
                continue
            }
            return err
        }
    }
    return nil
}

// chainLoader combines multiple loaders
type chainLoader struct {
    loaders []Loader
}

// ChainLoader creates a loader that applies multiple loaders in sequence.
func ChainLoader(loaders ...Loader) Loader {
    return &chainLoader{loaders: loaders}
}

func (c *chainLoader) Load(dest any) error {
    for _, l := range c.loaders {
        if err := l.Load(dest); err != nil {
            return err
        }
    }
    return nil
}

// LoadConfig populates a config struct using the provided loaders.
func LoadConfig[C any](cfg *C, loaders ...Loader) error {
    for _, l := range loaders {
        if err := l.Load(cfg); err != nil {
            return err
        }
    }
    return nil
}
```

### Step 3: Create `da/traverse.go`

Internal utilities for struct traversal:

```go
package da

import (
    "reflect"
    "sort"
    "strconv"
    "strings"
)

// component represents a discovered component with its order
type component struct {
    value reflect.Value
    order int
}

// traverse finds all pointer fields in a struct recursively
func traverse(v reflect.Value) []component {
    var components []component
    traverseRecursive(v, &components)
    sort.Slice(components, func(i, j int) bool {
        return components[i].order < components[j].order
    })
    return components
}

func traverseRecursive(v reflect.Value, components *[]component) {
    if v.Kind() == reflect.Ptr {
        if v.IsNil() {
            return
        }
        v = v.Elem()
    }
    if v.Kind() != reflect.Struct {
        return
    }

    t := v.Type()
    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        structField := t.Field(i)

        // Parse tag
        tag := structField.Tag.Get("da")
        if tag == "-" {
            continue
        }
        order := parseOrder(tag)

        // Handle different field types
        switch field.Kind() {
        case reflect.Ptr:
            if !field.IsNil() {
                *components = append(*components, component{value: field, order: order})
            }
        case reflect.Struct:
            // Recurse into embedded/nested structs
            traverseRecursive(field, components)
        case reflect.Slice:
            for j := 0; j < field.Len(); j++ {
                elem := field.Index(j)
                if elem.Kind() == reflect.Ptr && !elem.IsNil() {
                    *components = append(*components, component{value: elem, order: order})
                }
            }
        case reflect.Map:
            iter := field.MapRange()
            for iter.Next() {
                val := iter.Value()
                if val.Kind() == reflect.Ptr && !val.IsNil() {
                    *components = append(*components, component{value: val, order: order})
                }
            }
        }
    }
}

func parseOrder(tag string) int {
    for _, part := range strings.Split(tag, ",") {
        if strings.HasPrefix(part, "order=") {
            if n, err := strconv.Atoi(strings.TrimPrefix(part, "order=")); err == nil {
                return n
            }
        }
    }
    return 0
}
```

### Step 4: Create `da/lifecycle.go`

Main lifecycle functions:

```go
package da

import (
    "os"
    "os/signal"
    "reflect"
    "syscall"
)

// Link calls Link(c) on all Linkable[C] components
func Link[C any](c *C) error {
    v := reflect.ValueOf(c)
    components := traverse(v)

    for _, comp := range components {
        obj := comp.value.Interface()
        // Check if object implements Linkable[C]
        if linker, ok := obj.(Linkable[C]); ok {
            if err := linker.Link(c); err != nil {
                return err
            }
        }
    }
    return nil
}

// Start calls Start() on all Startable components
func Start[C any](c *C) error {
    v := reflect.ValueOf(c)
    components := traverse(v)

    for _, comp := range components {
        obj := comp.value.Interface()
        if starter, ok := obj.(Startable); ok {
            if err := starter.Start(); err != nil {
                return err
            }
        }
    }
    return nil
}

// Stop calls Stop() on all Stoppable components (reverse order)
func Stop[C any](c *C) error {
    v := reflect.ValueOf(c)
    components := traverse(v)

    // Reverse order for shutdown
    var firstErr error
    for i := len(components) - 1; i >= 0; i-- {
        obj := components[i].value.Interface()
        if stopper, ok := obj.(Stoppable); ok {
            if err := stopper.Stop(); err != nil && firstErr == nil {
                firstErr = err
            }
        }
    }
    return firstErr
}

// Run is Link -> Start -> wait -> Stop
func Run[C any](c *C) error {
    if err := Link(c); err != nil {
        return err
    }
    if err := Start(c); err != nil {
        Stop(c)
        return err
    }
    WaitForSignal()
    return Stop(c)
}

// WaitForSignal blocks until SIGINT or SIGTERM
func WaitForSignal() {
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
    <-ch
}
```

### Step 5: Delete Old Files

- Delete `da/container.go`
- Delete `da/container_test.go`
- Delete `da/application.go`
- Delete `da/application_test.go`

### Step 6: Create Tests

New test file `da/da_test.go` covering:
- Config loading from files
- Link with generic container type
- Start/Stop ordering
- Nested struct traversal
- Tag parsing

---

## Example Usage

```go
package main

import (
    "log"
    "github.com/michaelquigley/df/da"
)

// User-defined container with concrete types
type App struct {
    Config   *Config    `da:"-"`
    Database *Database  `da:"order=1"`
    Cache    *Cache     `da:"order=2"`

    Services struct {
        Auth  *AuthService  `da:"order=10"`
        Users *UserService  `da:"order=20"`
    }
}

type Config struct {
    DatabaseURL string `json:"database_url"`
    CacheURL    string `json:"cache_url"`
}

func main() {
    app := &App{Config: &Config{}}

    // Load configuration
    if err := da.LoadConfig(app.Config,
        da.FileLoader("config.yaml"),
        da.OptionalFileLoader("config.local.yaml"),
    ); err != nil {
        log.Fatal(err)
    }

    // Build components (explicit, no magic)
    app.Database = NewDatabase(app.Config.DatabaseURL)
    app.Cache = NewCache(app.Config.CacheURL)
    app.Services.Auth = NewAuthService()
    app.Services.Users = NewUserService()

    // Run application (link -> start -> wait -> stop)
    if err := da.Run(app); err != nil {
        log.Fatal(err)
    }
}

// Component implements Linkable[App] for type-safe wiring
type UserService struct {
    db    *Database
    cache *Cache
}

func (s *UserService) Link(app *App) error {
    s.db = app.Database
    s.cache = app.Cache
    return nil
}

func (s *UserService) Start() error {
    // Initialize...
    return nil
}

func (s *UserService) Stop() error {
    // Cleanup...
    return nil
}
```

---

## Verification

1. Write new tests in `da/da_test.go`
2. Create a simple example application
3. Verify config loading from JSON/YAML
4. Verify `Link()` passes correct container type to `Linkable[C]`
5. Verify `Start()`/`Stop()` respect ordering tags
6. Verify nested struct traversal works
7. Run `go test ./da/...`
