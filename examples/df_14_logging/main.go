package main

import (
	"log/slog"
	"time"

	"github.com/michaelquigley/df"
)

// Config defines the application configuration structure
type Config struct {
	AppName string `df:"app_name"`
	Debug   bool   `df:"debug"`
}

func main() {
	// demonstrate basic logging setup
	demonstrateBasicLogging()

	// demonstrate channel-based logging
	demonstrateChannelLogging()

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
	df.InitLogging()

	// global logging functions
	df.Debug("this is a debug message")
	df.Info("application starting up")
	df.Warn("this is a warning message")
	df.Error("this is an error message")

	// formatted logging
	username := "alice"
	sessionId := "abc-123"
	df.Infof("user %s logged in with session %s", username, sessionId)
	df.Debugf("processing %d items", 42)

	// structured logging with key-value pairs
	df.Info("user action completed",
		"user", username,
		"action", "login",
		"duration", 150*time.Millisecond,
		"success", true)

	println()
}

func demonstrateChannelLogging() {
	println("=== channel-based logging demonstration ===")

	// create loggers for different channels
	authLogger := df.LoggerChannel("auth")
	dbLogger := df.LoggerChannel("database")
	httpLogger := df.LoggerChannel("http")

	// log messages with different channels
	authLogger.Info("user authentication successful", "user", "bob")
	dbLogger.Warn("connection pool running low", "available", 2, "max", 10)
	httpLogger.Error("request failed", "status", 500, "path", "/api/users")

	// formatted logging with channels
	authLogger.Debugf("validating token for user %s", "charlie")
	dbLogger.Infof("query executed in %v", 25*time.Millisecond)

	println()
}

func demonstrateContextualLogging() {
	println("=== contextual logging demonstration ===")

	// create a logger with persistent context
	requestLogger := df.Logger().
		With("request_id", "req-456").
		With("user_id", "user-789")

	// all messages from this logger will include the context
	requestLogger.Info("processing request")
	requestLogger.Debug("validating input parameters")
	requestLogger.Warn("rate limit approaching", "current", 95, "limit", 100)

	// chain additional context
	operationLogger := requestLogger.With("operation", "create_user")
	operationLogger.Info("starting operation")
	operationLogger.Error("validation failed", "field", "email", "reason", "invalid format")

	println()
}

func demonstrateJSONLogging() {
	println("=== json logging demonstration ===")

	// configure logging for json output
	jsonOpts := df.DefaultLogOptions().JSON().NoColor()
	jsonOpts.Level = slog.LevelDebug
	df.InitLogging(jsonOpts)

	// log some messages in json format
	df.Info("json logging enabled")
	df.LoggerChannel("api").Error("request processing failed",
		"method", "POST",
		"path", "/api/orders",
		"status", 400,
		"error", "invalid payload")

	df.Logger().
		With("component", "payment").
		With("transaction_id", "txn-999").
		Warn("payment processing delayed", "delay", 5*time.Second)

	println()
}

func demonstrateApplicationIntegration() {
	println("=== application integration demonstration ===")

	// create application with configuration
	cfg := Config{
		AppName: "logging-demo",
		Debug:   true,
	}
	app := df.NewApplication(cfg)

	// add logging factory
	app.Factories = append(app.Factories, &LoggingFactory{})

	// initialize the application
	if err := app.Build(); err != nil {
		df.Error("failed to build application", "error", err)
		return
	}

	if err := app.Link(); err != nil {
		df.Error("failed to link application", "error", err)
		return
	}

	// logging is now configured via the application container
	df.Info("application initialized successfully", "app", cfg.AppName)

	// demonstrate that logger is available in container
	logger, found := df.Get[*slog.Logger](app.C)
	if found && logger != nil {
		logger.Info("logger retrieved from container")
	}
}

// LoggingFactory demonstrates how to integrate logging with df application framework
type LoggingFactory struct{}

func (f *LoggingFactory) Build(a *df.Application[Config]) error {
	// get configuration from application
	cfg := a.Cfg

	// create logging options based on configuration
	opts := df.DefaultLogOptions()
	if cfg.Debug {
		opts.Level = slog.LevelDebug
	}

	// initialize logging
	df.InitLogging(opts)

	// register logger in container for dependency injection
	// note: we would need access to the internal defaultLogger
	// for now, create a new logger instance to demonstrate the pattern
	handler := df.NewDfHandler(opts)
	logger := slog.New(handler)
	df.SetAs[*slog.Logger](a.C, logger)

	return nil
}
