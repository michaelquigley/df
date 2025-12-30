package dd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtraFieldBasicBind(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	data := map[string]any{
		"name":     "test",
		"unknown1": "value1",
		"unknown2": 42,
	}

	var cfg Config
	err := Bind(&cfg, data)

	assert.Nil(t, err)
	assert.Equal(t, "test", cfg.Name)
	assert.Equal(t, map[string]any{
		"unknown1": "value1",
		"unknown2": 42,
	}, cfg.Extra)
}

func TestExtraFieldBasicUnbind(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	cfg := Config{
		Name: "test",
		Extra: map[string]any{
			"custom1": "value1",
			"custom2": 123,
		},
	}

	result, err := Unbind(cfg)

	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"name":    "test",
		"custom1": "value1",
		"custom2": 123,
	}, result)
}

func TestExtraFieldNestedStruct(t *testing.T) {
	type Inner struct {
		Value string         `dd:"value"`
		Extra map[string]any `dd:",+extra"`
	}
	type Outer struct {
		Name  string         `dd:"name"`
		Inner Inner          `dd:"inner"`
		Extra map[string]any `dd:",+extra"`
	}

	data := map[string]any{
		"name":        "outer",
		"outer_extra": "at_outer",
		"inner": map[string]any{
			"value":       "inner_value",
			"inner_extra": "at_inner",
		},
	}

	var o Outer
	err := Bind(&o, data)

	assert.Nil(t, err)
	assert.Equal(t, "outer", o.Name)
	assert.Equal(t, "inner_value", o.Inner.Value)
	assert.Equal(t, map[string]any{"outer_extra": "at_outer"}, o.Extra)
	assert.Equal(t, map[string]any{"inner_extra": "at_inner"}, o.Inner.Extra)
}

func TestExtraFieldWrongType(t *testing.T) {
	type Bad struct {
		Name  string `dd:"name"`
		Extra string `dd:",+extra"` // wrong type!
	}

	data := map[string]any{"name": "test", "unknown": "value"}

	var b Bad
	err := Bind(&b, data)

	assert.NotNil(t, err)
	assert.IsType(t, &TypeMismatchError{}, err)
}

func TestExtraFieldMultiple(t *testing.T) {
	type Bad struct {
		Name   string         `dd:"name"`
		Extra1 map[string]any `dd:",+extra"`
		Extra2 map[string]any `dd:",+extra"` // second +extra field
	}

	data := map[string]any{"name": "test", "unknown": "value"}

	var b Bad
	err := Bind(&b, data)

	assert.NotNil(t, err)
	assert.IsType(t, &MultipleExtraFieldsError{}, err)
}

func TestExtraFieldUnbindCollision(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	cfg := Config{
		Name: "test",
		Extra: map[string]any{
			"name": "collision!", // conflicts with Name field
		},
	}

	_, err := Unbind(cfg)

	assert.NotNil(t, err)
	assert.IsType(t, &ValidationError{}, err)
}

func TestExtraFieldEmbeddedStruct(t *testing.T) {
	type Base struct {
		ID string `dd:"id"`
	}
	type Extended struct {
		Base
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	data := map[string]any{
		"id":      "123",
		"name":    "extended",
		"unknown": "captured",
	}

	var e Extended
	err := Bind(&e, data)

	assert.Nil(t, err)
	assert.Equal(t, "123", e.ID)
	assert.Equal(t, "extended", e.Name)
	assert.Equal(t, map[string]any{"unknown": "captured"}, e.Extra)
}

func TestExtraFieldRoundTrip(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	original := map[string]any{
		"name":     "test",
		"custom_a": "value_a",
		"custom_b": float64(42), // JSON numbers become float64
	}

	var cfg Config
	err := Bind(&cfg, original)
	assert.Nil(t, err)

	result, err := Unbind(cfg)
	assert.Nil(t, err)
	assert.Equal(t, original, result)
}

func TestExtraFieldEmpty(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	data := map[string]any{"name": "test"} // no extra keys

	var cfg Config
	err := Bind(&cfg, data)

	assert.Nil(t, err)
	assert.Equal(t, "test", cfg.Name)
	assert.Nil(t, cfg.Extra) // should remain nil when no extras
}

func TestExtraFieldMerge(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	cfg := Config{
		Name: "original",
		Extra: map[string]any{
			"existing": "kept",
		},
	}

	data := map[string]any{
		"name":    "updated",
		"new_key": "added",
	}

	err := Merge(&cfg, data)

	assert.Nil(t, err)
	assert.Equal(t, "updated", cfg.Name)
	// merge should add new extras to existing map
	assert.Equal(t, map[string]any{
		"existing": "kept",
		"new_key":  "added",
	}, cfg.Extra)
}

func TestExtraFieldInSlice(t *testing.T) {
	type Item struct {
		ID    string         `dd:"id"`
		Extra map[string]any `dd:",+extra"`
	}
	type Container struct {
		Items []Item `dd:"items"`
	}

	data := map[string]any{
		"items": []any{
			map[string]any{"id": "1", "unknown_a": "a"},
			map[string]any{"id": "2", "unknown_b": "b"},
		},
	}

	var c Container
	err := Bind(&c, data)

	assert.Nil(t, err)
	assert.Len(t, c.Items, 2)
	assert.Equal(t, "1", c.Items[0].ID)
	assert.Equal(t, map[string]any{"unknown_a": "a"}, c.Items[0].Extra)
	assert.Equal(t, "2", c.Items[1].ID)
	assert.Equal(t, map[string]any{"unknown_b": "b"}, c.Items[1].Extra)
}

func TestExtraFieldNilUnbind(t *testing.T) {
	type Config struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}

	cfg := Config{
		Name:  "test",
		Extra: nil, // nil extra field
	}

	result, err := Unbind(cfg)

	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"name": "test"}, result)
}

func TestExtraFieldWithNamedTag(t *testing.T) {
	type Config struct {
		Name      string         `dd:"name"`
		ExtraData map[string]any `dd:"extra_data,+extra"` // should ignore the name part
	}

	data := map[string]any{
		"name":    "test",
		"unknown": "captured",
	}

	var cfg Config
	err := Bind(&cfg, data)

	assert.Nil(t, err)
	assert.Equal(t, "test", cfg.Name)
	assert.Equal(t, map[string]any{"unknown": "captured"}, cfg.ExtraData)
}

func TestExtraFieldPointerStruct(t *testing.T) {
	type Inner struct {
		Value string         `dd:"value"`
		Extra map[string]any `dd:",+extra"`
	}
	type Outer struct {
		Name  string `dd:"name"`
		Inner *Inner `dd:"inner"`
	}

	data := map[string]any{
		"name": "outer",
		"inner": map[string]any{
			"value":   "inner_value",
			"unknown": "captured",
		},
	}

	var o Outer
	err := Bind(&o, data)

	assert.Nil(t, err)
	assert.Equal(t, "outer", o.Name)
	assert.NotNil(t, o.Inner)
	assert.Equal(t, "inner_value", o.Inner.Value)
	assert.Equal(t, map[string]any{"unknown": "captured"}, o.Inner.Extra)
}

func TestExtraFieldMapValues(t *testing.T) {
	type Config struct {
		Value string         `dd:"value"`
		Extra map[string]any `dd:",+extra"`
	}
	type Container struct {
		Configs map[string]Config `dd:"configs"`
	}

	data := map[string]any{
		"configs": map[string]any{
			"a": map[string]any{"value": "va", "extra_a": "ea"},
			"b": map[string]any{"value": "vb", "extra_b": "eb"},
		},
	}

	var c Container
	err := Bind(&c, data)

	assert.Nil(t, err)
	assert.Len(t, c.Configs, 2)
	assert.Equal(t, "va", c.Configs["a"].Value)
	assert.Equal(t, map[string]any{"extra_a": "ea"}, c.Configs["a"].Extra)
	assert.Equal(t, "vb", c.Configs["b"].Value)
	assert.Equal(t, map[string]any{"extra_b": "eb"}, c.Configs["b"].Extra)
}
