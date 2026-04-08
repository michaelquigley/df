package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df/dd"
)

type ServerConfig struct {
	Host    string
	Port    int
	Timeout int
	Debug   bool
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	SSL      bool
}

type TLSConfig struct {
	ServerName string
	MinVersion string
}

func (c *TLSConfig) ApplyDefaults() {
	c.ServerName = "localhost"
	c.MinVersion = "1.3"
}

type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	TLS      *TLSConfig
	Features []string
}

func main() {
	fmt.Println("=== df.Merge() configuration defaults example ===")
	fmt.Println("demonstrates how df.Merge() enables layered configuration:")
	fmt.Println("• application defaults (compiled-in)")
	fmt.Println("• environment overrides (dev/staging/prod)")
	fmt.Println("• user preferences (runtime customization)")
	fmt.Println("• optional nested structs with merge-time defaults")

	// step 1: pre-initialized config with sensible defaults
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

	fmt.Println("\n=== step 1: application defaults (compiled-in) ===")
	fmt.Printf("server: %+v\n", config.Server)
	fmt.Printf("database: %+v\n", config.Database)
	fmt.Printf("tls: %v\n", config.TLS)
	fmt.Printf("features: %v\n", config.Features)

	// step 2: partial configuration from environment/config file (only specifies overrides)
	partialData := map[string]any{
		"server": map[string]any{
			"host":  "api.example.com", // override for production
			"debug": true,              // enable for debugging
			// note: port and timeout not specified - will preserve defaults
		},
		"database": map[string]any{
			"host": "db.example.com", // production database server
			"port": 3306,             // mysql instead of postgresql
			// note: database, ssl not specified - will preserve defaults
		},
		"tls": map[string]any{
			"server_name": "api.example.com",
			// note: min_version not specified - ApplyDefaults will provide it
		},
		"features": []string{"basic", "auth", "premium"}, // add premium features
	}

	fmt.Println("\n=== step 2: partial configuration (environment overrides) ===")
	fmt.Printf("partial data: %+v\n", partialData)
	fmt.Println("note: only specifies values that should change from defaults")

	// merge partial data onto existing config, intelligently preserving defaults
	if err := dd.Merge(config, partialData); err != nil {
		log.Fatalf("failed to merge partial config: %v", err)
	}

	fmt.Println("\n=== step 3: final merged configuration ===")
	fmt.Printf("server: %+v\n", config.Server)
	fmt.Printf("database: %+v\n", config.Database)
	if config.TLS != nil {
		fmt.Printf("tls: %+v\n", *config.TLS)
	} else {
		fmt.Println("tls: <nil>")
	}
	fmt.Printf("features: %v\n", config.Features)

	fmt.Println("\n=== key differences vs df.Bind() ===")
	fmt.Printf("✓ server.port: %d (preserved - not in partial data)\n", config.Server.Port)
	fmt.Printf("✓ server.timeout: %d (preserved - not in partial data)\n", config.Server.Timeout)
	fmt.Printf("✓ database.database: %s (preserved - not in partial data)\n", config.Database.Database)
	fmt.Printf("✓ database.ssl: %t (preserved - not in partial data)\n", config.Database.SSL)
	if config.TLS != nil {
		fmt.Printf("✓ tls.min_version: %s (from ApplyDefaults on fresh allocation)\n", config.TLS.MinVersion)
	}
	fmt.Printf("• df.Bind() would have zeroed these fields, df.Merge() preserves them\n")
	fmt.Printf("• df.Bind() would not call ApplyDefaults for a fresh tls pointer\n")
	fmt.Printf("• this enables minimal config files with maximum flexibility\n")
}
