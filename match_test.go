package df

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchConstraint(t *testing.T) {
	t.Run("successful string match", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,match=\"1.0.0\""`
			Name    string `df:"name"`
		}

		data := map[string]any{
			"version": "1.0.0",
			"name":    "test",
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "1.0.0", config.Version)
		assert.Equal(t, "test", config.Name)
	})

	t.Run("failed string match", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,match=\"1.0.0\""`
		}

		data := map[string]any{
			"version": "2.0.0",
		}

		var config Config
		err := Bind(&config, data)
		assert.Error(t, err)

		var valueErr *ValueMismatchError
		assert.ErrorAs(t, err, &valueErr)
		assert.Equal(t, "Config.Version", valueErr.Path+"."+valueErr.Field)
		assert.Equal(t, "1.0.0", valueErr.Expected)
		assert.Equal(t, "2.0.0", valueErr.Actual)
	})

	t.Run("successful numeric match", func(t *testing.T) {
		type Config struct {
			Port int `df:"port,match=\"8080\""`
		}

		data := map[string]any{
			"port": 8080,
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, 8080, config.Port)
	})

	t.Run("failed numeric match", func(t *testing.T) {
		type Config struct {
			Port int `df:"port,match=\"8080\""`
		}

		data := map[string]any{
			"port": 9000,
		}

		var config Config
		err := Bind(&config, data)
		assert.Error(t, err)

		var valueErr *ValueMismatchError
		assert.ErrorAs(t, err, &valueErr)
		assert.Equal(t, "Config.Port", valueErr.Path+"."+valueErr.Field)
		assert.Equal(t, "8080", valueErr.Expected)
		assert.Equal(t, "9000", valueErr.Actual)
	})

	t.Run("successful boolean match", func(t *testing.T) {
		type Config struct {
			Debug bool `df:"debug,match=\"true\""`
		}

		data := map[string]any{
			"debug": true,
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, true, config.Debug)
	})

	t.Run("failed boolean match", func(t *testing.T) {
		type Config struct {
			Debug bool `df:"debug,match=\"true\""`
		}

		data := map[string]any{
			"debug": false,
		}

		var config Config
		err := Bind(&config, data)
		assert.Error(t, err)

		var valueErr *ValueMismatchError
		assert.ErrorAs(t, err, &valueErr)
		assert.Equal(t, "Config.Debug", valueErr.Path+"."+valueErr.Field)
		assert.Equal(t, "true", valueErr.Expected)
		assert.Equal(t, "false", valueErr.Actual)
	})

	t.Run("match with custom field name", func(t *testing.T) {
		type Config struct {
			APIVersion string `df:"api_version,match=\"v1\""`
		}

		data := map[string]any{
			"api_version": "v1",
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "v1", config.APIVersion)
	})

	t.Run("match with required flag", func(t *testing.T) {
		type Config struct {
			Mode string `df:"mode,required,match=\"production\""`
		}

		data := map[string]any{
			"mode": "production",
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "production", config.Mode)
	})

	t.Run("failed match with required flag", func(t *testing.T) {
		type Config struct {
			Mode string `df:"mode,required,match=\"production\""`
		}

		data := map[string]any{
			"mode": "development",
		}

		var config Config
		err := Bind(&config, data)
		assert.Error(t, err)

		var valueErr *ValueMismatchError
		assert.ErrorAs(t, err, &valueErr)
		assert.Equal(t, "production", valueErr.Expected)
		assert.Equal(t, "development", valueErr.Actual)
	})

	t.Run("match with secret flag", func(t *testing.T) {
		type Config struct {
			KeyType string `df:"key_type,secret,match=\"rsa\""`
		}

		data := map[string]any{
			"key_type": "rsa",
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "rsa", config.KeyType)
	})

	t.Run("missing field with match constraint - no error", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,match=\"1.0.0\""`
		}

		data := map[string]any{}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)              // match is only checked if field is present
		assert.Equal(t, "", config.Version) // zero value
	})

	t.Run("missing required field with match constraint", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,required,match=\"1.0.0\""`
		}

		data := map[string]any{}

		var config Config
		err := Bind(&config, data)
		assert.Error(t, err)

		var reqErr *RequiredFieldError
		assert.ErrorAs(t, err, &reqErr) // should get required error first
		assert.Equal(t, "Config", reqErr.Path)
		assert.Equal(t, "Version", reqErr.Field)
	})

	t.Run("nested struct with match constraint", func(t *testing.T) {
		type Database struct {
			Driver string `df:"driver,match=\"postgresql\""`
		}

		type Config struct {
			DB Database `df:"database"`
		}

		data := map[string]any{
			"database": map[string]any{
				"driver": "postgresql",
			},
		}

		var config Config
		err := Bind(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "postgresql", config.DB.Driver)
	})

	t.Run("nested struct with failed match constraint", func(t *testing.T) {
		type Database struct {
			Driver string `df:"driver,match=\"postgresql\""`
		}

		type Config struct {
			DB Database `df:"database"`
		}

		data := map[string]any{
			"database": map[string]any{
				"driver": "mysql",
			},
		}

		var config Config
		err := Bind(&config, data)
		assert.Error(t, err)

		var bindErr *BindingError
		assert.ErrorAs(t, err, &bindErr)
		assert.Equal(t, "Config", bindErr.Path)
		assert.Equal(t, "DB", bindErr.Field)

		var valueErr *ValueMismatchError
		assert.ErrorAs(t, bindErr.Cause, &valueErr)
		assert.Equal(t, "postgresql", valueErr.Expected)
		assert.Equal(t, "mysql", valueErr.Actual)
	})
}

func TestMatchConstraintParsing(t *testing.T) {
	t.Run("basic match parsing", func(t *testing.T) {
		type TestStruct struct {
			Field string `df:"field,match=\"value\""`
		}

		field, _ := getStructField[TestStruct]("Field")
		tag := parseDfTag(field)

		assert.True(t, tag.HasMatch)
		assert.Equal(t, "value", tag.MatchValue)
		assert.Equal(t, "field", tag.Name)
	})

	t.Run("match with other flags", func(t *testing.T) {
		type TestStruct struct {
			Field string `df:"custom_name,required,secret,match=\"test\""`
		}

		field, _ := getStructField[TestStruct]("Field")
		tag := parseDfTag(field)

		assert.True(t, tag.HasMatch)
		assert.Equal(t, "test", tag.MatchValue)
		assert.Equal(t, "custom_name", tag.Name)
		assert.True(t, tag.Required)
		assert.True(t, tag.Secret)
	})

	t.Run("match as first token", func(t *testing.T) {
		type TestStruct struct {
			Field string `df:"match=\"first\",required"`
		}

		field, _ := getStructField[TestStruct]("Field")
		tag := parseDfTag(field)

		assert.True(t, tag.HasMatch)
		assert.Equal(t, "first", tag.MatchValue)
		assert.Equal(t, "", tag.Name) // no name specified
		assert.True(t, tag.Required)
	})

	t.Run("empty match value", func(t *testing.T) {
		type TestStruct struct {
			Field string `df:"match=\"\""`
		}

		field, _ := getStructField[TestStruct]("Field")
		tag := parseDfTag(field)

		assert.True(t, tag.HasMatch)
		assert.Equal(t, "", tag.MatchValue)
	})

	t.Run("malformed match - no quotes", func(t *testing.T) {
		type TestStruct struct {
			Field string `df:"match=value"`
		}

		field, _ := getStructField[TestStruct]("Field")
		tag := parseDfTag(field)

		assert.False(t, tag.HasMatch)
		assert.Equal(t, "", tag.MatchValue)
	})

	t.Run("malformed match - incomplete quotes", func(t *testing.T) {
		type TestStruct struct {
			Field string `df:"match=\"value"`
		}

		field, _ := getStructField[TestStruct]("Field")
		tag := parseDfTag(field)

		assert.False(t, tag.HasMatch)
		assert.Equal(t, "", tag.MatchValue)
	})
}

func TestMatchConstraintWithMerge(t *testing.T) {
	t.Run("merge with successful match", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,match=\"1.0.0\""`
			Name    string `df:"name"`
		}

		// Start with existing data
		config := Config{
			Version: "1.0.0", // This should stay the same
			Name:    "original",
		}

		// Merge data that matches the constraint
		data := map[string]any{
			"version": "1.0.0", // Matches constraint
			"name":    "updated",
		}

		err := Merge(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "1.0.0", config.Version)
		assert.Equal(t, "updated", config.Name)
	})

	t.Run("merge with failed match", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,match=\"1.0.0\""`
			Name    string `df:"name"`
		}

		// Start with existing data
		config := Config{
			Version: "1.0.0",
			Name:    "original",
		}

		// Try to merge data that violates the constraint
		data := map[string]any{
			"version": "2.0.0", // Violates constraint
			"name":    "updated",
		}

		err := Merge(&config, data)
		assert.Error(t, err)

		var valueErr *ValueMismatchError
		assert.ErrorAs(t, err, &valueErr)
		assert.Equal(t, "1.0.0", valueErr.Expected)
		assert.Equal(t, "2.0.0", valueErr.Actual)

		// Original values should be preserved when merge fails
		assert.Equal(t, "1.0.0", config.Version)
		assert.Equal(t, "original", config.Name)
	})

	t.Run("merge preserves existing values when constraint field missing", func(t *testing.T) {
		type Config struct {
			Version string `df:"version,match=\"1.0.0\""`
			Name    string `df:"name"`
		}

		// Start with existing data
		config := Config{
			Version: "1.0.0",
			Name:    "original",
		}

		// Merge data that doesn't include the constrained field
		data := map[string]any{
			"name": "updated",
			// version is missing - should preserve existing value
		}

		err := Merge(&config, data)
		assert.NoError(t, err)
		assert.Equal(t, "1.0.0", config.Version) // Preserved
		assert.Equal(t, "updated", config.Name)  // Updated
	})
}

// helper function to get struct field by name
func getStructField[T any](fieldName string) (reflect.StructField, bool) {
	var zero T
	t := reflect.TypeOf(zero)
	return t.FieldByName(fieldName)
}
