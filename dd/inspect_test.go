package dd

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Name     string `df:"app_name"`
	Port     int
	Secret   string `df:"api_key,+secret"`
	Timeout  time.Duration
	Enabled  bool
	Database *testDB
	Services []testService
}

type testDB struct {
	Host     string
	Username string
	Password string `df:"+secret"`
	Port     int
}

type testService struct {
	Name string
	URL  string `df:"url"`
}

func TestInspect_BasicStruct(t *testing.T) {
	config := &testConfig{
		Name:    "myapp",
		Port:    8080,
		Secret:  "supersecret",
		Timeout: 30 * time.Second,
		Enabled: true,
	}

	result, err := Inspect(config)
	assert.NoError(t, err)

	// secret should be hidden by default
	assert.NotContains(t, result, "supersecret")
	assert.Contains(t, result, "testConfig {")
	assert.Contains(t, result, `"myapp"`)
	assert.Contains(t, result, "8080")
	assert.Contains(t, result, "30s")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "<nil>")
}

func TestInspect_WithSecrets(t *testing.T) {
	config := &testConfig{
		Secret: "supersecret",
	}

	// with ShowSecrets: false (default) - should show <set>
	result, err := Inspect(config)
	assert.NoError(t, err)
	assert.NotContains(t, result, "supersecret")
	assert.Contains(t, result, "api_key (secret)")
	assert.Contains(t, result, "<set>")

	// with ShowSecrets: true - should show actual value
	result, err = Inspect(config, &InspectOptions{ShowSecrets: true})
	assert.NoError(t, err)
	assert.Contains(t, result, "supersecret")
	assert.Contains(t, result, "api_key (secret)")
}

func TestInspect_NestedStruct(t *testing.T) {
	config := &testConfig{
		Name: "myapp",
		Database: &testDB{
			Host:     "localhost",
			Username: "user",
			Password: "secret123",
			Port:     5432,
		},
	}

	result, err := Inspect(config)
	assert.NoError(t, err)

	assert.Contains(t, result, "testConfig {")
	assert.Contains(t, result, "testDB {")
	assert.Contains(t, result, `"localhost"`)
	assert.Contains(t, result, `"user"`)
	assert.Contains(t, result, "5432")
	// password should show as <set>
	assert.NotContains(t, result, "secret123")
	assert.Contains(t, result, "<set>")
}

func TestInspect_NestedStructWithSecrets(t *testing.T) {
	config := &testConfig{
		Database: &testDB{
			Password: "secret123",
		},
	}

	result, err := Inspect(config, &InspectOptions{ShowSecrets: true})
	assert.NoError(t, err)

	assert.Contains(t, result, "password (secret)")
	assert.Contains(t, result, `"secret123"`)
}

func TestInspect_EmptySecrets(t *testing.T) {
	config := &testConfig{
		Secret: "", // empty secret
		Database: &testDB{
			Password: "", // empty password
		},
	}

	result, err := Inspect(config)
	assert.NoError(t, err)

	// should show <unset> for empty secret fields
	assert.Contains(t, result, "api_key (secret)")
	assert.Contains(t, result, "<unset>")
	assert.Contains(t, result, "password (secret)")
	assert.Contains(t, result, "<unset>")
}

func TestInspect_Slice(t *testing.T) {
	config := &testConfig{
		Services: []testService{
			{Name: "auth", URL: "http://auth:8080"},
			{Name: "api", URL: "http://api:8081"},
		},
	}

	result, err := Inspect(config)
	assert.NoError(t, err)

	assert.Contains(t, result, "[")
	assert.Contains(t, result, "[0]: testService {")
	assert.Contains(t, result, `"auth"`)
	assert.Contains(t, result, `"http://auth:8080"`)
	assert.Contains(t, result, "[1]: testService {")
	assert.Contains(t, result, `"api"`)
	assert.Contains(t, result, `"http://api:8081"`)
}

func TestInspect_EmptySlice(t *testing.T) {
	config := &testConfig{
		Services: []testService{},
	}

	result, err := Inspect(config)
	assert.NoError(t, err)

	assert.Contains(t, result, "[]")
}

func TestInspect_NilPointer(t *testing.T) {
	var config *testConfig

	result, err := Inspect(config)
	assert.NoError(t, err)
	assert.Equal(t, "<nil>", result)
}

func TestInspect_NilStructField(t *testing.T) {
	config := &testConfig{
		Database: nil,
	}

	result, err := Inspect(config)
	assert.NoError(t, err)
	assert.Contains(t, result, "<nil>")
}

func TestInspect_CustomOptions(t *testing.T) {
	config := &testConfig{
		Name: "test",
		Database: &testDB{
			Host: "localhost",
		},
	}

	opts := &InspectOptions{
		Indent:      "    ",
		ShowSecrets: true,
	}

	result, err := Inspect(config, opts)
	assert.NoError(t, err)

	// check custom indentation
	lines := strings.Split(result, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "app_name") && strings.Contains(line, ":") {
			assert.True(t, strings.HasPrefix(line, "    "))
			found = true
			break
		}
	}
	assert.True(t, found, "should find line with custom indentation")
}

func TestInspect_MaxDepth(t *testing.T) {
	config := &testConfig{
		Database: &testDB{
			Host: "localhost",
		},
	}

	opts := &InspectOptions{
		MaxDepth: 1,
	}

	result, err := Inspect(config, opts)
	assert.NoError(t, err)

	// should contain the struct but fields inside should be truncated due to depth limit
	assert.Contains(t, result, "testDB {")
	assert.Contains(t, result, "<max depth reached>")
}

type skippedFieldStruct struct {
	Public   string `df:"public_field"`
	Hidden   string `df:"-"`
	Internal string `df:"internal"`
}

func TestInspect_SkippedFields(t *testing.T) {
	s := &skippedFieldStruct{
		Public:   "visible",
		Hidden:   "should not appear",
		Internal: "also visible",
	}

	result, err := Inspect(s)
	assert.NoError(t, err)

	assert.Contains(t, result, `"visible"`)
	assert.Contains(t, result, `"also visible"`)
	assert.NotContains(t, result, "should not appear")
	assert.NotContains(t, result, "Hidden")
	assert.NotContains(t, result, "hidden")
}

type primitiveStruct struct {
	Str     string  `df:"string_field"`
	Bool    bool    `df:"bool_field"`
	Int     int     `df:"int_field"`
	Float   float64 `df:"float_field"`
	PtrStr  *string `df:"ptr_string_field"`
	PtrBool *bool   `df:"ptr_bool_field"`
}

func TestInspect_PrimitiveTypes(t *testing.T) {
	str := "hello"
	b := true
	s := &primitiveStruct{
		Str:     "test",
		Bool:    false,
		Int:     42,
		Float:   3.14159,
		PtrStr:  &str,
		PtrBool: &b,
	}

	result, err := Inspect(s)
	assert.NoError(t, err)

	assert.Contains(t, result, `"test"`)
	assert.Contains(t, result, "false")
	assert.Contains(t, result, "42")
	assert.Contains(t, result, "3.14159")
	assert.Contains(t, result, `"hello"`)
	assert.Contains(t, result, "true")
}

func TestInspect_NilPointers(t *testing.T) {
	s := &primitiveStruct{
		PtrStr:  nil,
		PtrBool: nil,
	}

	result, err := Inspect(s)
	assert.NoError(t, err)

	assert.Contains(t, result, "<nil>")
	assert.Contains(t, result, "<nil>")
}

func TestInspect_InvalidInput(t *testing.T) {
	// test with non-struct
	_, err := Inspect("not a struct")
	assert.Error(t, err)
	var typeMismatchErr *TypeMismatchError
	assert.True(t, errors.As(err, &typeMismatchErr))
	assert.Contains(t, err.Error(), "expected struct or pointer to struct, got string")

	// test with non-struct pointer
	s := "string"
	_, err = Inspect(&s)
	assert.Error(t, err)
	var typeMismatchErr2 *TypeMismatchError
	assert.True(t, errors.As(err, &typeMismatchErr2))
	assert.Contains(t, err.Error(), "expected struct or pointer to struct, got *string")
}

type emptyStruct struct{}

func TestInspect_EmptyStruct(t *testing.T) {
	s := &emptyStruct{}

	result, err := Inspect(s)
	assert.NoError(t, err)

	assert.Contains(t, result, "emptyStruct {")
	assert.Contains(t, result, "<no fields>")
}

type durationStruct struct {
	Timeout time.Duration `df:"timeout"`
}

func TestInspect_Duration(t *testing.T) {
	s := &durationStruct{
		Timeout: 5 * time.Minute,
	}

	result, err := Inspect(s)
	assert.NoError(t, err)

	assert.Contains(t, result, "5m0s")
}

func TestInspect_ZeroDuration(t *testing.T) {
	s := &durationStruct{
		Timeout: 0,
	}

	result, err := Inspect(s)
	assert.NoError(t, err)

	assert.Contains(t, result, "0s")
}
