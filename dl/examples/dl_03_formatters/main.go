package main

import (
	"log"
	"os"

	"github.com/michaelquigley/df/dl"
)

func main() {
	log.Println("=== formatting demonstration ===")
	
	// create log files
	structuredFile, _ := os.Create("logs/structured.log")
	defer structuredFile.Close()
	
	// configure channels with different formatting
	
	// pretty format for console (human readable)
	dl.ConfigureChannel("console", dl.DefaultOptions().Pretty().Color())
	
	// json format for structured logging
	dl.ConfigureChannel("structured", dl.DefaultOptions().JSON().SetOutput(structuredFile))
	
	// no color format for files
	dl.ConfigureChannel("nocolor", dl.DefaultOptions().Pretty().NoColor().SetOutput(os.Stdout))
	
	// demonstrate the same events with different formatting
	
	// user authentication event
	dl.ChannelLog("console").With("user_id", 123).With("username", "john_doe").With("session_id", "abc123").Info("user authenticated")
	dl.ChannelLog("structured").With("user_id", 123).With("username", "john_doe").With("session_id", "abc123").Info("user authenticated")
	dl.ChannelLog("nocolor").With("user_id", 123).With("username", "john_doe").With("session_id", "abc123").Info("user authenticated")
	
	// database error event
	dl.ChannelLog("console").With("host", "db.example.com").With("port", 5432).With("error", "connection timeout").Error("database connection failed")
	dl.ChannelLog("structured").With("host", "db.example.com").With("port", 5432).With("error", "connection timeout").Error("database connection failed")
	dl.ChannelLog("nocolor").With("host", "db.example.com").With("port", 5432).With("error", "connection timeout").Error("database connection failed")
	
	// performance metric event
	dl.ChannelLog("console").With("endpoint", "/api/users").With("duration_ms", 245).With("memory_mb", 12.5).Info("request completed")
	dl.ChannelLog("structured").With("endpoint", "/api/users").With("duration_ms", 245).With("memory_mb", 12.5).Info("request completed")
	dl.ChannelLog("nocolor").With("endpoint", "/api/users").With("duration_ms", 245).With("memory_mb", 12.5).Info("request completed")
	
	log.Println("=== check console output above for pretty formats ===")
	log.Println("=== check logs/structured.log for json format ===")
}