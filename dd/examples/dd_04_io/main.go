package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/michaelquigley/df/dd"
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
	Password string `dd:"password,+secret"`
}

type AppConfig struct {
	Name     string `dd:"app_name"`
	Version  string
	Server   ServerConfig
	Database DatabaseConfig
	Features []string `dd:"enabled_features"`
}

func main() {
	fmt.Println("=== df i/o example ===")
	fmt.Println("demonstrates JSON/YAML binding and unbinding from bytes, readers/writers, and files")

	// step 1: demonstrate bytes layer (base)
	fmt.Println("\n=== step 1: bytes layer ===")
	demonstrateBytesLayer()

	// step 2: demonstrate reader/writer layer
	fmt.Println("\n=== step 2: reader/writer layer ===")
	demonstrateReaderWriterLayer()

	// step 3: demonstrate file layer
	fmt.Println("\n=== step 3: file layer ===")
	demonstrateFileLayer()

	// step 4: demonstrate format conversion
	fmt.Println("\n=== step 4: format conversion (JSON to YAML) ===")
	demonstrateFormatConversion()

	// step 5: demonstrate error handling
	fmt.Println("\n=== step 5: error handling ===")
	demonstrateErrorHandling()

	fmt.Println("\n=== i/o example completed successfully! ===")
	fmt.Println("key benefits:")
	fmt.Println("- bind/unbind from bytes, strings, readers, writers, and files")
	fmt.Println("- consistent error reporting across all i/o types")
	fmt.Println("- seamless format conversion between JSON and YAML")
	fmt.Println("- layered api for flexibility and composability")
}

func demonstrateBytesLayer() {
	// bind from JSON bytes
	jsonBytes := []byte(`{
		"app_name": "myapp",
		"version": "1.0.0",
		"server": {"host": "localhost", "port": 8080}
	}`)

	var config AppConfig
	if err := dd.BindJSON(&config, jsonBytes); err != nil {
		log.Fatalf("failed to bind JSON bytes: %v", err)
	}
	fmt.Printf("bound from JSON bytes: %s v%s at %s:%d\n",
		config.Name, config.Version, config.Server.Host, config.Server.Port)

	// bind from JSON string (using []byte cast)
	jsonStr := `{"app_name": "stringapp", "version": "2.0.0"}`
	var strConfig AppConfig
	if err := dd.BindJSON(&strConfig, []byte(jsonStr)); err != nil {
		log.Fatalf("failed to bind JSON string: %v", err)
	}
	fmt.Printf("bound from JSON string: %s v%s\n", strConfig.Name, strConfig.Version)

	// unbind to JSON bytes
	config.Version = "1.1.0"
	data, err := dd.UnbindJSON(config)
	if err != nil {
		log.Fatalf("failed to unbind to JSON bytes: %v", err)
	}
	fmt.Printf("unbound to JSON bytes (%d bytes)\n", len(data))

	// bind from YAML bytes
	yamlBytes := []byte(`app_name: yamlapp
version: 3.0.0
server:
  host: api.example.com
  port: 9090
`)
	var yamlConfig AppConfig
	if err := dd.BindYAML(&yamlConfig, yamlBytes); err != nil {
		log.Fatalf("failed to bind YAML bytes: %v", err)
	}
	fmt.Printf("bound from YAML bytes: %s v%s at %s:%d\n",
		yamlConfig.Name, yamlConfig.Version, yamlConfig.Server.Host, yamlConfig.Server.Port)
}

func demonstrateReaderWriterLayer() {
	// bind from JSON reader
	jsonReader := strings.NewReader(`{"app_name": "readerapp", "version": "4.0.0"}`)
	var config AppConfig
	if err := dd.BindJSONReader(&config, jsonReader); err != nil {
		log.Fatalf("failed to bind from JSON reader: %v", err)
	}
	fmt.Printf("bound from JSON reader: %s v%s\n", config.Name, config.Version)

	// unbind to JSON writer
	config.Version = "4.1.0"
	var buf bytes.Buffer
	if err := dd.UnbindJSONWriter(config, &buf); err != nil {
		log.Fatalf("failed to unbind to JSON writer: %v", err)
	}
	fmt.Printf("unbound to JSON writer (%d bytes)\n", buf.Len())

	// bind from YAML reader
	yamlReader := strings.NewReader(`app_name: yamlreaderapp
version: 5.0.0
`)
	var yamlConfig AppConfig
	if err := dd.BindYAMLReader(&yamlConfig, yamlReader); err != nil {
		log.Fatalf("failed to bind from YAML reader: %v", err)
	}
	fmt.Printf("bound from YAML reader: %s v%s\n", yamlConfig.Name, yamlConfig.Version)
}

func demonstrateFileLayer() {
	// create sample config for file operations
	config := AppConfig{
		Name:    "fileapp",
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

	// write to JSON file
	if err := dd.UnbindJSONFile(config, "config.json"); err != nil {
		log.Fatalf("failed to write JSON file: %v", err)
	}
	fmt.Println("wrote config.json")

	// write to YAML file
	if err := dd.UnbindYAMLFile(config, "config.yaml"); err != nil {
		log.Fatalf("failed to write YAML file: %v", err)
	}
	fmt.Println("wrote config.yaml")

	// read from JSON file
	var jsonConfig AppConfig
	if err := dd.BindJSONFile(&jsonConfig, "config.json"); err != nil {
		log.Fatalf("failed to read JSON file: %v", err)
	}
	fmt.Printf("read from config.json: %s v%s\n", jsonConfig.Name, jsonConfig.Version)

	// read from YAML file
	var yamlConfig AppConfig
	if err := dd.BindYAMLFile(&yamlConfig, "config.yaml"); err != nil {
		log.Fatalf("failed to read YAML file: %v", err)
	}
	fmt.Printf("read from config.yaml: %s v%s\n", yamlConfig.Name, yamlConfig.Version)

	// cleanup
	os.Remove("config.json")
	os.Remove("config.yaml")
	fmt.Println("cleaned up temporary files")
}

func demonstrateFormatConversion() {
	// start with JSON bytes
	jsonBytes := []byte(`{
		"app_name": "convertapp",
		"version": "1.0.0",
		"server": {"host": "localhost", "port": 8080}
	}`)

	// bind from JSON
	var config AppConfig
	if err := dd.BindJSON(&config, jsonBytes); err != nil {
		log.Fatalf("failed to bind JSON: %v", err)
	}

	// unbind to YAML
	yamlData, err := dd.UnbindYAML(config)
	if err != nil {
		log.Fatalf("failed to unbind to YAML: %v", err)
	}

	fmt.Println("converted JSON to YAML:")
	fmt.Println(string(yamlData))
}

func demonstrateErrorHandling() {
	var config AppConfig

	// try to load non-existent file
	err := dd.BindJSONFile(&config, "nonexistent.json")
	if err != nil {
		fmt.Printf("expected file error: %v\n", err)
	}

	// try to parse invalid JSON
	invalidJSON := []byte(`{"invalid": json`)
	err = dd.BindJSON(&config, invalidJSON)
	if err != nil {
		fmt.Printf("expected parse error: %v\n", err)
	}

	// try to parse invalid YAML
	invalidYAML := []byte(`invalid: yaml: content:`)
	err = dd.BindYAML(&config, invalidYAML)
	if err != nil {
		fmt.Printf("expected YAML parse error: %v\n", err)
	}
}
