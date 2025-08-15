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
	Password string `df:"secret"`
	Database string
}

type ServiceConfig struct {
	Name          string
	URL           string `df:"url"`
	Enabled       bool
	CheckInterval time.Duration
}

func main() {
	// sample configuration data
	configData := map[string]any{
		"app_name": "mywebapp",
		"port":     8080,
		"api_key":  "sk-1234567890abcdef",
		"timeout":  "30s",
		"debug":    true,
		"database": map[string]any{
			"host":     "localhost",
			"port":     5432,
			"username": "appuser",
			"password": "supersecretpassword",
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

	// bind configuration
	config, err := df.New[AppConfig](configData)
	if err != nil {
		log.Fatalf("failed to bind config: %v", err)
	}

	fmt.Println("=== inspect configuration (secrets hidden) ===")
	output, err := df.Inspect(config)
	if err != nil {
		log.Fatalf("failed to inspect config: %v", err)
	}
	fmt.Println(output)

	fmt.Println("\n=== inspect configuration (secrets visible) ===")
	output, err = df.Inspect(config, &df.InspectOptions{ShowSecrets: true})
	if err != nil {
		log.Fatalf("failed to inspect config with secrets: %v", err)
	}
	fmt.Println(output)

	fmt.Println("\n=== inspect configuration (custom formatting) ===")
	output, err = df.Inspect(config, &df.InspectOptions{
		Indent:      "    ",
		ShowSecrets: false,
		MaxDepth:    3,
	})
	if err != nil {
		log.Fatalf("failed to inspect config with custom options: %v", err)
	}
	fmt.Println(output)
}
