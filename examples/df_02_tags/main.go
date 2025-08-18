package main

import (
	"fmt"
	"log"
	"time"

	"github.com/michaelquigley/df"
)

// APIConfiguration demonstrates field naming and requirements
type APIConfiguration struct {
	ServiceName string        `df:"service_name,required"` // custom name + required
	Version     string        `df:",required"`             // default name + required
	Host        string        // default snake_case: "host"
	Port        int           // default snake_case: "port"
	BasePath    string        `df:"base_path"` // custom name only
	APIKey      string        `df:",secret"`   // custom name + secret
	DebugMode   bool          `df:"debug"`     // custom name only
	Timeout     time.Duration // default snake_case: "timeout"
	Internal    string        `df:"-"` // excluded from binding/unbinding
}

// UserProfile demonstrates privacy controls and field naming
type UserProfile struct {
	Username    string           `df:",required"`    // required with default name
	Email       string           `df:",required"`    // required email
	DisplayName string           `df:"display_name"` // custom field name
	FirstName   string           `df:"first_name"`   // custom field name
	LastName    string           `df:"last_name"`    // custom field name
	Phone       string           // default name: "phone"
	Password    string           `df:",secret"`    // secret field
	SSN         string           `df:"ssn,secret"` // custom name + secret
	Preferences *UserPreferences // nested struct
}

// UserPreferences demonstrates nested struct tag behavior
type UserPreferences struct {
	Theme           string // default: "theme"
	Language        string // default: "language"
	Notifications   bool   `df:"enable_notifications"` // custom name
	Newsletter      bool   // default: "newsletter"
	PrivateProfile  bool   `df:"private_profile"` // custom name
	SessionToken    string `df:",secret"`         // secret in nested struct
	InternalSetting string `df:"-"`               // excluded
}

// SystemSettings demonstrates hierarchical configuration
type SystemSettings struct {
	Environment string          `df:"env,required"` // required environment
	Database    *DatabaseConfig // nested config
	Cache       *CacheConfig    // optional nested struct
	Features    *FeatureFlags   // feature toggles
}

// DatabaseConfig with security and naming
type DatabaseConfig struct {
	Host     string `df:",required"` // required host
	Port     int    // default: "port"
	Name     string `df:"database_name,required"` // custom name + required
	SSL      bool   `df:"enable_ssl"`             // custom name
	Username string `df:",required"`              // required username
	Password string `df:",secret"`                // secret password
}

// CacheConfig with provider flexibility
type CacheConfig struct {
	Provider  string        `df:",required"` // required provider
	Host      string        // default: "host"
	Port      int           // default: "port"
	Password  string        `df:",secret"`     // secret password
	TTL       time.Duration `df:"default_ttl"` // custom name
	MaxMemory int64         `df:"max_memory"`  // custom name
}

// LoggingConfig demonstrates complex naming patterns
type LoggingConfig struct {
	Level     string `df:",required"` // required log level
	Output    string // default: "output"
	Format    string // default: "format"
	Filename  string `df:"log_file"`           // custom name
	MaxSize   int    `df:"max_size_mb"`        // custom name
	Rotate    bool   `df:"enable_rotation"`    // custom name
	Sensitive string `df:"debug_token,secret"` // secret + custom name
}

// FeatureFlags for system capabilities
type FeatureFlags struct {
	EnableBeta     bool `df:"beta_features"`       // custom name
	EnableMetrics  bool `df:"metrics_collection"`  // custom name
	EnableTracing  bool `df:"distributed_tracing"` // custom name
	EnableDebug    bool // default: "enable_debug"
	EnableAPI      bool `df:"api_enabled"`     // custom name
	ExperimentalUI bool `df:"experimental_ui"` // custom name
}

// DemoContainer for testing all configurations
type DemoContainer struct {
	API     APIConfiguration `df:"api_config"`
	User    UserProfile      `df:"user_profile"`
	System  SystemSettings   `df:"system_settings"`
	Logging LoggingConfig    `df:"logging"`
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
	if err := df.Bind(&incomplete, incompleteData); err != nil {
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
	if err := df.Bind(&api, completeData); err != nil {
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

	unbound, err := df.Unbind(api)
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
	inspectData, err := df.Inspect(api)
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
	if err := df.Bind(&container, complexData); err != nil {
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
		Field1 string `df:"custom_name"`            // custom name only
		Field2 string `df:",secret"`                // secret only
		Field3 string `df:",required"`              // required only
		Field4 string `df:",required"`              // required only
		Field5 string `df:"field5,secret"`          // custom name + secret
		Field6 string `df:"field6,required"`        // custom name + required
		Field7 string `df:"field7,required,secret"` // custom name + required + secret
		Field8 string `df:",required,secret"`       // default name + required + secret
		Field9 string `df:"-"`                      // excluded
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
	if err := df.Bind(&demo, flagData); err != nil {
		log.Fatalf("failed to bind flag demo: %v", err)
	}

	// unbind to see field name mapping
	unboundDemo, err := df.Unbind(demo)
	if err != nil {
		log.Fatalf("failed to unbind flag demo: %v", err)
	}

	fmt.Printf("field name mappings:\n")
	for field, value := range unboundDemo {
		fmt.Printf("  %s: %v\n", field, value)
	}

	fmt.Println("\n=== struct tag features summary ===")
	fmt.Println("✓ custom field names: `df:\"custom_name\"`")
	fmt.Println("✓ required fields: `df:\",required\"`")
	fmt.Println("✓ secret fields: `df:\",secret\"` (hidden in inspect)")
	fmt.Println("✓ field exclusion: `df:\"-\"` (not bound/unbound)")
	fmt.Println("✓ multiple flags: `df:\"name,required,secret\"`")
	fmt.Println("✓ nested struct support: tags apply recursively")

	fmt.Println("\n=== struct tag features example completed successfully! ===")
}
