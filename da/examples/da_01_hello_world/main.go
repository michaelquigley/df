package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df/da"
)

// configuration for our application
type Config struct {
	AppName     string `json:"app_name"`
	DatabaseURL string `json:"database_url"`
	LogLevel    string `json:"log_level"`
}

// a database connection
type Database struct {
	URL       string
	Connected bool
}

func (d *Database) Connect() error {
	fmt.Printf("connecting to database: %s\n", d.URL)
	d.Connected = true
	return nil
}

func (d *Database) Start() error {
	return d.Connect()
}

func (d *Database) Stop() error {
	fmt.Printf("disconnecting from database: %s\n", d.URL)
	d.Connected = false
	return nil
}

// a logger service
type Logger struct {
	Level string
}

func (l *Logger) Info(msg string) {
	if l.Level == "info" || l.Level == "debug" {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

func (l *Logger) Start() error {
	fmt.Printf("starting logger with level: %s\n", l.Level)
	return nil
}

func (l *Logger) Stop() error {
	fmt.Println("stopping logger")
	return nil
}

// factory that creates our database
type DatabaseFactory struct{}

func (f *DatabaseFactory) Build(a *da.Application[Config]) error {
	cfg, _ := da.Get[Config](a.C)

	db := &Database{
		URL: cfg.DatabaseURL,
	}

	da.SetAs[*Database](a.C, db)
	return nil
}

// factory that creates our logger
type LoggerFactory struct{}

func (f *LoggerFactory) Build(a *da.Application[Config]) error {
	cfg, _ := da.Get[Config](a.C)

	logger := &Logger{
		Level: cfg.LogLevel,
	}

	da.SetAs[*Logger](a.C, logger)
	return nil
}

func main() {
	// create initial configuration
	cfg := Config{
		AppName:     "example app",
		DatabaseURL: "postgres://localhost:5432/mydb",
		LogLevel:    "info",
	}

	// create application with factories
	app := da.NewApplication(cfg)
	da.WithFactory(app, &DatabaseFactory{})
	da.WithFactory(app, &LoggerFactory{})

	// initialize: build objects and link dependencies
	if err := app.Initialize(); err != nil {
		log.Fatal(err)
	}

	// start all services
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	// show what's in our container
	fmt.Println("\n=== container contents ===")
	output, _ := app.C.Inspect(da.InspectHuman)
	fmt.Println(output)

	// use our services
	logger, _ := da.Get[*Logger](app.C)
	logger.Info("application started successfully")

	db, _ := da.Get[*Database](app.C)
	fmt.Printf("database connected: %v\n", db.Connected)

	// demonstrate named objects
	da.SetNamed(app.C, "audit", &Logger{Level: "debug"})
	da.SetNamed(app.C, "cache", &Database{URL: "redis://localhost:6379"})

	// show all loggers
	loggers := da.OfType[*Logger](app.C)
	fmt.Printf("\nfound %d loggers:\n", len(loggers))
	for i, l := range loggers {
		fmt.Printf("  [%d] level: %s\n", i, l.Level)
	}

	// find all startable services
	startables := da.AsType[da.Startable](app.C)
	fmt.Printf("\nfound %d startable services\n", len(startables))

	// clean shutdown
	if err := app.Stop(); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
}
