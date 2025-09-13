package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dl"
)

// Config defines the application configuration structure
type Config struct {
	AppName string `dd:"app_name"`
	Debug   bool   `dd:"debug"`
}

func main() {
	// demonstrate basic logging setup
	demonstrateBasicLogging()

	// demonstrate channel-based logging
	demonstrateChannelLogging()

	// demonstrate per-channel configuration
	demonstratePerChannelConfiguration()

	// demonstrate contextual logging with builder pattern
	demonstrateContextualLogging()

	// demonstrate json output format
	demonstrateJSONLogging()

	// demonstrate integration with df application framework
	demonstrateApplicationIntegration()
}

func demonstrateBasicLogging() {
	println("=== basic logging demonstration ===")

	// initialize logging with default options
	dl.Init()

	// global logging functions
	dl.Log().Debug("this is a debug message")
	dl.Log().Info("application starting up")
	dl.Log().Warn("this is a warning message")
	dl.Log().Error("this is an error message")

	// formatted logging
	username := "alice"
	sessionId := "abc-123"
	dl.Log().Infof("user %s logged in with session %s", username, sessionId)
	dl.Log().Debugf("processing %d items", 42)

	// structured logging with key-value pairs
	dl.Log().
		With("user", username).
		With("action", "login").
		With("duration", 150*time.Millisecond).
		With("success", true).
		Info("user action completed")

	println()
}

func demonstrateChannelLogging() {
	println("=== channel-based logging demonstration ===")

	// create loggers for different channels
	authLogger := dl.ChannelLog("auth")
	dbLogger := dl.ChannelLog("database")
	httpLogger := dl.ChannelLog("http")

	// log messages with different channels
	authLogger.With("user", "bob").Info("user authentication successful")
	dbLogger.With("available", 2).With("max", 10).Warn("connection pool running low")
	httpLogger.With("status", 500).With("path", "/api/users").Error("request failed")

	// formatted logging with channels
	authLogger.Debugf("validating token for user %s", "charlie")
	dbLogger.Infof("query executed in %v", 25*time.Millisecond)

	println()
}

func demonstratePerChannelConfiguration() {
	println("=== per-channel configuration demonstration ===")

	// configure different destinations for different channels

	// create a temporary file for database logs
	dbFile, err := os.CreateTemp("", "database-*.log")
	if err != nil {
		dl.Log().With("error", err).Error("failed to create temp file")
		return
	}
	defer os.Remove(dbFile.Name())
	defer dbFile.Close()

	// configure database channel to log to file
	dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))

	// configure http channel for JSON format to stderr
	dl.ConfigureChannel("http", dl.DefaultOptions().JSON().SetOutput(os.Stderr))

	// configure error channel with different colors
	dl.ConfigureChannel("errors", dl.DefaultOptions().Color())

	// now log to different channels
	dl.ChannelLog("database").Info("database connection established")
	dl.ChannelLog("database").With("query", "SELECT * FROM users").Debug("executing query")

	dl.ChannelLog("http").With("method", "GET").With("path", "/api/users").Info("http request received")

	dl.ChannelLog("errors").With("code", 500).Error("internal server error")

	// normal channel logging still goes to console
	dl.ChannelLog("auth").Info("user authenticated successfully")

	// show what was written to the database log file
	dbFile.Seek(0, 0)
	content := make([]byte, 1024)
	n, _ := dbFile.Read(content)
	if n > 0 {
		println("\n--- content of database log file ---")
		println(string(content[:n]))
	}

	println()
}

func demonstrateContextualLogging() {
	println("=== contextual logging demonstration ===")

	// create a logger with persistent context
	requestLogger := dl.Log().
		With("request_id", "req-456").
		With("user_id", "user-789")

	// all messages from this logger will include the context
	requestLogger.Info("processing request")
	requestLogger.Debug("validating input parameters")
	requestLogger.With("current", 95).With("limit", 100).Warn("rate limit approaching")

	// chain additional context
	operationLogger := requestLogger.With("operation", "create_user")
	operationLogger.Info("starting operation")
	operationLogger.With("field", "email").With("reason", "invalid format").Error("validation failed")

	println()
}

func demonstrateJSONLogging() {
	println("=== json logging demonstration ===")

	// configure logging for json output
	jsonOpts := dl.DefaultOptions().JSON().NoColor()
	jsonOpts.Level = slog.LevelDebug
	dl.Init(jsonOpts)

	// log some messages in json format
	dl.Log().Info("json logging enabled")

	dl.ChannelLog("api").
		With("method", "POST").
		With("path", "/api/orders").
		With("status", 400).
		With("error", "invalid payload").
		Error("request processing failed")

	dl.Log().
		With("component", "payment").
		With("transaction_id", "txn-999").
		With("delay", 5*time.Second).
		Warn("payment processing delayed")

	println()
}

func demonstrateApplicationIntegration() {
	println("=== application integration demonstration ===")

	// create application with configuration
	cfg := Config{
		AppName: "logging-demo",
		Debug:   true,
	}
	app := da.NewApplication(cfg)

	// add logging factory
	app.Factories = append(app.Factories, &LoggingFactory{})

	// initialize the application
	if err := app.Build(); err != nil {
		dl.Log().With("error", err).Error("failed to build application")
		return
	}

	if err := app.Link(); err != nil {
		dl.Log().With("error", err).Error("failed to link application")
		return
	}

	// logging is now configured via the application container
	dl.Log().With("app", cfg.AppName).Info("application initialized successfully")

	// demonstrate that logger is available in container
	logger, found := da.Get[*slog.Logger](app.C)
	if found && logger != nil {
		logger.Info("logger retrieved from container")
	}
}

// LoggingFactory demonstrates how to integrate logging with df application framework
type LoggingFactory struct{}

func (f *LoggingFactory) Build(a *da.Application[Config]) error {
	// get configuration from application
	cfg := a.Cfg

	// create logging options based on configuration
	opts := dl.DefaultOptions()
	if cfg.Debug {
		opts.Level = slog.LevelDebug
	}

	// initialize logging
	dl.Init(opts)

	// register logger in container for dependency injection
	// for demonstration, create a new logger instance with the same options
	handler := dl.NewDfHandler(opts)
	logger := slog.New(handler)
	da.SetAs[*slog.Logger](a.C, logger)

	return nil
}
