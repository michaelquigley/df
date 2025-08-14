package df

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicBindTo(t *testing.T) {
	config := &struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}

	data := map[string]any{
		"host": "example.com",
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "example.com", config.Host)
	assert.Equal(t, 8080, config.Port) // preserved default
}

func TestBindToPreservesDefaults(t *testing.T) {
	config := &struct {
		Host    string
		Port    int
		Timeout int
		Debug   bool
	}{
		Host:    "localhost",
		Port:    8080,
		Timeout: 30,
		Debug:   false,
	}

	data := map[string]any{
		"port":  9090,
		"debug": true,
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "localhost", config.Host) // preserved
	assert.Equal(t, 9090, config.Port)       // overridden
	assert.Equal(t, 30, config.Timeout)      // preserved
	assert.Equal(t, true, config.Debug)      // overridden
}

func TestBindToNestedStruct(t *testing.T) {
	type Database struct {
		Host string
		Port int
	}

	config := &struct {
		App Database
	}{
		App: Database{
			Host: "localhost",
			Port: 5432,
		},
	}

	data := map[string]any{
		"app": map[string]any{
			"host": "db.example.com",
		},
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "db.example.com", config.App.Host)
	assert.Equal(t, 5432, config.App.Port) // preserved
}

func TestBindToPointerField(t *testing.T) {
	config := &struct {
		Host *string
		Port *int
	}{
		Host: stringPtr("localhost"),
		Port: intPtr(8080),
	}

	data := map[string]any{
		"host": "example.com",
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "example.com", *config.Host)
	assert.Equal(t, 8080, *config.Port) // preserved
}

func TestBindToNilPointerField(t *testing.T) {
	config := &struct {
		Host *string
		Port *int
	}{}

	data := map[string]any{
		"host": "example.com",
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "example.com", *config.Host)
	assert.Nil(t, config.Port) // remains nil
}

func TestBindToStructPointer(t *testing.T) {
	type Database struct {
		Host string
		Port int
	}

	config := &struct {
		DB *Database
	}{
		DB: &Database{
			Host: "localhost",
			Port: 5432,
		},
	}

	data := map[string]any{
		"d_b": map[string]any{
			"host": "db.example.com",
		},
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "db.example.com", config.DB.Host)
	assert.Equal(t, 5432, config.DB.Port) // preserved
}

func TestBindToNilStructPointer(t *testing.T) {
	type Database struct {
		Host string
		Port int
	}

	config := &struct {
		DB *Database
	}{}

	data := map[string]any{
		"d_b": map[string]any{
			"host": "db.example.com",
			"port": 3306,
		},
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "db.example.com", config.DB.Host)
	assert.Equal(t, 3306, config.DB.Port)
}

func TestBindToSlice(t *testing.T) {
	config := &struct {
		Tags []string
	}{
		Tags: []string{"default", "tag"},
	}

	data := map[string]any{
		"tags": []string{"new", "tags"},
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, []string{"new", "tags"}, config.Tags) // replaced entirely
}

func TestBindToEmptySlice(t *testing.T) {
	config := &struct {
		Tags []string
	}{
		Tags: []string{"default", "tag"},
	}

	data := map[string]any{}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, []string{"default", "tag"}, config.Tags) // preserved
}

func TestBindToRequiredField(t *testing.T) {
	config := &struct {
		Host string `df:",required"`
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}

	data := map[string]any{
		"port": 9090,
	}

	err := BindTo(config, data)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "required field missing")
}

func TestBindToWithStructTags(t *testing.T) {
	config := &struct {
		Host string `df:"server_host"`
		Port int    `df:"server_port"`
	}{
		Host: "localhost",
		Port: 8080,
	}

	data := map[string]any{
		"server_host": "example.com",
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "example.com", config.Host)
	assert.Equal(t, 8080, config.Port) // preserved
}

func TestBindToSkippedField(t *testing.T) {
	config := &struct {
		Host   string
		Secret string `df:"-"`
	}{
		Host:   "localhost",
		Secret: "default-secret",
	}

	data := map[string]any{
		"host":   "example.com",
		"secret": "new-secret",
	}

	err := BindTo(config, data)
	assert.Nil(t, err)
	assert.Equal(t, "example.com", config.Host)
	assert.Equal(t, "default-secret", config.Secret) // preserved (skipped)
}

// helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}