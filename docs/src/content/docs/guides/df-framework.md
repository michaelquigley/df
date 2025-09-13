---
title: dynamic foundation
description: Complete guide for using the df framework as a unified system - combining dd, dl, and da packages.
---

**Building dynamic, configuration-driven applications**

Applications often need different behavior based on environment, deployment target, or runtime configuration. This typically requires loading settings from files, routing logs to appropriate destinations, and managing object dependencies across different components... all of the common scenarios when building large monoliths at scale.


`df` is comprised of three foundational packages. The three packages work independently but complement each other well. 

* `dd` (_dynamic foundation for data_) handles converting between Go structs and data formats like JSON or YAML. It provides foundational components for building all manager of flexible data structures for configuration, application data, as well as data in motion.

* `dl` (_dynamic foundation for logging_) routes different types of logs to different outputs with independent formatting. It's designed to provide a solid foundation for getting coherent trace information out of your applications.

* `da` (_dynamic foundation for applications_) manages object creation and lifecycle through factories and dependency injection. It's a small, fast, extensible application container that provides clear abstractions for managing your components.

This guide demonstrates integration patterns where configuration drives logging setup, factory behavior, and application structure. Each section shows progressively more complex scenarios, from basic configuration loading to full plugin architectures with dynamic reconfiguration.

## Framework Overview

The `df` framework consists of three complementary packages:

- **`dd`** - Dynamic Data: Struct ↔ map conversion and configuration loading
- **`dl`** - Dynamic Logging: Channel-based logging with intelligent routing  
- **`da`** - Dynamic Application: Dependency injection and lifecycle management

## Progressive Integration Patterns

### 1. Configuration-Driven Application - Basic

**Load configuration, set up logging, manage lifecycle**

```go
package main

import (
    "os"
    "github.com/michaelquigley/df/dd"
    "github.com/michaelquigley/df/dl"
    "github.com/michaelquigley/df/da"
)

// Configuration structure with all framework settings
type AppConfig struct {
    // Application settings
    AppName string `dd:"app_name"`
    Port    int    `dd:"port"`
    
    // Database settings  
    Database DatabaseConfig `dd:"database"`
    
    // Logging settings
    Logging LoggingConfig `dd:"logging"`
}

type DatabaseConfig struct {
    URL      string `dd:"url"`
    PoolSize int    `dd:"pool_size"`
}

type LoggingConfig struct {
    Level     string `dd:"level"`
    UseJSON   bool   `dd:"use_json"`
    OutputDir string `dd:"output_dir"`
}

func main() {
    // 1. Load configuration with dd
    config, err := dd.BindFromYAML[AppConfig]("config.yaml")
    if err != nil {
        panic(err)
    }
    
    // 2. Setup logging with dl
    setupLogging(config.Logging)
    
    // 3. Create application with da
    app := da.NewApplication(*config)
    da.WithFactory(app, &DatabaseFactory{})
    da.WithFactory(app, &APIServerFactory{})
    
    // 4. Initialize and start
    app.Initialize()
    app.Start()
    
    dl.Log().With("app", config.AppName).Info("application started")
    
    // Keep running...
    select {}
}

func setupLogging(config LoggingConfig) {
    // Configure logging channels based on config
    if config.UseJSON {
        dl.ConfigureChannel("api", dl.DefaultOptions().JSON())
    }
    
    if config.OutputDir != "" {
        dbFile, _ := os.Create(config.OutputDir + "/database.log")
        dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))
    }
}
```

### 2. Factory Integration - Configuration + Logging + Injection

**Factories that use configuration and logging together**

```go
// Factory that uses both configuration and logging
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(app *da.Application[AppConfig]) error {
    config := app.Config
    
    // Create database with configuration from dd
    db := &Database{
        URL:      config.Database.URL,
        PoolSize: config.Database.PoolSize,
    }
    
    // Register in container (da)
    da.Set(app.R, db)
    
    // Log with structured logging (dl)
    dl.ChannelLog("database").With("url", db.URL).With("pool_size", db.PoolSize).Info("database configured")
    
    return nil
}

type Database struct {
    URL      string
    PoolSize int
    logger   *dl.Builder
}

// Lifecycle integration
func (d *Database) Link(container *da.Container) error {
    // Create dedicated logger for this database instance
    d.logger = dl.ChannelLog("database").With("component", "database")
    return nil
}

func (d *Database) Start() error {
    d.logger.Info("connecting to database")
    // connection logic...
    d.logger.With("status", "connected").Info("database ready")
    return nil
}

func (d *Database) Stop() error {
    d.logger.Info("disconnecting from database")
    // cleanup logic...
    d.logger.Info("database disconnected")
    return nil
}
```

### 3. Dynamic Configuration - Runtime Reconfiguration

**Applications that can reload configuration at runtime**

```go
type ConfigurableApp struct {
    currentConfig *AppConfig
    app          *da.Application[AppConfig]
    configFile   string
}

func NewConfigurableApp(configFile string) *ConfigurableApp {
    return &ConfigurableApp{
        configFile: configFile,
    }
}

func (a *ConfigurableApp) Start() error {
    // Initial load
    return a.LoadConfig()
}

func (a *ConfigurableApp) LoadConfig() error {
    // Load configuration with dd
    config, err := dd.BindFromYAML[AppConfig](a.configFile)
    if err != nil {
        return err
    }
    
    // Check if this is a reload
    if a.currentConfig != nil {
        dl.Log().Info("reloading configuration")
        
        // Compare configs and reconfigure as needed
        if config.Logging != a.currentConfig.Logging {
            a.reconfigureLogging(config.Logging)
        }
        
        // Merge new config with existing application
        err := a.reconfigureApplication(*config)
        if err != nil {
            return err
        }
    } else {
        // Initial setup
        a.setupLogging(config.Logging)
        a.app = a.createApplication(*config)
        err := a.app.Initialize()
        if err != nil {
            return err
        }
        err = a.app.Start()
        if err != nil {
            return err
        }
    }
    
    a.currentConfig = config
    dl.Log().With("app", config.AppName).Info("configuration loaded")
    return nil
}

func (a *ConfigurableApp) reconfigureLogging(config LoggingConfig) {
    dl.Log().Info("reconfiguring logging")
    
    // Reconfigure channels with new settings
    if config.UseJSON {
        dl.ConfigureChannel("api", dl.DefaultOptions().JSON())
    } else {
        dl.ConfigureChannel("api", dl.DefaultOptions().Pretty())
    }
}

func (a *ConfigurableApp) reconfigureApplication(newConfig AppConfig) error {
    // Use dd.Merge to update configuration in place
    oldConfigData, _ := dd.Unbind(a.currentConfig)
    newConfigData, _ := dd.Unbind(&newConfig)
    
    // Intelligent merge of configurations
    err := dd.Merge(a.currentConfig, newConfigData)
    if err != nil {
        return err
    }
    
    // Notify services of configuration change
    configServices := da.AsType[ConfigurationReloader](a.app.R)
    for _, service := range configServices {
        err := service.ReloadConfiguration(*a.currentConfig)
        if err != nil {
            dl.ChannelLog("errors").With("service", service).Error("configuration reload failed")
        }
    }
    
    return nil
}
```

### 4. Plugin Architecture - Fully Dynamic Applications

**Load plugins with their own configuration, logging, and lifecycle**

```go
// Plugin configuration loaded from files
type PluginConfig struct {
    Name     string                 `dd:"name"`
    Type     string                 `dd:"type"`
    Enabled  bool                   `dd:"enabled"`
    Config   map[string]any         `dd:"config"`
    Logging  *LoggingConfig         `dd:"logging,omitempty"`
}

type PluginFactory struct{}

func (f *PluginFactory) Build(app *da.Application[AppConfig]) error {
    // Load plugin configurations
    pluginConfigs, err := loadPluginConfigs("plugins/")
    if err != nil {
        return err
    }
    
    for _, pluginConfig := range pluginConfigs {
        if !pluginConfig.Enabled {
            continue
        }
        
        // Setup plugin-specific logging
        if pluginConfig.Logging != nil {
            setupPluginLogging(pluginConfig.Name, *pluginConfig.Logging)
        }
        
        // Create plugin instance
        plugin, err := createPlugin(pluginConfig)
        if err != nil {
            dl.ChannelLog("plugins").With("plugin", pluginConfig.Name).Error("failed to create plugin")
            continue
        }
        
        // Register plugin in container
        da.SetNamed(app.R, pluginConfig.Name, plugin)
        da.SetNamed[Plugin](app.R, pluginConfig.Name, plugin)
        
        dl.ChannelLog("plugins").With("plugin", pluginConfig.Name).Info("plugin loaded")
    }
    
    return nil
}

func loadPluginConfigs(dir string) ([]PluginConfig, error) {
    // Find all plugin config files
    files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
    if err != nil {
        return nil, err
    }
    
    var configs []PluginConfig
    for _, file := range files {
        // Use dd to load each plugin config
        config, err := dd.BindFromYAML[PluginConfig](file)
        if err != nil {
            dl.ChannelLog("plugins").With("file", file).Warn("failed to load plugin config")
            continue
        }
        configs = append(configs, *config)
    }
    
    return configs, nil
}

func createPlugin(config PluginConfig) (Plugin, error) {
    // Use dd to convert generic config to specific plugin config
    switch config.Type {
    case "http":
        var httpConfig HTTPPluginConfig
        err := dd.Bind(&httpConfig, config.Config)
        if err != nil {
            return nil, err
        }
        return NewHTTPPlugin(config.Name, httpConfig), nil
        
    case "database":
        var dbConfig DatabasePluginConfig  
        err := dd.Bind(&dbConfig, config.Config)
        if err != nil {
            return nil, err
        }
        return NewDatabasePlugin(config.Name, dbConfig), nil
        
    default:
        return nil, fmt.Errorf("unknown plugin type: %s", config.Type)
    }
}
```

### 5. Microservice Orchestration - Multiple Services

**Coordinate multiple services with shared configuration and logging**

```go
type ServiceOrchestrator struct {
    services map[string]*da.Application[ServiceConfig]
    logger   *dl.Builder
}

type ServiceConfig struct {
    ServiceName string            `dd:"service_name"`
    Port        int               `dd:"port"`
    Database    DatabaseConfig    `dd:"database"`
    Logging     LoggingConfig     `dd:"logging"`
    Features    []string          `dd:"features"`
}

func NewServiceOrchestrator() *ServiceOrchestrator {
    return &ServiceOrchestrator{
        services: make(map[string]*da.Application[ServiceConfig]),
        logger:   dl.ChannelLog("orchestrator"),
    }
}

func (o *ServiceOrchestrator) LoadServices(configDir string) error {
    // Find all service configurations
    configFiles, err := filepath.Glob(filepath.Join(configDir, "*.yaml"))
    if err != nil {
        return err
    }
    
    for _, configFile := range configFiles {
        err := o.loadService(configFile)
        if err != nil {
            o.logger.With("file", configFile).Error("failed to load service")
            continue
        }
    }
    
    return nil
}

func (o *ServiceOrchestrator) loadService(configFile string) error {
    // Load service configuration with dd
    config, err := dd.BindFromYAML[ServiceConfig](configFile)
    if err != nil {
        return err
    }
    
    serviceName := config.ServiceName
    
    // Setup service-specific logging with dl
    o.setupServiceLogging(serviceName, config.Logging)
    
    // Create service application with da
    app := da.NewApplication(*config)
    
    // Register common factories
    da.WithFactory(app, &DatabaseFactory{})
    da.WithFactory(app, &LoggerFactory{})
    
    // Register feature-specific factories based on configuration
    for _, feature := range config.Features {
        factory := getFeatureFactory(feature)
        if factory != nil {
            da.WithFactory(app, factory)
        }
    }
    
    // Initialize service
    err = app.Initialize()
    if err != nil {
        return err
    }
    
    o.services[serviceName] = app
    o.logger.With("service", serviceName).Info("service loaded")
    
    return nil
}

func (o *ServiceOrchestrator) StartAll() error {
    for name, app := range o.services {
        err := app.Start()
        if err != nil {
            o.logger.With("service", name).Error("failed to start service")
            return err
        }
        o.logger.With("service", name).Info("service started")
    }
    return nil
}

func (o *ServiceOrchestrator) StopAll() error {
    for name, app := range o.services {
        err := app.Stop()
        if err != nil {
            o.logger.With("service", name).Error("failed to stop service")
        } else {
            o.logger.With("service", name).Info("service stopped")
        }
    }
    return nil
}
```

### 6. Configuration Validation and Migration

**Validate configurations and migrate between versions**

```go
type ConfigManager struct {
    validator *ConfigValidator
    migrator  *ConfigMigrator
    logger    *dl.Builder
}

type ConfigValidator struct{}

func (v *ConfigValidator) Validate(config *AppConfig) error {
    errors := make([]string, 0)
    
    // Use dd inspection to validate structure
    if config.Port <= 0 || config.Port > 65535 {
        errors = append(errors, "port must be between 1 and 65535")
    }
    
    if config.Database.URL == "" {
        errors = append(errors, "database.url is required")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
    }
    
    return nil
}

type ConfigMigrator struct{}

func (m *ConfigMigrator) MigrateToLatest(configData map[string]any) (map[string]any, error) {
    version, exists := configData["version"]
    if !exists {
        version = "1.0"
    }
    
    versionStr, ok := version.(string)
    if !ok {
        return nil, fmt.Errorf("invalid version format")
    }
    
    switch versionStr {
    case "1.0":
        return m.migrateFrom1_0To1_1(configData)
    case "1.1":
        return m.migrateFrom1_1To1_2(configData)
    case "1.2":
        return configData, nil // latest version
    default:
        return nil, fmt.Errorf("unsupported version: %s", versionStr)
    }
}

func (m *ConfigMigrator) migrateFrom1_0To1_1(data map[string]any) (map[string]any, error) {
    // Example migration: split database_url into host and port
    if dbURL, exists := data["database_url"]; exists {
        delete(data, "database_url")
        
        // Create nested database config
        data["database"] = map[string]any{
            "url":       dbURL,
            "pool_size": 10, // new field with default
        }
    }
    
    data["version"] = "1.1"
    return data, nil
}

func LoadAndMigrateConfig(filename string) (*AppConfig, error) {
    // Load as generic data first
    var rawData map[string]any
    err := dd.BindFromYAML(&rawData, filename)
    if err != nil {
        return nil, err
    }
    
    // Migrate to latest version
    manager := &ConfigManager{
        validator: &ConfigValidator{},
        migrator:  &ConfigMigrator{},
        logger:    dl.ChannelLog("config"),
    }
    
    migratedData, err := manager.migrator.MigrateToLatest(rawData)
    if err != nil {
        return nil, err
    }
    
    // Convert to typed config
    config, err := dd.New[AppConfig](migratedData)
    if err != nil {
        return nil, err
    }
    
    // Validate final configuration
    err = manager.validator.Validate(config)
    if err != nil {
        return nil, err
    }
    
    manager.logger.Info("configuration loaded and validated")
    return config, nil
}
```

## Integration Patterns Summary

| Pattern | dd Usage | dl Usage | da Usage |
|---------|----------|----------|----------|
| **Basic App** | Load config | Setup logging | Manage lifecycle |
| **Factory Integration** | Config in factories | Structured logging | Dependency injection |
| **Dynamic Reconfig** | Merge configs | Reconfigure channels | Reload services |
| **Plugin Architecture** | Plugin configs | Plugin logging | Plugin lifecycle |
| **Microservices** | Service configs | Service logging | Service orchestration |
| **Config Management** | Validation/migration | Operation logging | Component coordination |

## Best Practices

### Configuration
- Use `dd.Merge()` for layered configuration (defaults → env → user)
- Validate configurations after loading but before application start
- Use typed configurations with struct tags for clear field mapping
- Implement configuration migration for version compatibility

### Logging  
- Setup logging early in application startup
- Use channel-based routing for different concerns (db, api, errors)
- Include request/correlation IDs for tracing across services
- Configure different output formats for different environments

### Application Management
- Design factories to be order-independent (dependencies resolved in linking)
- Implement all three lifecycle interfaces (Startable, Stoppable, Linkable) where appropriate
- Use container introspection for debugging and health checks
- Plan for graceful shutdown with proper resource cleanup

### Integration
- Let `dd` handle all data transformation between formats
- Use `dl` for all application logging instead of direct `slog` usage
- Centralize object creation in `da` factories rather than scattered throughout code
- Design for configuration-driven behavior to maximize application flexibility

---

*This guide shows the df framework's power when all three packages work together. Each package solves specific problems, but together they enable building truly dynamic, configuration-driven applications at any scale.*