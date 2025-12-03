package dd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type IOTestStruct struct {
	Name  string `dd:"name"`
	Age   int    `dd:"age"`
	Email string `dd:"email"`
}

// --- Bytes Layer Tests ---

func TestBindJSON(t *testing.T) {
	jsonContent := []byte(`{
		"name": "John Doe",
		"age": 30,
		"email": "john@example.com"
	}`)

	var result IOTestStruct
	if err := BindJSON(&result, jsonContent); err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}

	if result.Name != "John Doe" {
		t.Errorf("expected Name='John Doe', got '%s'", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("expected Age=30, got %d", result.Age)
	}
	if result.Email != "john@example.com" {
		t.Errorf("expected Email='john@example.com', got '%s'", result.Email)
	}
}

func TestBindJSONFromString(t *testing.T) {
	jsonStr := `{"name": "String User", "age": 25, "email": "string@example.com"}`

	var result IOTestStruct
	if err := BindJSON(&result, []byte(jsonStr)); err != nil {
		t.Fatalf("BindJSON from string failed: %v", err)
	}

	if result.Name != "String User" {
		t.Errorf("expected Name='String User', got '%s'", result.Name)
	}
}

func TestBindJSONInvalid(t *testing.T) {
	invalidJSON := []byte(`{"invalid": json`)

	var result IOTestStruct
	err := BindJSON(&result, invalidJSON)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	var convErr *ConversionError
	if !errors.As(err, &convErr) {
		t.Errorf("expected ConversionError, got %T", err)
	}
}

func TestBindYAML(t *testing.T) {
	yamlContent := []byte(`name: Jane Doe
age: 25
email: jane@example.com
`)

	var result IOTestStruct
	if err := BindYAML(&result, yamlContent); err != nil {
		t.Fatalf("BindYAML failed: %v", err)
	}

	if result.Name != "Jane Doe" {
		t.Errorf("expected Name='Jane Doe', got '%s'", result.Name)
	}
	if result.Age != 25 {
		t.Errorf("expected Age=25, got %d", result.Age)
	}
	if result.Email != "jane@example.com" {
		t.Errorf("expected Email='jane@example.com', got '%s'", result.Email)
	}
}

func TestNewJSON(t *testing.T) {
	jsonContent := []byte(`{
		"name": "New User",
		"age": 22,
		"email": "new@example.com"
	}`)

	result, err := NewJSON[IOTestStruct](jsonContent)
	if err != nil {
		t.Fatalf("NewJSON failed: %v", err)
	}

	if result.Name != "New User" {
		t.Errorf("expected Name='New User', got '%s'", result.Name)
	}
	if result.Age != 22 {
		t.Errorf("expected Age=22, got %d", result.Age)
	}
}

func TestNewYAML(t *testing.T) {
	yamlContent := []byte(`name: New YAML User
age: 27
email: newyaml@example.com
`)

	result, err := NewYAML[IOTestStruct](yamlContent)
	if err != nil {
		t.Fatalf("NewYAML failed: %v", err)
	}

	if result.Name != "New YAML User" {
		t.Errorf("expected Name='New YAML User', got '%s'", result.Name)
	}
	if result.Age != 27 {
		t.Errorf("expected Age=27, got %d", result.Age)
	}
}

func TestMergeJSON(t *testing.T) {
	target := IOTestStruct{
		Name: "Original Name",
		Age:  20,
	}

	jsonContent := []byte(`{
		"age": 35,
		"email": "merged@example.com"
	}`)

	if err := MergeJSON(&target, jsonContent); err != nil {
		t.Fatalf("MergeJSON failed: %v", err)
	}

	if target.Name != "Original Name" {
		t.Errorf("expected Name='Original Name', got '%s'", target.Name)
	}
	if target.Age != 35 {
		t.Errorf("expected Age=35, got %d", target.Age)
	}
	if target.Email != "merged@example.com" {
		t.Errorf("expected Email='merged@example.com', got '%s'", target.Email)
	}
}

func TestMergeYAML(t *testing.T) {
	target := IOTestStruct{
		Name:  "Original YAML Name",
		Email: "original@example.com",
	}

	yamlContent := []byte(`age: 40
email: mergedyaml@example.com
`)

	if err := MergeYAML(&target, yamlContent); err != nil {
		t.Fatalf("MergeYAML failed: %v", err)
	}

	if target.Name != "Original YAML Name" {
		t.Errorf("expected Name='Original YAML Name', got '%s'", target.Name)
	}
	if target.Age != 40 {
		t.Errorf("expected Age=40, got %d", target.Age)
	}
	if target.Email != "mergedyaml@example.com" {
		t.Errorf("expected Email='mergedyaml@example.com', got '%s'", target.Email)
	}
}

func TestUnbindJSON(t *testing.T) {
	source := IOTestStruct{
		Name:  "Bob Smith",
		Age:   35,
		Email: "bob@example.com",
	}

	data, err := UnbindJSON(source)
	if err != nil {
		t.Fatalf("UnbindJSON failed: %v", err)
	}

	// verify round-trip
	var result IOTestStruct
	if err := BindJSON(&result, data); err != nil {
		t.Fatalf("failed to read back JSON: %v", err)
	}

	if result.Name != source.Name {
		t.Errorf("expected Name='%s', got '%s'", source.Name, result.Name)
	}
	if result.Age != source.Age {
		t.Errorf("expected Age=%d, got %d", source.Age, result.Age)
	}
}

func TestUnbindYAML(t *testing.T) {
	source := IOTestStruct{
		Name:  "Alice Johnson",
		Age:   28,
		Email: "alice@example.com",
	}

	data, err := UnbindYAML(source)
	if err != nil {
		t.Fatalf("UnbindYAML failed: %v", err)
	}

	// verify round-trip
	var result IOTestStruct
	if err := BindYAML(&result, data); err != nil {
		t.Fatalf("failed to read back YAML: %v", err)
	}

	if result.Name != source.Name {
		t.Errorf("expected Name='%s', got '%s'", source.Name, result.Name)
	}
	if result.Age != source.Age {
		t.Errorf("expected Age=%d, got %d", source.Age, result.Age)
	}
}

// --- Reader/Writer Layer Tests ---

func TestBindJSONReader(t *testing.T) {
	jsonContent := `{"name": "Reader User", "age": 33, "email": "reader@example.com"}`
	r := strings.NewReader(jsonContent)

	var result IOTestStruct
	if err := BindJSONReader(&result, r); err != nil {
		t.Fatalf("BindJSONReader failed: %v", err)
	}

	if result.Name != "Reader User" {
		t.Errorf("expected Name='Reader User', got '%s'", result.Name)
	}
	if result.Age != 33 {
		t.Errorf("expected Age=33, got %d", result.Age)
	}
}

func TestBindYAMLReader(t *testing.T) {
	yamlContent := `name: YAML Reader User
age: 44
email: yamlreader@example.com
`
	r := strings.NewReader(yamlContent)

	var result IOTestStruct
	if err := BindYAMLReader(&result, r); err != nil {
		t.Fatalf("BindYAMLReader failed: %v", err)
	}

	if result.Name != "YAML Reader User" {
		t.Errorf("expected Name='YAML Reader User', got '%s'", result.Name)
	}
	if result.Age != 44 {
		t.Errorf("expected Age=44, got %d", result.Age)
	}
}

func TestNewJSONReader(t *testing.T) {
	jsonContent := `{"name": "New Reader", "age": 55, "email": "newreader@example.com"}`
	r := strings.NewReader(jsonContent)

	result, err := NewJSONReader[IOTestStruct](r)
	if err != nil {
		t.Fatalf("NewJSONReader failed: %v", err)
	}

	if result.Name != "New Reader" {
		t.Errorf("expected Name='New Reader', got '%s'", result.Name)
	}
}

func TestNewYAMLReader(t *testing.T) {
	yamlContent := `name: New YAML Reader
age: 66
email: newyamlreader@example.com
`
	r := strings.NewReader(yamlContent)

	result, err := NewYAMLReader[IOTestStruct](r)
	if err != nil {
		t.Fatalf("NewYAMLReader failed: %v", err)
	}

	if result.Name != "New YAML Reader" {
		t.Errorf("expected Name='New YAML Reader', got '%s'", result.Name)
	}
}

func TestMergeJSONReader(t *testing.T) {
	target := IOTestStruct{
		Name: "Original Reader Name",
		Age:  10,
	}

	r := strings.NewReader(`{"age": 77, "email": "mergereader@example.com"}`)

	if err := MergeJSONReader(&target, r); err != nil {
		t.Fatalf("MergeJSONReader failed: %v", err)
	}

	if target.Name != "Original Reader Name" {
		t.Errorf("expected Name='Original Reader Name', got '%s'", target.Name)
	}
	if target.Age != 77 {
		t.Errorf("expected Age=77, got %d", target.Age)
	}
}

func TestMergeYAMLReader(t *testing.T) {
	target := IOTestStruct{
		Name:  "Original YAML Reader Name",
		Email: "original@example.com",
	}

	r := strings.NewReader(`age: 88
email: mergeyamlreader@example.com
`)

	if err := MergeYAMLReader(&target, r); err != nil {
		t.Fatalf("MergeYAMLReader failed: %v", err)
	}

	if target.Name != "Original YAML Reader Name" {
		t.Errorf("expected Name='Original YAML Reader Name', got '%s'", target.Name)
	}
	if target.Age != 88 {
		t.Errorf("expected Age=88, got %d", target.Age)
	}
}

func TestUnbindJSONWriter(t *testing.T) {
	source := IOTestStruct{
		Name:  "Writer User",
		Age:   99,
		Email: "writer@example.com",
	}

	var buf bytes.Buffer
	if err := UnbindJSONWriter(source, &buf); err != nil {
		t.Fatalf("UnbindJSONWriter failed: %v", err)
	}

	// verify round-trip
	var result IOTestStruct
	if err := BindJSON(&result, buf.Bytes()); err != nil {
		t.Fatalf("failed to read back JSON: %v", err)
	}

	if result.Name != source.Name {
		t.Errorf("expected Name='%s', got '%s'", source.Name, result.Name)
	}
}

func TestUnbindYAMLWriter(t *testing.T) {
	source := IOTestStruct{
		Name:  "YAML Writer User",
		Age:   111,
		Email: "yamlwriter@example.com",
	}

	var buf bytes.Buffer
	if err := UnbindYAMLWriter(source, &buf); err != nil {
		t.Fatalf("UnbindYAMLWriter failed: %v", err)
	}

	// verify round-trip
	var result IOTestStruct
	if err := BindYAML(&result, buf.Bytes()); err != nil {
		t.Fatalf("failed to read back YAML: %v", err)
	}

	if result.Name != source.Name {
		t.Errorf("expected Name='%s', got '%s'", source.Name, result.Name)
	}
}

// --- File Layer Tests ---

func TestBindJSONFile(t *testing.T) {
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

	var result IOTestStruct
	if err := BindJSONFile(&result, jsonFile); err != nil {
		t.Fatalf("BindJSONFile failed: %v", err)
	}

	if result.Name != "John Doe" {
		t.Errorf("expected Name='John Doe', got '%s'", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("expected Age=30, got %d", result.Age)
	}
	if result.Email != "john@example.com" {
		t.Errorf("expected Email='john@example.com', got '%s'", result.Email)
	}
}

func TestBindYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `name: Jane Doe
age: 25
email: jane@example.com
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML file: %v", err)
	}

	var result IOTestStruct
	if err := BindYAMLFile(&result, yamlFile); err != nil {
		t.Fatalf("BindYAMLFile failed: %v", err)
	}

	if result.Name != "Jane Doe" {
		t.Errorf("expected Name='Jane Doe', got '%s'", result.Name)
	}
	if result.Age != 25 {
		t.Errorf("expected Age=25, got %d", result.Age)
	}
	if result.Email != "jane@example.com" {
		t.Errorf("expected Email='jane@example.com', got '%s'", result.Email)
	}
}

func TestUnbindJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "output.json")

	source := IOTestStruct{
		Name:  "Bob Smith",
		Age:   35,
		Email: "bob@example.com",
	}

	if err := UnbindJSONFile(source, jsonFile); err != nil {
		t.Fatalf("UnbindJSONFile failed: %v", err)
	}

	// read back and verify
	var result IOTestStruct
	if err := BindJSONFile(&result, jsonFile); err != nil {
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

func TestUnbindYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "output.yaml")

	source := IOTestStruct{
		Name:  "Alice Johnson",
		Age:   28,
		Email: "alice@example.com",
	}

	if err := UnbindYAMLFile(source, yamlFile); err != nil {
		t.Fatalf("UnbindYAMLFile failed: %v", err)
	}

	// read back and verify
	var result IOTestStruct
	if err := BindYAMLFile(&result, yamlFile); err != nil {
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

func TestBindJSONFileNotFound(t *testing.T) {
	var result IOTestStruct
	err := BindJSONFile(&result, "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestBindYAMLFileNotFound(t *testing.T) {
	var result IOTestStruct
	err := BindYAMLFile(&result, "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestNewJSONFile(t *testing.T) {
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

	result, err := NewJSONFile[IOTestStruct](jsonFile)
	if err != nil {
		t.Fatalf("NewJSONFile failed: %v", err)
	}

	if result.Name != "New User" {
		t.Errorf("expected Name='New User', got '%s'", result.Name)
	}
	if result.Age != 22 {
		t.Errorf("expected Age=22, got %d", result.Age)
	}
	if result.Email != "new@example.com" {
		t.Errorf("expected Email='new@example.com', got '%s'", result.Email)
	}
}

func TestNewYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yaml")

	yamlContent := `name: New YAML User
age: 27
email: newyaml@example.com
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML file: %v", err)
	}

	result, err := NewYAMLFile[IOTestStruct](yamlFile)
	if err != nil {
		t.Fatalf("NewYAMLFile failed: %v", err)
	}

	if result.Name != "New YAML User" {
		t.Errorf("expected Name='New YAML User', got '%s'", result.Name)
	}
	if result.Age != 27 {
		t.Errorf("expected Age=27, got %d", result.Age)
	}
	if result.Email != "newyaml@example.com" {
		t.Errorf("expected Email='newyaml@example.com', got '%s'", result.Email)
	}
}

func TestNewJSONFileNotFound(t *testing.T) {
	_, err := NewJSONFile[IOTestStruct]("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestNewYAMLFileNotFound(t *testing.T) {
	_, err := NewYAMLFile[IOTestStruct]("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestMergeJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "merge.json")

	target := IOTestStruct{
		Name: "Original Name",
		Age:  20,
	}

	jsonContent := `{
		"age": 35,
		"email": "merged@example.com"
	}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test JSON file: %v", err)
	}

	if err := MergeJSONFile(&target, jsonFile); err != nil {
		t.Fatalf("MergeJSONFile failed: %v", err)
	}

	if target.Name != "Original Name" {
		t.Errorf("expected Name='Original Name', got '%s'", target.Name)
	}
	if target.Age != 35 {
		t.Errorf("expected Age=35, got %d", target.Age)
	}
	if target.Email != "merged@example.com" {
		t.Errorf("expected Email='merged@example.com', got '%s'", target.Email)
	}
}

func TestMergeYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "merge.yaml")

	target := IOTestStruct{
		Name:  "Original YAML Name",
		Email: "original@example.com",
	}

	yamlContent := `age: 40
email: mergedyaml@example.com
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML file: %v", err)
	}

	if err := MergeYAMLFile(&target, yamlFile); err != nil {
		t.Fatalf("MergeYAMLFile failed: %v", err)
	}

	if target.Name != "Original YAML Name" {
		t.Errorf("expected Name='Original YAML Name', got '%s'", target.Name)
	}
	if target.Age != 40 {
		t.Errorf("expected Age=40, got %d", target.Age)
	}
	if target.Email != "mergedyaml@example.com" {
		t.Errorf("expected Email='mergedyaml@example.com', got '%s'", target.Email)
	}
}

func TestMergeJSONFileNotFound(t *testing.T) {
	var target IOTestStruct
	err := MergeJSONFile(&target, "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}

func TestMergeYAMLFileNotFound(t *testing.T) {
	var target IOTestStruct
	err := MergeYAMLFile(&target, "/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	var fileErr *FileError
	if !errors.As(err, &fileErr) {
		t.Errorf("expected FileError, got %T", err)
	}
}
