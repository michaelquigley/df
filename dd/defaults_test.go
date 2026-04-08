package dd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mergeDefaultSubConfig struct {
	Name            string
	Port            int
	defaultsApplied int
}

func (c *mergeDefaultSubConfig) ApplyDefaults() {
	c.defaultsApplied++
	if c.Name == "" {
		c.Name = "default name"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
}

func TestMergeAppliesDefaultsToNilStructPointer(t *testing.T) {
	config := &struct {
		Sub *mergeDefaultSubConfig
	}{}

	data := map[string]any{
		"sub": map[string]any{
			"port": 9090,
		},
	}

	err := Merge(config, data)
	assert.NoError(t, err)
	if assert.NotNil(t, config.Sub) {
		assert.Equal(t, "default name", config.Sub.Name)
		assert.Equal(t, 9090, config.Sub.Port)
		assert.Equal(t, 1, config.Sub.defaultsApplied)
	}
}

func TestMergeDoesNotApplyDefaultsToExistingStructPointer(t *testing.T) {
	config := &struct {
		Sub *mergeDefaultSubConfig
	}{
		Sub: &mergeDefaultSubConfig{
			Name: "existing",
			Port: 7070,
		},
	}

	data := map[string]any{
		"sub": map[string]any{
			"port": 9090,
		},
	}

	err := Merge(config, data)
	assert.NoError(t, err)
	if assert.NotNil(t, config.Sub) {
		assert.Equal(t, "existing", config.Sub.Name)
		assert.Equal(t, 9090, config.Sub.Port)
		assert.Equal(t, 0, config.Sub.defaultsApplied)
	}
}

func TestBindDoesNotApplyDefaultsToFreshStructPointer(t *testing.T) {
	config := &struct {
		Sub *mergeDefaultSubConfig
	}{}

	data := map[string]any{
		"sub": map[string]any{
			"port": 9090,
		},
	}

	err := Bind(config, data)
	assert.NoError(t, err)
	if assert.NotNil(t, config.Sub) {
		assert.Equal(t, "", config.Sub.Name)
		assert.Equal(t, 9090, config.Sub.Port)
		assert.Equal(t, 0, config.Sub.defaultsApplied)
	}
}

func TestMergeAppliesDefaultsToStructPointerSliceElements(t *testing.T) {
	config := &struct {
		Subs []*mergeDefaultSubConfig
	}{}

	data := map[string]any{
		"subs": []any{
			map[string]any{"port": 9090},
			map[string]any{"name": "explicit"},
		},
	}

	err := Merge(config, data)
	assert.NoError(t, err)
	if assert.Len(t, config.Subs, 2) {
		assert.Equal(t, "default name", config.Subs[0].Name)
		assert.Equal(t, 9090, config.Subs[0].Port)
		assert.Equal(t, 1, config.Subs[0].defaultsApplied)

		assert.Equal(t, "explicit", config.Subs[1].Name)
		assert.Equal(t, 8080, config.Subs[1].Port)
		assert.Equal(t, 1, config.Subs[1].defaultsApplied)
	}
}

func TestMergeAppliesDefaultsToStructPointerMapValues(t *testing.T) {
	config := &struct {
		Subs map[string]*mergeDefaultSubConfig
	}{}

	data := map[string]any{
		"subs": map[string]any{
			"first":  map[string]any{"port": 9090},
			"second": map[string]any{"name": "explicit"},
		},
	}

	err := Merge(config, data)
	assert.NoError(t, err)
	if assert.Len(t, config.Subs, 2) {
		assert.Equal(t, "default name", config.Subs["first"].Name)
		assert.Equal(t, 9090, config.Subs["first"].Port)
		assert.Equal(t, 1, config.Subs["first"].defaultsApplied)

		assert.Equal(t, "explicit", config.Subs["second"].Name)
		assert.Equal(t, 8080, config.Subs["second"].Port)
		assert.Equal(t, 1, config.Subs["second"].defaultsApplied)
	}
}

func TestMergeAppliesDefaultsToStructValueSliceElements(t *testing.T) {
	config := &struct {
		Subs []mergeDefaultSubConfig
	}{}

	data := map[string]any{
		"subs": []any{
			map[string]any{"port": 9090},
		},
	}

	err := Merge(config, data)
	assert.NoError(t, err)
	if assert.Len(t, config.Subs, 1) {
		assert.Equal(t, "default name", config.Subs[0].Name)
		assert.Equal(t, 9090, config.Subs[0].Port)
		assert.Equal(t, 1, config.Subs[0].defaultsApplied)
	}
}
