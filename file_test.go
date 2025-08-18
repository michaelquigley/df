package df

import (
	"errors"
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
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestBindFromYAMLFileNotFound(t *testing.T) {
	var result TestStruct
	err := BindFromYAML(&result, "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestNewFromJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	jsonContent := `{
		"name": "New User",
		"age": 22,
		"email": "new@example.com"
	}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test JSON file: %v", err)
	}

	result, err := NewFromJSON[TestStruct](jsonFile)
	if err != nil {
		t.Fatalf("NewFromJSON failed: %v", err)
	}

	if result.Name != "New User" {
		t.Errorf("Expected Name='New User', got '%s'", result.Name)
	}
	if result.Age != 22 {
		t.Errorf("Expected Age=22, got %d", result.Age)
	}
	if result.Email != "new@example.com" {
		t.Errorf("Expected Email='new@example.com', got '%s'", result.Email)
	}
}

func TestNewFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `name: New YAML User
age: 27
email: newyaml@example.com
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML file: %v", err)
	}

	result, err := NewFromYAML[TestStruct](yamlFile)
	if err != nil {
		t.Fatalf("NewFromYAML failed: %v", err)
	}

	if result.Name != "New YAML User" {
		t.Errorf("Expected Name='New YAML User', got '%s'", result.Name)
	}
	if result.Age != 27 {
		t.Errorf("Expected Age=27, got %d", result.Age)
	}
	if result.Email != "newyaml@example.com" {
		t.Errorf("Expected Email='newyaml@example.com', got '%s'", result.Email)
	}
}

func TestNewFromJSONFileNotFound(t *testing.T) {
	_, err := NewFromJSON[TestStruct]("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestNewFromYAMLFileNotFound(t *testing.T) {
	_, err := NewFromYAML[TestStruct]("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestMergeFromJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "merge.json")

	// start with a partial struct
	target := TestStruct{
		Name: "Original Name",
		Age:  20,
	}

	// JSON with partial data to merge
	jsonContent := `{
		"age": 35,
		"email": "merged@example.com"
	}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test JSON file: %v", err)
	}

	if err := MergeFromJSON(&target, jsonFile); err != nil {
		t.Fatalf("MergeFromJSON failed: %v", err)
	}

	// name should remain unchanged
	if target.Name != "Original Name" {
		t.Errorf("Expected Name='Original Name', got '%s'", target.Name)
	}
	// age should be updated
	if target.Age != 35 {
		t.Errorf("Expected Age=35, got %d", target.Age)
	}
	// email should be set
	if target.Email != "merged@example.com" {
		t.Errorf("Expected Email='merged@example.com', got '%s'", target.Email)
	}
}

func TestMergeFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "merge.yaml")

	// start with a partial struct
	target := TestStruct{
		Name:  "Original YAML Name",
		Email: "original@example.com",
	}

	// YAML with partial data to merge
	yamlContent := `age: 40
email: mergedyaml@example.com
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML file: %v", err)
	}

	if err := MergeFromYAML(&target, yamlFile); err != nil {
		t.Fatalf("MergeFromYAML failed: %v", err)
	}

	// name should remain unchanged
	if target.Name != "Original YAML Name" {
		t.Errorf("Expected Name='Original YAML Name', got '%s'", target.Name)
	}
	// age should be set
	if target.Age != 40 {
		t.Errorf("Expected Age=40, got %d", target.Age)
	}
	// email should be updated
	if target.Email != "mergedyaml@example.com" {
		t.Errorf("Expected Email='mergedyaml@example.com', got '%s'", target.Email)
	}
}

func TestMergeFromJSONFileNotFound(t *testing.T) {
	var target TestStruct
	err := MergeFromJSON(&target, "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestMergeFromYAMLFileNotFound(t *testing.T) {
	var target TestStruct
	err := MergeFromYAML(&target, "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}
