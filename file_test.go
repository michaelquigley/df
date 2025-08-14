package df

import (
	"os"
	"path/filepath"
	"testing"
)

type TestStruct struct {
	Name  string `df:"name"`
	Age   int    `df:"age"`
	Email string `df:"email"`
}

func TestBindFromJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	jsonContent := `{
		"name": "John Doe",
		"age": 30,
		"email": "john@example.com"
	}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test JSON file: %v", err)
	}

	var result TestStruct
	if err := BindFromJSON(&result, jsonFile); err != nil {
		t.Fatalf("BindFromJSON failed: %v", err)
	}

	if result.Name != "John Doe" {
		t.Errorf("Expected Name='John Doe', got '%s'", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("Expected Age=30, got %d", result.Age)
	}
	if result.Email != "john@example.com" {
		t.Errorf("Expected Email='john@example.com', got '%s'", result.Email)
	}
}

func TestBindFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `name: Jane Doe
age: 25
email: jane@example.com
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML file: %v", err)
	}

	var result TestStruct
	if err := BindFromYAML(&result, yamlFile); err != nil {
		t.Fatalf("BindFromYAML failed: %v", err)
	}

	if result.Name != "Jane Doe" {
		t.Errorf("Expected Name='Jane Doe', got '%s'", result.Name)
	}
	if result.Age != 25 {
		t.Errorf("Expected Age=25, got %d", result.Age)
	}
	if result.Email != "jane@example.com" {
		t.Errorf("Expected Email='jane@example.com', got '%s'", result.Email)
	}
}

func TestUnbindToJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "output.json")

	source := TestStruct{
		Name:  "Bob Smith",
		Age:   35,
		Email: "bob@example.com",
	}

	if err := UnbindToJSON(source, jsonFile); err != nil {
		t.Fatalf("UnbindToJSON failed: %v", err)
	}

	// read back and verify
	var result TestStruct
	if err := BindFromJSON(&result, jsonFile); err != nil {
		t.Fatalf("failed to read back JSON: %v", err)
	}

	if result.Name != source.Name {
		t.Errorf("expected Name='%s', got '%s'", source.Name, result.Name)
	}
	if result.Age != source.Age {
		t.Errorf("expected Age=%d, got %d", source.Age, result.Age)
	}
	if result.Email != source.Email {
		t.Errorf("expected Email='%s', got '%s'", source.Email, result.Email)
	}
}

func TestUnbindToYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "output.yaml")

	source := TestStruct{
		Name:  "Alice Johnson",
		Age:   28,
		Email: "alice@example.com",
	}

	if err := UnbindToYAML(source, yamlFile); err != nil {
		t.Fatalf("UnbindToYAML failed: %v", err)
	}

	// read back and verify
	var result TestStruct
	if err := BindFromYAML(&result, yamlFile); err != nil {
		t.Fatalf("failed to read back YAML: %v", err)
	}

	if result.Name != source.Name {
		t.Errorf("expected Name='%s', got '%s'", source.Name, result.Name)
	}
	if result.Age != source.Age {
		t.Errorf("expected Age=%d, got %d", source.Age, result.Age)
	}
	if result.Email != source.Email {
		t.Errorf("expected Email='%s', got '%s'", source.Email, result.Email)
	}
}

func TestBindFromJSONFileNotFound(t *testing.T) {
	var result TestStruct
	err := BindFromJSON(&result, "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestBindFromYAMLFileNotFound(t *testing.T) {
	var result TestStruct
	err := BindFromYAML(&result, "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
