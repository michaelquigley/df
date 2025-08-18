package df

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AppConfig struct {
	Name     string `df:"app_name"`
	Port     int    `df:"port"`
	Debug    bool   `df:"debug"`
	Database struct {
		Host     string `df:"host"`
		Port     int    `df:"port"`
		Username string `df:"username"`
		Password string `df:"password,secret"`
	} `df:"database"`
}

func TestNewContainer(t *testing.T) {
	defaults := AppConfig{
		Name:  "myapp",
		Port:  8080,
		Debug: false,
		Database: struct {
			Host     string `df:"host"`
			Port     int    `df:"port"`
			Username string `df:"username"`
			Password string `df:"password,secret"`
		}{
			Host:     "localhost",
			Port:     5432,
			Username: "user",
			Password: "pass",
		},
	}

	container := NewContainer(defaults)
	require.NotNil(t, container)

	config := container.Config()
	require.NotNil(t, config)

	assert.Equal(t, "myapp", config.Name)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, false, config.Debug)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
}

func TestContainerMergeFromJSON(t *testing.T) {
	defaults := AppConfig{
		Name:  "myapp",
		Port:  8080,
		Debug: false,
		Database: struct {
			Host     string `df:"host"`
			Port     int    `df:"port"`
			Username string `df:"username"`
			Password string `df:"password,secret"`
		}{
			Host:     "localhost",
			Port:     5432,
			Username: "user",
			Password: "pass",
		},
	}

	container := NewContainer(defaults)

	// create temporary JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "config.json")
	jsonData := `{
		"app_name": "production-app",
		"port": 9090,
		"debug": true,
		"database": {
			"host": "db.example.com",
			"username": "produser"
		}
	}`
	err := os.WriteFile(jsonFile, []byte(jsonData), 0644)
	require.NoError(t, err)

	// merge from JSON
	err = container.MergeFromJSON(jsonFile)
	require.NoError(t, err)

	config := container.Config()
	assert.Equal(t, "production-app", config.Name)
	assert.Equal(t, 9090, config.Port)
	assert.Equal(t, true, config.Debug)
	assert.Equal(t, "db.example.com", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port) // unchanged from defaults
	assert.Equal(t, "produser", config.Database.Username)
	assert.Equal(t, "pass", config.Database.Password) // unchanged from defaults
}

func TestContainerMergeFromYAML(t *testing.T) {
	defaults := AppConfig{
		Name:  "myapp",
		Port:  8080,
		Debug: false,
		Database: struct {
			Host     string `df:"host"`
			Port     int    `df:"port"`
			Username string `df:"username"`
			Password string `df:"password,secret"`
		}{
			Host:     "localhost",
			Port:     5432,
			Username: "user",
			Password: "pass",
		},
	}

	container := NewContainer(defaults)

	// create temporary YAML file
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")
	yamlData := `
app_name: staging-app
port: 7070
database:
  host: staging.db.com
  port: 3306
  password: staging-secret
`
	err := os.WriteFile(yamlFile, []byte(yamlData), 0644)
	require.NoError(t, err)

	// merge from YAML
	err = container.MergeFromYAML(yamlFile)
	require.NoError(t, err)

	config := container.Config()
	assert.Equal(t, "staging-app", config.Name)
	assert.Equal(t, 7070, config.Port)
	assert.Equal(t, false, config.Debug) // unchanged from defaults
	assert.Equal(t, "staging.db.com", config.Database.Host)
	assert.Equal(t, 3306, config.Database.Port)
	assert.Equal(t, "user", config.Database.Username) // unchanged from defaults
	assert.Equal(t, "staging-secret", config.Database.Password)
}

func TestContainerLoadConfigFiles(t *testing.T) {
	defaults := AppConfig{
		Name:  "myapp",
		Port:  8080,
		Debug: false,
		Database: struct {
			Host     string `df:"host"`
			Port     int    `df:"port"`
			Username string `df:"username"`
			Password string `df:"password,secret"`
		}{
			Host:     "localhost",
			Port:     5432,
			Username: "user",
			Password: "pass",
		},
	}

	container := NewContainer(defaults)

	tmpDir := t.TempDir()

	// create base config file
	baseConfigFile := filepath.Join(tmpDir, "base.yaml")
	baseConfigData := `
app_name: base-app
port: 6000
database:
  host: base.db.com
`
	err := os.WriteFile(baseConfigFile, []byte(baseConfigData), 0644)
	require.NoError(t, err)

	// create override config file
	overrideConfigFile := filepath.Join(tmpDir, "override.json")
	overrideConfigData := `{
		"port": 7000,
		"debug": true,
		"database": {
			"username": "override-user"
		}
	}`
	err = os.WriteFile(overrideConfigFile, []byte(overrideConfigData), 0644)
	require.NoError(t, err)

	// load both files in order
	err = container.LoadConfigFiles([]string{baseConfigFile, overrideConfigFile})
	require.NoError(t, err)

	config := container.Config()
	assert.Equal(t, "base-app", config.Name)        // from base
	assert.Equal(t, 7000, config.Port)             // from override
	assert.Equal(t, true, config.Debug)            // from override
	assert.Equal(t, "base.db.com", config.Database.Host)    // from base
	assert.Equal(t, 5432, config.Database.Port)    // from defaults
	assert.Equal(t, "override-user", config.Database.Username) // from override
	assert.Equal(t, "pass", config.Database.Password) // from defaults
}

func TestContainerLoadConfigFiles_MissingFileSkipped(t *testing.T) {
	defaults := AppConfig{Name: "myapp", Port: 8080}
	container := NewContainer(defaults)

	tmpDir := t.TempDir()

	// create one valid file
	validFile := filepath.Join(tmpDir, "valid.json")
	validData := `{"app_name": "updated-app"}`
	err := os.WriteFile(validFile, []byte(validData), 0644)
	require.NoError(t, err)

	// include a missing file in the list
	missingFile := filepath.Join(tmpDir, "missing.json")
	
	// should skip missing file and process valid file
	err = container.LoadConfigFiles([]string{missingFile, validFile})
	require.NoError(t, err)

	config := container.Config()
	assert.Equal(t, "updated-app", config.Name) // from valid file
	assert.Equal(t, 8080, config.Port)         // from defaults
}

func TestContainerMergeFromJSON_FileNotFound(t *testing.T) {
	defaults := AppConfig{Name: "myapp"}
	container := NewContainer(defaults)

	err := container.MergeFromJSON("/nonexistent/file.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to merge config from JSON file")
	assert.Contains(t, err.Error(), "/nonexistent/file.json")
}

func TestContainerMergeFromYAML_FileNotFound(t *testing.T) {
	defaults := AppConfig{Name: "myapp"}
	container := NewContainer(defaults)

	err := container.MergeFromYAML("/nonexistent/file.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to merge config from YAML file")
	assert.Contains(t, err.Error(), "/nonexistent/file.yaml")
}

func TestContainerLoadConfigFiles_UnsupportedFileType(t *testing.T) {
	defaults := AppConfig{Name: "myapp"}
	container := NewContainer(defaults)

	tmpDir := t.TempDir()
	unsupportedFile := filepath.Join(tmpDir, "config.txt")
	err := os.WriteFile(unsupportedFile, []byte("some content"), 0644)
	require.NoError(t, err)

	err = container.LoadConfigFiles([]string{unsupportedFile})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config file type")
}

func TestIsJSONFile(t *testing.T) {
	assert.True(t, isJSONFile("config.json"))
	assert.True(t, isJSONFile("/path/to/config.json"))
	assert.False(t, isJSONFile("config.yaml"))
	assert.False(t, isJSONFile("config.yml"))
	assert.False(t, isJSONFile("config.txt"))
	assert.False(t, isJSONFile("json"))
}

func TestIsYAMLFile(t *testing.T) {
	assert.True(t, isYAMLFile("config.yaml"))
	assert.True(t, isYAMLFile("config.yml"))
	assert.True(t, isYAMLFile("/path/to/config.yaml"))
	assert.True(t, isYAMLFile("/path/to/config.yml"))
	assert.False(t, isYAMLFile("config.json"))
	assert.False(t, isYAMLFile("config.txt"))
	assert.False(t, isYAMLFile("yaml"))
	assert.False(t, isYAMLFile("yml"))
}