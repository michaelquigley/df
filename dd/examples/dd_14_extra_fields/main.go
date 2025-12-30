package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df/dd"
)

// AppConfig demonstrates basic extra field capture
type AppConfig struct {
	Name    string `dd:",+required"`
	Version string
	Extra   map[string]any `dd:",+extra"`
}

// ServiceConfig demonstrates nested structs with independent extras
type ServiceConfig struct {
	Host     string `dd:"host,+required"`
	Port     int
	Settings *Settings
	Extra    map[string]any `dd:",+extra"`
}

// Settings is a nested struct that also captures its own extras
type Settings struct {
	Timeout int
	Retries int
	Extra   map[string]any `dd:",+extra"`
}

// BaseInfo is an embedded struct
type BaseInfo struct {
	ID      string `dd:",+required"`
	Created string
}

// Document demonstrates embedded struct behavior with extras
type Document struct {
	BaseInfo // embedded struct shares parent's key namespace
	Title    string
	Extra    map[string]any `dd:",+extra"`
}

// Item demonstrates extra fields in slice elements
type Item struct {
	Name  string
	Extra map[string]any `dd:",+extra"`
}

// Container holds a slice of items with extras
type Container struct {
	Items []Item `dd:"items"`
}

func main() {
	fmt.Println("=== df.dd extra fields example ===")
	fmt.Println("demonstrates the +extra tag for capturing unmatched data keys")
	fmt.Println("during binding and merging them back during unbinding")

	// step 1: basic extra field capture
	fmt.Println("\n=== step 1: basic extra field capture ===")
	appData := map[string]any{
		"name":        "myapp",
		"version":     "1.0.0",
		"author":      "john doe",     // extra - not in struct
		"license":     "MIT",          // extra - not in struct
		"description": "a sample app", // extra - not in struct
	}

	app, err := dd.New[AppConfig](appData)
	if err != nil {
		log.Fatalf("failed to bind app config: %v", err)
	}

	fmt.Printf("bound fields:\n")
	fmt.Printf("  name: %s\n", app.Name)
	fmt.Printf("  version: %s\n", app.Version)
	fmt.Printf("captured extras:\n")
	for k, v := range app.Extra {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// step 2: unbind with extras merged back
	fmt.Println("\n=== step 2: unbind with extras merged back ===")
	unbound, err := dd.Unbind(app)
	if err != nil {
		log.Fatalf("failed to unbind app config: %v", err)
	}

	fmt.Printf("unbound map contains all original keys:\n")
	for k, v := range unbound {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// step 3: nested structs with independent extras
	fmt.Println("\n=== step 3: nested structs with independent extras ===")
	serviceData := map[string]any{
		"host":       "api.example.com",
		"port":       8080,
		"region":     "us-east", // extra at service level
		"datacenter": "dc1",     // extra at service level
		"settings": map[string]any{
			"timeout":       30,
			"retries":       3,
			"debug_mode":    true,  // extra at settings level
			"trace_enabled": false, // extra at settings level
		},
	}

	service, err := dd.New[ServiceConfig](serviceData)
	if err != nil {
		log.Fatalf("failed to bind service config: %v", err)
	}

	fmt.Printf("service level extras: %v\n", service.Extra)
	fmt.Printf("settings level extras: %v\n", service.Settings.Extra)

	// step 4: round-trip preservation
	fmt.Println("\n=== step 4: round-trip preservation ===")
	original := map[string]any{
		"name":         "roundtrip-test",
		"version":      "2.0",
		"custom_field": "preserved",
		"metadata": map[string]any{
			"key1": "value1",
			"key2": float64(42),
		},
	}

	var roundtrip AppConfig
	if err := dd.Bind(&roundtrip, original); err != nil {
		log.Fatalf("failed to bind roundtrip: %v", err)
	}

	result, err := dd.Unbind(roundtrip)
	if err != nil {
		log.Fatalf("failed to unbind roundtrip: %v", err)
	}

	fmt.Printf("original keys preserved after round-trip:\n")
	for k, v := range result {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// step 5: merge behavior with existing extras
	fmt.Println("\n=== step 5: merge behavior with existing extras ===")
	config := &AppConfig{
		Name:    "merge-test",
		Version: "1.0",
		Extra: map[string]any{
			"existing_key": "original_value",
			"keep_this":    "unchanged",
		},
	}

	fmt.Printf("before merge - extras: %v\n", config.Extra)

	mergeData := map[string]any{
		"name":    "merge-test", // required field must be present
		"version": "1.1",        // update existing field
		"new_key": "new_value",  // add new extra
	}

	if err := dd.Merge(config, mergeData); err != nil {
		log.Fatalf("failed to merge: %v", err)
	}

	fmt.Printf("after merge - version: %s\n", config.Version)
	fmt.Printf("after merge - extras: %v\n", config.Extra)

	// step 6: embedded struct behavior
	fmt.Println("\n=== step 6: embedded struct behavior ===")
	docData := map[string]any{
		"id":       "doc-001",
		"created":  "2024-01-15",
		"title":    "example document",
		"author":   "jane smith", // extra - not in BaseInfo or Document
		"category": "tutorial",   // extra - not in BaseInfo or Document
	}

	doc, err := dd.New[Document](docData)
	if err != nil {
		log.Fatalf("failed to bind document: %v", err)
	}

	fmt.Printf("embedded fields (from BaseInfo):\n")
	fmt.Printf("  id: %s\n", doc.ID)
	fmt.Printf("  created: %s\n", doc.Created)
	fmt.Printf("document field:\n")
	fmt.Printf("  title: %s\n", doc.Title)
	fmt.Printf("extras captured at document level:\n")
	for k, v := range doc.Extra {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// step 7: extra fields in slice of structs
	fmt.Println("\n=== step 7: extra fields in slice of structs ===")
	containerData := map[string]any{
		"items": []any{
			map[string]any{"name": "item1", "color": "red", "size": "large"},
			map[string]any{"name": "item2", "weight": 5.5, "material": "wood"},
			map[string]any{"name": "item3", "tags": []any{"new", "featured"}},
		},
	}

	container, err := dd.New[Container](containerData)
	if err != nil {
		log.Fatalf("failed to bind container: %v", err)
	}

	fmt.Printf("each item captures its own extras:\n")
	for i, item := range container.Items {
		fmt.Printf("  item %d (%s): extras = %v\n", i+1, item.Name, item.Extra)
	}

	// step 8: empty extras behavior
	fmt.Println("\n=== step 8: empty extras behavior ===")
	noExtrasData := map[string]any{
		"name":    "no-extras",
		"version": "1.0",
		// no extra keys
	}

	noExtras, err := dd.New[AppConfig](noExtrasData)
	if err != nil {
		log.Fatalf("failed to bind no-extras: %v", err)
	}

	if noExtras.Extra == nil {
		fmt.Printf("when no extra keys exist, Extra field remains nil\n")
	} else {
		fmt.Printf("extras: %v\n", noExtras.Extra)
	}

	fmt.Println("\n=== extra fields summary ===")
	fmt.Println("tag: `dd:\",+extra\"` captures unmatched keys")
	fmt.Println("type: must be map[string]any")
	fmt.Println("bind: unknown keys are collected into the extra field")
	fmt.Println("unbind: extra contents are merged back into output")
	fmt.Println("nested: each struct captures its own extras independently")
	fmt.Println("embedded: embedded struct fields share parent's namespace")
	fmt.Println("merge: new extras are added to existing extra map")

	fmt.Println("\n=== extra fields example completed successfully! ===")
}
