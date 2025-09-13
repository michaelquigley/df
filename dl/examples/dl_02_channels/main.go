package main

import (
	"log"
	"os"

	"github.com/michaelquigley/df/dl"
)

func main() {
	// configure different channels for different purposes
	
	// database channel: logs to file
	dbFile, _ := os.Create("logs/database.log")
	defer dbFile.Close()
	dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))
	
	// http channel: json format to stderr
	dl.ConfigureChannel("http", dl.DefaultOptions().JSON().SetOutput(os.Stderr))
	
	// errors channel: colored console output
	dl.ConfigureChannel("errors", dl.DefaultOptions().Color())
	
	// demonstrate different channels routing to different destinations
	log.Println("=== channel routing demonstration ===")
	
	// database operations log to file
	dl.ChannelLog("database").With("user_id", 123).With("username", "john_doe").Info("user login")
	dl.ChannelLog("database").With("table", "users").With("duration_ms", 45).Info("query executed")
	
	// http requests log to stderr in json format
	dl.ChannelLog("http").With("method", "GET").With("path", "/api/users").With("status", 200).Info("request processed")
	dl.ChannelLog("http").With("method", "POST").With("path", "/api/users").With("status", 201).Info("request processed")
	
	// errors log to colored console
	dl.ChannelLog("errors").With("field", "email").With("value", "invalid-email").Error("validation failed")
	dl.ChannelLog("errors").With("client_ip", "192.168.1.100").With("requests", 95).Warn("rate limit approaching")
	
	// default channel uses standard console output
	dl.Log().With("version", "1.0.0").Info("application started")
	dl.Log().With("port", 8080).Info("listening on port")
	
	log.Println("=== check logs/database.log for database channel output ===")
	log.Println("=== check stderr for http channel json output ===")
	log.Println("=== errors and default logs appear on console ===")
}