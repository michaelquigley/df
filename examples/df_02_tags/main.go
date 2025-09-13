package main

import (
	"fmt"
	"log"
	"time"

	"github.com/michaelquigley/df/dd"
)

// APIConfiguration demonstrates field naming and requirements
type APIConfiguration struct {
	ServiceName string        `dd:"+required"` // default name + required
	Version     string        `dd:"+required"` // default name + required
	Host        string        // default snake_case: "host"
	Port        int           // default snake_case: "port"
	BasePath    string        `dd:"base_path"` // custom name only
	APIKey      string        `dd:"+secret"`   // secret
	DebugMode   bool          `dd:"debug"`     // custom name only
	Timeout     time.Duration // default snake_case: "timeout"
	Internal    string        `dd:"-"` // excluded from binding/unbinding
}

// UserProfile demonstrates privacy controls and field naming
type UserProfile struct {
	Username    string           `dd:",+required"`   // required with default name
	Email       string           `dd:",+required"`   // required email
	DisplayName string           `dd:"display_name"` // custom field name
	FirstName   string           `dd:"first_name"`   // custom field name
	LastName    string           `dd:"last_name"`    // custom field name
	Phone       string           // default name: "phone"
	Password    string           `dd:",+secret"`    // secret field
	SSN         string           `dd:"ssn,+secret"` // custom name + secret
	Preferences *UserPreferences // nested struct
}

// UserPreferences demonstrates nested struct tag behavior
type UserPreferences struct {
	Theme           string // default: "theme"
	Language        string // default: "language"
	Notifications   bool   `dd:"enable_notifications"` // custom name
	Newsletter      bool   // default: "newsletter"
	PrivateProfile  bool   `dd:"private_profile"` // custom name
	SessionToken    string `dd:",+secret"`        // secret in nested struct
	InternalSetting string `dd:"-"`               // excluded
}

// SystemSettings demonstrates hierarchical configuration
type SystemSettings struct {
	Environment string          `dd:"env,+required"` // required environment
	Database    *DatabaseConfig // nested config
	Cache       *CacheConfig    // optional nested struct
	Features    *FeatureFlags   // feature toggles
}

// DatabaseConfig with security and naming
type DatabaseConfig struct {
	Host     string `dd:",+required"` // required host
	Port     int    // default: "port"
	Name     string `dd:"database_name,+required"` // custom name + required
	SSL      bool   `dd:"enable_ssl"`              // custom name
	Username string `dd:",+required"`              // required username
	Password string `dd:",+secret"`                // secret password
}

// CacheConfig with provider flexibility
type CacheConfig struct {
	Provider  string        `dd:",+required"` // required provider
	Host      string        // default: "host"
	Port      int           // default: "port"
	Password  string        `dd:",+secret"`    // secret password
	TTL       time.Duration `dd:"default_ttl"` // custom name
	MaxMemory int64         `dd:"max_memory"`  // custom name
}

// LoggingConfig demonstrates complex naming patterns
type LoggingConfig struct {
	Level     string `dd:",+required"` // required log level
	Output    string // default: "output"
	Format    string // default: "format"
	Filename  string `dd:"log_file"`            // custom name
	MaxSize   int    `dd:"max_size_mb"`         // custom name
	Rotate    bool   `dd:"enable_rotation"`     // custom name
	Sensitive string `dd:"debug_token,+secret"` // secret + custom name
}

// FeatureFlags for system capabilities
type FeatureFlags struct {
	EnableBeta     bool `dd:"beta_features"`       // custom name
	EnableMetrics  bool `dd:"metrics_collection"`  // custom name
	EnableTracing  bool `dd:"distributed_tracing"` // custom name
	EnableDebug    bool // default: "enable_debug"
	EnableAPI      bool `dd:"api_enabled"`     // custom name
	ExperimentalUI bool `dd:"experimental_ui"` // custom name
}

// DemoContainer for testing all configurations
type DemoContainer struct {
	API     APIConfiguration `dd:"api_config"`
	User    UserProfile      `dd:"user_profile"`
	System  SystemSettings   `dd:"system_settings"`
	Logging LoggingConfig    `dd:"logging"`
}

func main() {
	fmt.Println("=== df struct tag features example ===")
	fmt.Println("demonstrates all supported struct tag features:")
	fmt.Println("custom field names, required fields, secret fields, and field exclusion")

	// step 1: demonstrate required field validation
	fmt.Println("\n=== step 1: required field validation ===")
	incompleteData := map[string]any{
		"host": "api.example.com",
		"port": 8080,
		// missing required service_name and version
	}

	var incomplete APIConfiguration
	if err := dd.Bind(&incomplete, incompleteData); err != nil {
		fmt.Printf("✓ expected error for missing required fields: %v\n", err)
	}

	// step 2: demonstrate successful binding with all required fields
	fmt.Println("\n=== step 2: successful binding with custom field names ===")
	completeData := map[string]any{
		"service_name": "user-api",        // maps to ServiceName field
		"version":      "1.2.3",           // maps to Version field
		"host":         "api.example.com", // maps to Host field
		"port":         8080,              // maps to Port field
		"base_path":    "/api/v1",         // maps to BasePath field
		"api_key":      "secret123",       // maps to APIKey field (secret)
		"debug":        true,              // maps to DebugMode field
		"timeout":      "30s",             // maps to Timeout field
	}

	var api APIConfiguration
	if err := dd.Bind(&api, completeData); err != nil {
		log.Fatalf("failed to bind complete API config: %v", err)
	}

	fmt.Printf("✓ bound API configuration:\n")
	fmt.Printf("  service: %s v%s\n", api.ServiceName, api.Version)
	fmt.Printf("  endpoint: %s:%d%s\n", api.Host, api.Port, api.BasePath)
	fmt.Printf("  debug mode: %t\n", api.DebugMode)
	fmt.Printf("  timeout: %s\n", api.Timeout)
	// note: api_key (secret field) not displayed

	// step 3: demonstrate field exclusion
	fmt.Println("\n=== step 3: field exclusion with df:\"-\" ===")
	api.Internal = "this should not be bound or unbound"

	unbound, err := dd.Unbind(api)
	if err != nil {
		log.Fatalf("failed to unbind API config: %v", err)
	}

	if _, exists := unbound["internal"]; exists {
		fmt.Printf("✗ internal field was unexpectedly included\n")
	} else {
		fmt.Printf("✓ internal field correctly excluded from unbinding\n")
	}

	// step 4: demonstrate secret field handling with inspection
	fmt.Println("\n=== step 4: secret field handling ===")
	inspectData, err := dd.Inspect(api)
	if err != nil {
		log.Fatalf("failed to inspect API config: %v", err)
	}

	fmt.Printf("inspected fields (secrets hidden by default):\n%s\n", inspectData)

	// step 5: demonstrate complex nested structure
	fmt.Println("\n=== step 5: complex nested structure with multiple tag types ===")
	complexData := map[string]any{
		"api_config": map[string]any{
			"service_name": "complex-service",
			"version":      "2.0.0",
			"host":         "localhost",
			"port":         9000,
			"api_key":      "super-secret-key",
			"debug":        false,
			"timeout":      "45s",
		},
		"user_profile": map[string]any{
			"username":     "testuser",
			"email":        "test@example.com",
			"display_name": "Test User",
			"first_name":   "Test",
			"last_name":    "User",
			"phone":        "555-0123",
			"password":     "secretpassword",
			"ssn":          "123-45-6789",
			"preferences": map[string]any{
				"theme":                "dark",
				"language":             "en",
				"enable_notifications": true,
				"newsletter":           false,
				"private_profile":      true,
				"session_token":        "session123",
			},
		},
		"system_settings": map[string]any{
			"env": "production",
			"database": map[string]any{
				"host":          "db.example.com",
				"port":          5432,
				"database_name": "appdb",
				"enable_ssl":    true,
				"username":      "dbuser",
				"password":      "dbpassword",
			},
		},
		"logging": map[string]any{
			"level":           "info",
			"output":          "stdout",
			"format":          "json",
			"log_file":        "/var/log/app.log",
			"max_size_mb":     100,
			"enable_rotation": true,
			"debug_token":     "debug-secret",
		},
	}

	var container DemoContainer
	if err := dd.Bind(&container, complexData); err != nil {
		log.Fatalf("failed to bind complex structure: %v", err)
	}

	fmt.Printf("✓ bound complex nested structure:\n")
	fmt.Printf("  api: %s v%s (%s:%d)\n",
		container.API.ServiceName, container.API.Version,
		container.API.Host, container.API.Port)
	fmt.Printf("  user: %s (%s)\n",
		container.User.Username, container.User.Email)
	fmt.Printf("  environment: %s\n", container.System.Environment)
	fmt.Printf("  database: %s@%s:%d (ssl: %t)\n",
		container.System.Database.Username, container.System.Database.Host,
		container.System.Database.Port, container.System.Database.SSL)

	// step 6: demonstrate flag combination examples
	fmt.Println("\n=== step 6: flag combination examples ===")
	type FlagDemo struct {
		Field1 string `dd:"custom_name"`              // custom name only
		Field2 string `dd:",+secret"`                 // secret only
		Field3 string `dd:",+required"`               // required only
		Field4 string `dd:",+required"`               // required only
		Field5 string `dd:"field5,+secret"`           // custom name + secret
		Field6 string `dd:"field6,+required"`         // custom name + required
		Field7 string `dd:"field7,+required,+secret"` // custom name + required + secret
		Field8 string `dd:",+required,+secret"`       // default name + required + secret
		Field9 string `dd:"-"`                        // excluded
	}

	flagData := map[string]any{
		"custom_name": "value1",
		"field2":      "secret_value",
		"field3":      "required_value",
		"field4":      "another_required",
		"field5":      "secret_with_custom_name",
		"field6":      "required_with_custom_name",
		"field7":      "all_flags_combined",
		"field8":      "default_name_with_flags",
	}

	var demo FlagDemo
	if err := dd.Bind(&demo, flagData); err != nil {
		log.Fatalf("failed to bind flag demo: %v", err)
	}

	// unbind to see field name mapping
	unboundDemo, err := dd.Unbind(demo)
	if err != nil {
		log.Fatalf("failed to unbind flag demo: %v", err)
	}

	fmt.Printf("field name mappings:\n")
	for field, value := range unboundDemo {
		fmt.Printf("  %s: %v\n", field, value)
	}

	fmt.Println("\n=== struct tag features summary ===")
	fmt.Println("✓ custom field names: `dd:\"custom_name\"`")
	fmt.Println("✓ required fields: `dd:\",+required\"`")
	fmt.Println("✓ secret fields: `dd:\",+secret\"` (hidden in inspect)")
	fmt.Println("✓ field exclusion: `dd:\"-\"` (not bound/unbound)")
	fmt.Println("✓ multiple flags: `dd:\"name,+required,+secret\"`")
	fmt.Println("✓ nested struct support: tags apply recursively")

	fmt.Println("\n=== struct tag features example completed successfully! ===")
}
