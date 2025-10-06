package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df/dd"
)

// ServerConfig represents a single server's configuration
type ServerConfig struct {
	Host string `dd:"host"`
	Port int    `dd:"port"`
	Role string `dd:"role"`
}

// CachePolicy represents caching configuration
type CachePolicy struct {
	TTL     int  `dd:"ttl"`
	Enabled bool `dd:"enabled"`
}

// User represents a user in the system
type User struct {
	ID    int    `dd:"id"`
	Name  string `dd:"name"`
	Email string `dd:"email"`
}

// ClusterConfig demonstrates various typed map configurations
type ClusterConfig struct {
	Name    string                   `dd:"name"`
	Servers map[int]ServerConfig     `dd:"servers"`      // int keys for server IDs
	Cache   map[string]CachePolicy   `dd:"cache"`        // string keys for cache names
	Users   map[int]*User            `dd:"users"`        // pointer values
	Groups  map[string][]string      `dd:"groups"`       // slice values
	Envs    map[string]map[string]string `dd:"envs"`    // nested maps
	Flags   map[bool]string          `dd:"flags"`        // bool keys
}

func main() {
	fmt.Println("=== df.dd typed maps example ===")
	fmt.Println("demonstrates map[K]V support with various key and value types")
	fmt.Println()

	// example configuration data with typed maps
	// note: JSON/YAML always have string keys, they are coerced to target types
	configData := map[string]any{
		"name": "production-cluster",
		"servers": map[string]any{
			"1":  map[string]any{"host": "server1.example.com", "port": 8080, "role": "primary"},
			"2":  map[string]any{"host": "server2.example.com", "port": 8080, "role": "replica"},
			"10": map[string]any{"host": "server10.example.com", "port": 8081, "role": "backup"},
		},
		"cache": map[string]any{
			"users":    map[string]any{"ttl": 300, "enabled": true},
			"sessions": map[string]any{"ttl": 600, "enabled": true},
			"products": map[string]any{"ttl": 1800, "enabled": false},
		},
		"users": map[string]any{
			"1001": map[string]any{"id": 1001, "name": "Alice", "email": "alice@example.com"},
			"1002": map[string]any{"id": 1002, "name": "Bob", "email": "bob@example.com"},
			"1003": map[string]any{"id": 1003, "name": "Charlie", "email": "charlie@example.com"},
		},
		"groups": map[string]any{
			"admins":     []any{"alice", "bob"},
			"developers": []any{"charlie", "diana", "eve"},
			"viewers":    []any{"frank", "grace"},
		},
		"envs": map[string]any{
			"dev": map[string]any{
				"db_host": "localhost",
				"api_url": "http://localhost:8080",
			},
			"prod": map[string]any{
				"db_host": "db.prod.example.com",
				"api_url": "https://api.example.com",
			},
		},
		"flags": map[string]any{
			"true":  "active",
			"false": "inactive",
		},
	}

	fmt.Println("=== 1. binding typed maps from data ===")
	config, err := dd.New[ClusterConfig](configData)
	if err != nil {
		log.Fatalf("failed to bind config: %v", err)
	}
	fmt.Printf("cluster name: %s\n", config.Name)
	fmt.Println()

	fmt.Println("=== 2. accessing maps with int keys ===")
	fmt.Println("servers (map[int]ServerConfig):")
	for id, server := range config.Servers {
		fmt.Printf("  server %d: %s:%d (%s)\n", id, server.Host, server.Port, server.Role)
	}
	fmt.Println()
	fmt.Printf("direct access: server 1 = %s\n", config.Servers[1].Host)
	fmt.Println()

	fmt.Println("=== 3. accessing maps with string keys ===")
	fmt.Println("cache policies (map[string]CachePolicy):")
	for name, policy := range config.Cache {
		status := "disabled"
		if policy.Enabled {
			status = "enabled"
		}
		fmt.Printf("  %s: ttl=%ds (%s)\n", name, policy.TTL, status)
	}
	fmt.Println()

	fmt.Println("=== 4. maps with pointer values ===")
	fmt.Println("users (map[int]*User):")
	for id, user := range config.Users {
		fmt.Printf("  user %d: %s (%s)\n", id, user.Name, user.Email)
	}
	fmt.Println()
	alice := config.Users[1001]
	fmt.Printf("direct access: user 1001 = %s (%s)\n", alice.Name, alice.Email)
	fmt.Println()

	fmt.Println("=== 5. maps with slice values ===")
	fmt.Println("groups (map[string][]string):")
	for group, members := range config.Groups {
		fmt.Printf("  %s: %v\n", group, members)
	}
	fmt.Println()

	fmt.Println("=== 6. nested maps ===")
	fmt.Println("environments (map[string]map[string]string):")
	for env, vars := range config.Envs {
		fmt.Printf("  %s:\n", env)
		for key, value := range vars {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}
	fmt.Println()

	fmt.Println("=== 7. maps with bool keys ===")
	fmt.Println("flags (map[bool]string):")
	for flag, status := range config.Flags {
		fmt.Printf("  %v: %s\n", flag, status)
	}
	fmt.Println()

	fmt.Println("=== 8. unbinding typed maps back to data ===")
	result, err := dd.Unbind(config)
	if err != nil {
		log.Fatalf("failed to unbind config: %v", err)
	}
	fmt.Println("unbind successful - keys converted to strings for JSON/YAML")

	// demonstrate that int keys became strings
	serversMap := result["servers"].(map[string]any)
	fmt.Printf("server keys in result: ")
	for key := range serversMap {
		fmt.Printf("%q ", key)
	}
	fmt.Println("(all strings)")
	fmt.Println()

	fmt.Println("=== 9. roundtrip: bind → unbind → bind ===")
	var restored ClusterConfig
	err = dd.Bind(&restored, result)
	if err != nil {
		log.Fatalf("failed to rebind config: %v", err)
	}
	fmt.Printf("roundtrip successful!\n")
	fmt.Printf("original server 1: %s\n", config.Servers[1].Host)
	fmt.Printf("restored server 1: %s\n", restored.Servers[1].Host)
	fmt.Println()

	fmt.Println("=== key takeaways ===")
	fmt.Println("• JSON/YAML always have string keys - dd coerces them to target types")
	fmt.Println("• supported key types: string, int*, uint*, float*, bool")
	fmt.Println("• value types: any supported dd type (primitives, structs, pointers, slices, maps)")
	fmt.Println("• unbind converts all keys back to strings for JSON/YAML compatibility")
	fmt.Println("• use typed maps for: ID-based lookups, indexed data, configuration sets")
}
