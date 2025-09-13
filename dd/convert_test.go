package dd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUint32(t *testing.T) {
	root := &struct {
		Id    string
		Count uint32
	}{}

	data := map[string]any{
		"id":    "uint32",
		"count": 33,
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "uint32", root.Id)
	assert.Equal(t, uint32(33), root.Count)
}

func TestTimeDuration(t *testing.T) {
	root := &struct {
		Duration time.Duration
	}{}

	data := map[string]any{
		"duration": "30s",
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(30)*time.Second, root.Duration)
}

func TestFloatWithIntData(t *testing.T) {
	basic := &struct {
		FloatValue float64
	}{}

	data := map[string]any{
		"float_value": 56,
	}

	err := Bind(basic, data)
	assert.Nil(t, err)
	assert.Equal(t, float64(56.0), basic.FloatValue)
}

func TestSliceTypeCompatibility(t *testing.T) {
	// ensure we can accept []interface{} as input for a []string field
	withArray := &struct {
		Items []string
	}{}

	arr := []interface{}{"a", "b", "c"}
	data := map[string]any{
		"items": arr,
	}
	err := Bind(withArray, data)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, withArray.Items)
}

func TestCoercion(t *testing.T) {
	data := map[string]any{
		"port":     "8080", // string → int
		"timeout":  30.5,   // float → int
		"enabled":  "true", // string → bool
		"duration": "5m",   // string → time.Duration
	}

	type coercion struct {
		Port     int           `dd:"port"`
		Timeout  int           `dd:"timeout"`
		Enabled  bool          `dd:"enabled"`
		Duration time.Duration `dd:"duration"`
	}

	config, err := New[coercion](data)
	assert.NoError(t, err)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, 30, config.Timeout)
	assert.Equal(t, true, config.Enabled)
}
