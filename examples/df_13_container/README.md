# Container Example

This example demonstrates the df container builder pattern for dependency injection.

## What it shows

1. **Configuration Binding**: Load configuration from data into structs
2. **Factory Registration**: Register factories that create service objects
3. **Object Creation**: Execute all factories to create objects
4. **Dependency Wiring**: Link objects together with their dependencies
5. **Application Execution**: Use the final container to run the application

## Components

- `appConfig`: Configuration struct with server settings
- `database`: Simple database service that can connect
- `server`: Web server that depends on database and config
- `app`: Main application that coordinates everything

## Flow

1. Configuration data is bound to `appConfig` struct
2. Factories are registered for `database`, `server`, and `app`
3. All objects are created via their factories
4. Dependencies are wired: database connects, server gets database, app gets server
5. Final application runs using all configured services

## Run

```bash
go run main.go
```

Expected output shows each phase completing successfully, then the application starting with all dependencies properly wired.