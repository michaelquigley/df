package main

import (
	"fmt"
	"log"
	"time"

	"github.com/michaelquigley/df"
)

type AppConfig struct {
	Name     string `df:"app_name"`
	Port     int
	APIKey   string `df:"api_key,secret"`
	Timeout  time.Duration
	Debug    bool
	Database *DatabaseConfig
	Services []ServiceConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string `df:",secret"`
	Database string
}

type ServiceConfig struct {
	Name          string
	URL           string
	Enabled       bool
	CheckInterval time.Duration
}

func main() {
	fmt.Println("=== df.Inspect() debugging example ===")
	fmt.Println("demonstrates how to debug and validate bound configuration")
	fmt.Println("with automatic secret filtering and human-readable output")

	// sample configuration data with both public and secret fields
	configData := map[string]any{
		"app_name": "mywebapp",
		"port":     8080,
		"api_key":  "sk-1234567890abcdef", // secret field
		"timeout":  "30s",
		"debug":    true,
		"database": map[string]any{
			"host":     "localhost",
			"port":     5432,
			"username": "appuser",
			"password": "supersecretpassword", // secret field
			"database": "myapp_db",
		},
		"services": []any{
			map[string]any{
				"name":    "auth-service",
				"url":     "http://auth.example.com:8081",
				"enabled": true,
			},
			map[string]any{
				"name":           "payment-service",
				"url":            "http://payments.example.com:8082",
				"enabled":        false,
				"check_interval": "5m",
			},
		},
	}

	fmt.Println("\n=== input configuration data ===")
	fmt.Printf("note: contains secret fields (api_key, password)\n")

	// bind configuration to strongly-typed structs
	config, err := df.New[AppConfig](configData)
	if err != nil {
		log.Fatalf("failed to bind config: %v", err)
	}

	fmt.Println("\n=== 1. default inspect (secrets automatically hidden) ===")
	output, err := df.Inspect(config)
	if err != nil {
		log.Fatalf("failed to inspect config: %v", err)
	}
	fmt.Println(output)
	fmt.Println("notice: secret fields are replaced with <set> (or <unset>) for security")

	fmt.Println("\n=== 2. inspect with secrets visible (debug mode) ===")
	output, err = df.Inspect(config, &df.InspectOptions{ShowSecrets: true})
	if err != nil {
		log.Fatalf("failed to inspect config with secrets: %v", err)
	}
	fmt.Println(output)
	fmt.Println("use ShowSecrets: true only in secure debugging environments")

	fmt.Println("\n=== 3. inspect with custom formatting ===")
	output, err = df.Inspect(config, &df.InspectOptions{
		Indent:      "    ", // wider indentation
		ShowSecrets: false,  // keep secrets hidden
		MaxDepth:    3,      // limit nesting depth
	})
	if err != nil {
		log.Fatalf("failed to inspect config with custom options: %v", err)
	}
	fmt.Println(output)
	fmt.Println("custom options allow control over output formatting")

	fmt.Println("\n=== use cases for df.Inspect() ===")
	fmt.Println("• debug configuration loading issues")
	fmt.Println("• validate config values after binding")
	fmt.Println("• generate human-readable config reports")
	fmt.Println("• log config state (with secrets filtered)")
	fmt.Println("• troubleshoot type conversion problems")
}
