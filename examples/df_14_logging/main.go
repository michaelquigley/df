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
	df.Log().Debug("this is a debug message")
	df.Log().Info("application starting up")
	df.Log().Warn("this is a warning message")
	df.Log().Error("this is an error message")

	// formatted logging
	username := "alice"
	sessionId := "abc-123"
	df.Log().Infof("user %s logged in with session %s", username, sessionId)
	df.Log().Debugf("processing %d items", 42)

	// structured logging with key-value pairs
	df.Log().
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
	authLogger := df.ChannelLog("auth")
	dbLogger := df.ChannelLog("database")
	httpLogger := df.ChannelLog("http")

	// log messages with different channels
	authLogger.With("user", "bob").Info("user authentication successful")
	dbLogger.With("available", 2).With("max", 10).Warn("connection pool running low")
	httpLogger.With("status", 500).With("path", "/api/users").Error("request failed")

	// formatted logging with channels
	authLogger.Debugf("validating token for user %s", "charlie")
	dbLogger.Infof("query executed in %v", 25*time.Millisecond)

	println()
}

func demonstrateContextualLogging() {
	println("=== contextual logging demonstration ===")

	// create a logger with persistent context
	requestLogger := df.Log().
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
	jsonOpts := df.DefaultLogOptions().JSON().NoColor()
	jsonOpts.Level = slog.LevelDebug
	df.InitLogging(jsonOpts)

	// log some messages in json format
	df.Log().Info("json logging enabled")

	df.ChannelLog("api").
		With("method", "POST").
		With("path", "/api/orders").
		With("status", 400).
		With("error", "invalid payload").
		Error("request processing failed")

	df.Log().
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
	app := df.NewApplication(cfg)

	// add logging factory
	app.Factories = append(app.Factories, &LoggingFactory{})

	// initialize the application
	if err := app.Build(); err != nil {
		df.Log().With("error", err).Error("failed to build application")
		return
	}

	if err := app.Link(); err != nil {
		df.Log().With("error", err).Error("failed to link application")
		return
	}

	// logging is now configured via the application container
	df.Log().With("app", cfg.AppName).Info("application initialized successfully")

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
