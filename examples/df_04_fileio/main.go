package main

import (
	"fmt"
	"log"
	"os"

	"github.com/michaelquigley/df"
)

type ServerConfig struct {
	Host    string
	Port    int
	Timeout int
	Debug   bool
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string `df:"password,secret"`
}

type AppConfig struct {
	Name     string         `df:"app_name"`
	Version  string
	Server   ServerConfig
	Database DatabaseConfig
	Features []string       `df:"enabled_features"`
}

func main() {
	fmt.Println("=== df file i/o example ===")
	fmt.Println("demonstrates direct JSON/YAML file binding and unbinding")
	fmt.Println("with comprehensive error handling and format conversion")

	// step 1: create sample configuration files
	fmt.Println("\n=== step 1: creating sample configuration files ===")
	if err := createSampleFiles(); err != nil {
		log.Fatalf("failed to create sample files: %v", err)
	}
	fmt.Printf("✓ created config.json and config.yaml\n")

	// step 2: load configuration from JSON file
	fmt.Println("\n=== step 2: loading configuration from JSON file ===")
	var jsonConfig AppConfig
	if err := df.BindFromJSON(&jsonConfig, "config.json"); err != nil {
		log.Fatalf("failed to load JSON config: %v", err)
	}
	fmt.Printf("✓ loaded from config.json: %s v%s\n", jsonConfig.Name, jsonConfig.Version)
	fmt.Printf("  server: %s:%d (timeout: %ds, debug: %t)\n",
		jsonConfig.Server.Host, jsonConfig.Server.Port, jsonConfig.Server.Timeout, jsonConfig.Server.Debug)

	// step 3: load configuration from YAML file
	fmt.Println("\n=== step 3: loading configuration from YAML file ===")
	var yamlConfig AppConfig
	if err := df.BindFromYAML(&yamlConfig, "config.yaml"); err != nil {
		log.Fatalf("failed to load YAML config: %v", err)
	}
	fmt.Printf("✓ loaded from config.yaml: %s v%s\n", yamlConfig.Name, yamlConfig.Version)
	fmt.Printf("  database: %s@%s:%d/%s\n",
		yamlConfig.Database.Username, yamlConfig.Database.Host,
		yamlConfig.Database.Port, yamlConfig.Database.Database)

	// step 4: modify configuration and save to files
	fmt.Println("\n=== step 4: modifying and saving configuration ===")
	jsonConfig.Version = "1.2.0"
	jsonConfig.Server.Debug = true
	jsonConfig.Features = append(jsonConfig.Features, "monitoring", "metrics")

	// save modified config as JSON
	if err := df.UnbindToJSON(jsonConfig, "output.json"); err != nil {
		log.Fatalf("failed to save JSON config: %v", err)
	}
	fmt.Printf("✓ saved modified config to output.json\n")

	// save modified config as YAML
	if err := df.UnbindToYAML(jsonConfig, "output.yaml"); err != nil {
		log.Fatalf("failed to save YAML config: %v", err)
	}
	fmt.Printf("✓ saved modified config to output.yaml\n")

	// step 5: demonstrate error handling
	fmt.Println("\n=== step 5: error handling demonstration ===")
	var errorConfig AppConfig

	// try to load non-existent file
	err := df.BindFromJSON(&errorConfig, "nonexistent.json")
	if err != nil {
		fmt.Printf("expected file error: %v\n", err)
	}

	// try to load invalid JSON
	if err := os.WriteFile("invalid.json", []byte(`{"invalid": json`), 0644); err == nil {
		err = df.BindFromJSON(&errorConfig, "invalid.json")
		if err != nil {
			fmt.Printf("expected parse error: %v\n", err)
		}
	}

	fmt.Println("\n=== step 6: format conversion (JSON to YAML) ===")
	// load from JSON and save as YAML
	var convertConfig AppConfig
	if err := df.BindFromJSON(&convertConfig, "output.json"); err != nil {
		log.Fatalf("failed to load for conversion: %v", err)
	}

	if err := df.UnbindToYAML(convertConfig, "converted.yaml"); err != nil {
		log.Fatalf("failed to convert to YAML: %v", err)
	}
	fmt.Printf("✓ converted output.json → converted.yaml\n")

	// cleanup temporary files
	fmt.Println("\n=== cleanup ===")
	files := []string{"config.json", "config.yaml", "output.json", "output.yaml", "converted.yaml", "invalid.json"}
	for _, file := range files {
		if err := os.Remove(file); err == nil {
			fmt.Printf("✓ removed %s\n", file)
		}
	}

	fmt.Println("\n=== file i/o example completed successfully! ===")
	fmt.Println("key benefits:")
	fmt.Println("• direct file operations without manual JSON/YAML handling")
	fmt.Println("• consistent error reporting for file and parsing issues")
	fmt.Println("• seamless format conversion between JSON and YAML")
	fmt.Println("• production-ready configuration loading patterns")
}

func createSampleFiles() error {
	// create sample JSON config
	jsonConfig := AppConfig{
		Name:    "mywebapp",
		Version: "1.0.0",
		Server: ServerConfig{
			Host:    "localhost",
			Port:    8080,
			Timeout: 30,
			Debug:   false,
		},
		Database: DatabaseConfig{
			Driver:   "postgresql",
			Host:     "localhost",
			Port:     5432,
			Database: "myapp",
			Username: "appuser",
			Password: "secret123",
		},
		Features: []string{"auth", "logging"},
	}

	if err := df.UnbindToJSON(jsonConfig, "config.json"); err != nil {
		return fmt.Errorf("failed to create config.json: %v", err)
	}

	// create sample YAML config (different values to show variety)
	yamlConfig := AppConfig{
		Name:    "mywebapp-yaml",
		Version: "1.1.0",
		Server: ServerConfig{
			Host:    "api.example.com",
			Port:    9090,
			Timeout: 45,
			Debug:   true,
		},
		Database: DatabaseConfig{
			Driver:   "mysql",
			Host:     "db.example.com",
			Port:     3306,
			Database: "prod_db",
			Username: "produser",
			Password: "prodpass456",
		},
		Features: []string{"auth", "logging", "caching"},
	}

	if err := df.UnbindToYAML(yamlConfig, "config.yaml"); err != nil {
		return fmt.Errorf("failed to create config.yaml: %v", err)
	}

	return nil
}
