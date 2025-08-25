package df

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnbindBasic(t *testing.T) {
	s := &struct{ StringValue string }{StringValue: "oh, wow!"}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"string_value": "oh, wow!"}, m)
}

func TestUnbindRenaming(t *testing.T) {
	s := &struct {
		SomeInt int `df:"some_int_,required"`
	}{SomeInt: 46}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"some_int_": 46}, m)
}

func TestUnbindArrays(t *testing.T) {
	s := &struct {
		StringArray []string
		IntArray    []int
	}{StringArray: []string{"one", "two", "three"}, IntArray: []int{1, 2, 3}}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"string_array": []interface{}{"one", "two", "three"},
		"int_array":    []interface{}{1, 2, 3},
	}, m)
}

func TestUnbindNestedPtrAndValue(t *testing.T) {
	s1 := &struct {
		Id     string
		Nested *nestedType
	}{Id: "NestedPtr", Nested: &nestedType{Name: "Different"}}
	m1, err := Unbind(s1)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "NestedPtr",
		"nested": map[string]any{
			"name":  "Different",
			"count": 0,
		},
	}, m1)

	s2 := &struct {
		Id     string
		Nested nestedType
	}{Id: "NestedVal", Nested: nestedType{Name: "V"}}
	m2, err := Unbind(s2)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "NestedVal",
		"nested": map[string]any{
			"name":  "V",
			"count": 0,
		},
	}, m2)
}

func TestUnbindStructTypeArray(t *testing.T) {
	s := &struct {
		Id      string
		Nesteds []*nestedType
	}{Id: "StructTypeArray", Nesteds: []*nestedType{{Name: "a"}, {Name: "b"}}}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "StructTypeArray",
		"nesteds": []interface{}{
			map[string]any{"name": "a", "count": 0},
			map[string]any{"name": "b", "count": 0},
		},
	}, m)
}

func TestUnbindAnonymousStruct(t *testing.T) {
	s := &struct {
		Id     string
		Nested struct{ Name string }
	}{Id: "AnonymousStruct"}
	s.Nested.Name = "oh, wow!"
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "AnonymousStruct",
		"nested": map[string]any{
			"name": "oh, wow!",
		},
	}, m)
}

func TestUnbindUint32(t *testing.T) {
	s := &struct {
		Id    string
		Count uint32
	}{Id: "uint32", Count: 33}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"id": "uint32", "count": uint32(33)}, m)
}

func TestUnbindTimeDuration(t *testing.T) {
	s := &struct{ Duration time.Duration }{Duration: 30 * time.Second}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"duration": "30s"}, m)
}

func TestUnbindSnakeCaseDefault(t *testing.T) {
	s := &struct{ OhWow int }{OhWow: 42}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"oh_wow": 42}, m)
}

func TestUnbindSkipTag(t *testing.T) {
	s := &struct {
		A int `df:"-"`
		B int
	}{A: 1, B: 2}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"b": 2}, m)
}

func TestUnbindOmitNilPtr(t *testing.T) {
	s := &struct {
		Name   string
		Nested *nestedType
	}{Name: "root", Nested: nil}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"name": "root"}, m)
}

func TestUnbindDynamicSlice(t *testing.T) {
	type root struct {
		Items []Dynamic
	}

	r := &root{
		Items: []Dynamic{
			&dynA{Name: "alpha"},
			&dynB{Count: 42},
		},
	}

	m, err := Unbind(r)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"items": []interface{}{
			map[string]any{"type": "a", "name": "alpha"},
			map[string]any{"type": "b", "count": 42},
		},
	}, m)
}

// customMarshalType is a custom type for testing Marshaler
type customMarshalType struct {
	Value string
}

// MarshalDf implements the Marshaler interface for *customMarshalType
func (c *customMarshalType) MarshalDf() (map[string]any, error) {
	return map[string]any{
		"value": "custom-" + c.Value,
	}, nil
}

func TestUnbindCustomMarshaler(t *testing.T) {
	// test with a pointer field
	t.Run("pointer field", func(t *testing.T) {
		source := &struct {
			Custom *customMarshalType
		}{
			Custom: &customMarshalType{Value: "hello"},
		}
		m, err := Unbind(source)
		assert.NoError(t, err)
		expected := map[string]any{
			"custom": map[string]any{
				"value": "custom-hello",
			},
		}
		assert.Equal(t, expected, m)
	})

	// test with a value field
	t.Run("value field", func(t *testing.T) {
		source := &struct {
			Custom customMarshalType
		}{
			Custom: customMarshalType{Value: "world"},
		}
		m, err := Unbind(source)
		assert.NoError(t, err)
		expected := map[string]any{
			"custom": map[string]any{
				"value": "custom-world",
			},
		}
		assert.Equal(t, expected, m)
	})
}

func TestUnbindCustomStringConstType(t *testing.T) {
	source := &struct {
		Name   string
		Status Status
	}{
		Name:   "test item",
		Status: StatusActive,
	}

	m, err := Unbind(source)
	assert.NoError(t, err)
	expected := map[string]any{
		"name":   "test item",
		"status": Status("active"), // Custom types are preserved in Unbind
	}
	assert.Equal(t, expected, m)
}
