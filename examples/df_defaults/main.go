package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df"
)

type ServerConfig struct {
	Host    string `df:"host"`
	Port    int    `df:"port"`
	Timeout int    `df:"timeout"`
	Debug   bool   `df:"debug"`
}

type DatabaseConfig struct {
	Host     string `df:"host"`
	Port     int    `df:"port"`
	Database string `df:"database"`
	SSL      bool   `df:"ssl"`
}

type AppConfig struct {
	Server   ServerConfig   `df:"server"`
	Database DatabaseConfig `df:"database"`
	Features []string       `df:"features"`
}

func main() {
	// pre-initialized config with default values
	config := &AppConfig{
		Server: ServerConfig{
			Host:    "localhost",
			Port:    8080,
			Timeout: 30,
			Debug:   false,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "myapp",
			SSL:      true,
		},
		Features: []string{"basic", "auth"},
	}

	fmt.Println("=== original config with defaults ===")
	fmt.Printf("server: %+v\n", config.Server)
	fmt.Printf("database: %+v\n", config.Database)
	fmt.Printf("features: %v\n", config.Features)

	// partial configuration data (only overrides specific values)
	partialData := map[string]any{
		"server": map[string]any{
			"host":  "api.example.com",
			"debug": true,
		},
		"database": map[string]any{
			"host": "db.example.com",
			"port": 3306,
		},
		"features": []string{"basic", "auth", "premium"},
	}

	// bind partial data to existing config, preserving defaults
	if err := df.BindTo(config, partialData); err != nil {
		log.Fatalf("failed to bind partial config: %v", err)
	}

	fmt.Println("\n=== config after BindTo with partial data ===")
	fmt.Printf("server: %+v\n", config.Server)
	fmt.Printf("database: %+v\n", config.Database)
	fmt.Printf("features: %v\n", config.Features)

	fmt.Println("\n=== preserved defaults ===")
	fmt.Printf("server.port: %d (preserved)\n", config.Server.Port)
	fmt.Printf("server.timeout: %d (preserved)\n", config.Server.Timeout)
	fmt.Printf("database.database: %s (preserved)\n", config.Database.Database)
	fmt.Printf("database.ssl: %t (preserved)\n", config.Database.SSL)
}
