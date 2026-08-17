package dd

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullableNestedJSONField(t *testing.T) {
	type message struct {
		Content *string `dd:",+nullable"`
	}
	type request struct {
		Messages []message
	}

	for _, test := range []struct {
		name string
		opts []*Options
	}{
		{name: "forgiving"},
		{name: "strict", opts: []*Options{Strict()}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var target request
			err := BindJSON(&target, []byte(`{"messages":[{"content":null}]}`), test.opts...)
			require.NoError(t, err)
			require.Len(t, target.Messages, 1)
			assert.Nil(t, target.Messages[0].Content)
		})
	}
}

func TestNullableYAMLField(t *testing.T) {
	type document struct {
		AllowedModels []string `dd:",+nullable"`
	}

	for _, test := range []struct {
		name string
		opts []*Options
	}{
		{name: "forgiving"},
		{name: "strict", opts: []*Options{Strict()}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var target document
			err := BindYAML(&target, []byte("allowed_models: null\n"), test.opts...)
			require.NoError(t, err)
			assert.Nil(t, target.AllowedModels)
		})
	}
}

func TestNullableFieldDoesNotAllocateEmbeddedPointer(t *testing.T) {
	type NullableEmbedded struct {
		Content *string `dd:",+nullable"`
	}
	type document struct {
		*NullableEmbedded
	}

	for _, test := range []struct {
		name string
		opts []*Options
	}{
		{name: "forgiving"},
		{name: "strict", opts: []*Options{Strict()}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var target document
			err := BindJSON(&target, []byte(`{"content":null}`), test.opts...)
			require.NoError(t, err)
			assert.Nil(t, target.NullableEmbedded)
		})
	}
}

func TestNullableEmbeddedFieldIsNotCapturedAsExtra(t *testing.T) {
	type NullableEmbedded struct {
		Content *string        `dd:",+nullable"`
		Extra   map[string]any `dd:",+extra"`
	}
	type document struct {
		*NullableEmbedded
	}

	var target document
	err := BindJSON(&target, []byte(`{"content":null,"extension":true}`), Strict())
	require.NoError(t, err)
	require.NotNil(t, target.NullableEmbedded)
	assert.Nil(t, target.Content)
	assert.Equal(t, map[string]any{"extension": true}, target.Extra)
}

func TestNullableBindsValuesAndTreatsNullAsAbsent(t *testing.T) {
	type document struct {
		Name      string         `dd:",+nullable"`
		Tags      []string       `dd:",+nullable"`
		ExpiresAt time.Time      `dd:",+nullable"`
		Metadata  map[string]any `dd:",+nullable"`
	}

	var nulls document
	err := BindJSON(&nulls, []byte(`{"name":null,"tags":null,"expires_at":null,"metadata":null}`))
	require.NoError(t, err)
	assert.Empty(t, nulls.Name)
	assert.Nil(t, nulls.Tags)
	assert.True(t, nulls.ExpiresAt.IsZero())
	assert.Nil(t, nulls.Metadata)

	var values document
	err = BindJSON(&values, []byte(`{"name":"example","tags":["one"],"expires_at":"2026-08-17T12:00:00Z","metadata":{"enabled":true}}`))
	require.NoError(t, err)
	assert.Equal(t, "example", values.Name)
	assert.Equal(t, []string{"one"}, values.Tags)
	assert.Equal(t, time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC), values.ExpiresAt)
	assert.Equal(t, map[string]any{"enabled": true}, values.Metadata)
}

func TestNullableRequiredNullIsMissing(t *testing.T) {
	type document struct {
		Name string `dd:",+required,+nullable"`
	}

	for _, data := range [][]byte{
		[]byte(`{}`),
		[]byte(`{"name":null}`),
	} {
		var target document
		err := BindJSON(&target, data)
		var requiredField *RequiredFieldError
		assert.ErrorAs(t, err, &requiredField)
	}
}

func TestNullableMergePreservesExistingValue(t *testing.T) {
	type config struct {
		AllowedModels []string `dd:",+nullable"`
		Name          *string  `dd:",+nullable"`
	}

	name := "existing"
	target := config{
		AllowedModels: []string{"model-a"},
		Name:          &name,
	}
	err := Merge(&target, map[string]any{
		"allowed_models": nil,
		"name":           nil,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"model-a"}, target.AllowedModels)
	require.NotNil(t, target.Name)
	assert.Equal(t, "existing", *target.Name)
}

func TestNullableOpaqueField(t *testing.T) {
	type document struct {
		Config map[string]any `dd:",+opaque,+nullable"`
	}

	target := document{Config: map[string]any{"existing": true}}
	err := Merge(&target, map[string]any{"config": nil})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"existing": true}, target.Config)
}

func TestNullWithoutNullableStillFails(t *testing.T) {
	type document struct {
		Content *string
	}

	var target document
	err := BindJSON(&target, []byte(`{"content":null}`))
	var bindingError *BindingError
	assert.True(t, errors.As(err, &bindingError))
}
