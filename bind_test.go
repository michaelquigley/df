package df

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicBind(t *testing.T) {
	basic := &struct {
		StringValue string
	}{}

	data := map[string]any{
		"string_value": "oh, wow!",
	}

	err := Bind(basic, data)
	assert.Nil(t, err)
	assert.Equal(t, "oh, wow!", basic.StringValue)
}

func TestRenaming(t *testing.T) {
	renamed := &struct {
		SomeInt int `df:"some_int_,required"`
	}{}

	data := map[string]any{
		"some_int_": 46,
	}

	err := Bind(renamed, data)
	assert.Nil(t, err)
	assert.Equal(t, 46, renamed.SomeInt)
}

func TestStringArray(t *testing.T) {
	withArray := &struct {
		StringArray []string
	}{}

	data := map[string]any{
		"string_array": []string{"one", "two", "three"},
	}

	err := Bind(withArray, data)
	assert.Nil(t, err)
	assert.EqualValues(t, []string{"one", "two", "three"}, withArray.StringArray)
}

func TestIntArray(t *testing.T) {
	withArray := &struct {
		IntArray []int
	}{}

	data := map[string]any{
		"int_array": []int{1, 2, 3, 4, 5, 6},
	}

	err := Bind(withArray, data)
	assert.Nil(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, withArray.IntArray)
}

func TestRequired(t *testing.T) {
	required := &struct {
		Required int `df:",required"`
	}{}

	data := make(map[string]any)

	err := Bind(required, data)
	assert.Error(t, err)
}

func TestNestedPtr(t *testing.T) {
	root := &struct {
		Id     string
		Nested *nestedType
	}{}

	data := map[string]any{
		"id": "TestNested",
		"nested": map[string]any{
			"name": "Different",
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "TestNested", root.Id)
	assert.NotNil(t, root.Nested)
	assert.Equal(t, "Different", root.Nested.Name)
	assert.Equal(t, 0, root.Nested.Count) // default zero value
}

func TestNestedValue(t *testing.T) {
	root := &struct {
		Id     string
		Nested nestedType
	}{}

	data := map[string]any{
		"id": "TestNested",
		"nested": map[string]any{
			"name": "Different",
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "TestNested", root.Id)
	assert.Equal(t, "Different", root.Nested.Name)
	assert.Equal(t, 0, root.Nested.Count)
}

func TestStructTypeArray(t *testing.T) {
	root := &struct {
		Id      string
		Nesteds []*nestedType
	}{}

	data := map[string]any{
		"id": "StructTypeArray",
		"nesteds": []map[string]any{
			{"name": "a"},
			{"name": "b"},
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "StructTypeArray", root.Id)
	assert.Equal(t, 2, len(root.Nesteds))
	assert.Equal(t, "a", root.Nesteds[0].Name)
	assert.Equal(t, "b", root.Nesteds[1].Name)
}

func TestAnonymousStruct(t *testing.T) {
	root := &struct {
		Id     string
		Nested struct {
			Name string
		}
	}{}

	data := map[string]any{
		"id": "AnonymousStruct",
		"nested": map[string]any{
			"name": "oh, wow!",
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "AnonymousStruct", root.Id)
	assert.Equal(t, "oh, wow!", root.Nested.Name)
}

func TestSnakeCaseDefault(t *testing.T) {
	// verify that untagged field names default to snake_case keys
	s := &struct{ OhWow int }{}
	data := map[string]any{"oh_wow": 42}
	err := Bind(s, data)
	assert.Nil(t, err)
	assert.Equal(t, 42, s.OhWow)
}

func TestBindDynamicSlice(t *testing.T) {
	type root struct {
		Items []Dynamic
	}

	data := map[string]any{
		"items": []map[string]any{
			{"type": "a", "name": "alpha"},
			{"type": "b", "count": 42},
		},
	}

	opts := &Options{
		DynamicBinders: map[string]func(map[string]any) (Dynamic, error){
			"a": func(m map[string]any) (Dynamic, error) {
				name, _ := m["name"].(string)
				return &dynA{Name: name}, nil
			},
			"b": func(m map[string]any) (Dynamic, error) {
				countAny := m["count"]
				count := 0
				switch v := countAny.(type) {
				case int:
					count = v
				case int64:
					count = int(v)
				case float64:
					count = int(v)
				}
				return &dynB{Count: count}, nil
			},
		},
	}

	r := &root{}
	err := Bind(r, data, opts)
	assert.NoError(t, err)
	if assert.Len(t, r.Items, 2) {
		a, ok := r.Items[0].(*dynA)
		assert.True(t, ok)
		assert.Equal(t, "alpha", a.Name)

		b, ok := r.Items[1].(*dynB)
		assert.True(t, ok)
		assert.Equal(t, 42, b.Count)
	}
}

func TestBindDynamicPerFieldBinders(t *testing.T) {
	type root struct {
		Action Dynamic
		Metric Dynamic
	}

	data := map[string]any{
		"action": map[string]any{"type": "a", "name": "alpha"},
		"metric": map[string]any{"type": "b", "count": 7},
	}

	opts := &Options{
		FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
			"root.Action": {
				"a": func(m map[string]any) (Dynamic, error) {
					name, _ := m["name"].(string)
					return &dynA{Name: name}, nil
				},
			},
			"root.Metric": {
				"b": func(m map[string]any) (Dynamic, error) {
					countAny := m["count"]
					count := 0
					switch v := countAny.(type) {
					case int:
						count = v
					case int64:
						count = int(v)
					case float64:
						count = int(v)
					}
					return &dynB{Count: count}, nil
				},
			},
		},
	}

	r := &root{}
	err := Bind(r, data, opts)
	assert.NoError(t, err)
	if assert.NotNil(t, r.Action) {
		a, ok := r.Action.(*dynA)
		assert.True(t, ok)
		assert.Equal(t, "alpha", a.Name)
	}
	if assert.NotNil(t, r.Metric) {
		b, ok := r.Metric.(*dynB)
		assert.True(t, ok)
		assert.Equal(t, 7, b.Count)
	}
}

func TestBindDynamicPerFieldSliceBinders(t *testing.T) {
	type root struct {
		Items  []Dynamic
		Others []Dynamic
	}

	data := map[string]any{
		"items":  []map[string]any{{"type": "a", "name": "A1"}, {"type": "a", "name": "A2"}},
		"others": []map[string]any{{"type": "b", "count": 1}, {"type": "b", "count": 2}},
	}

	opts := &Options{
		FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
			"root.Items": {
				"a": func(m map[string]any) (Dynamic, error) {
					name, _ := m["name"].(string)
					return &dynA{Name: name}, nil
				},
			},
			"root.Others": {
				"b": func(m map[string]any) (Dynamic, error) {
					countAny := m["count"]
					count := 0
					switch v := countAny.(type) {
					case int:
						count = v
					case int64:
						count = int(v)
					case float64:
						count = int(v)
					}
					return &dynB{Count: count}, nil
				},
			},
		},
	}

	r := &root{}
	err := Bind(r, data, opts)
	assert.NoError(t, err)
	if assert.Len(t, r.Items, 2) {
		a1, _ := r.Items[0].(*dynA)
		a2, _ := r.Items[1].(*dynA)
		assert.Equal(t, "A1", a1.Name)
		assert.Equal(t, "A2", a2.Name)
	}
	if assert.Len(t, r.Others, 2) {
		b1, _ := r.Others[0].(*dynB)
		b2, _ := r.Others[1].(*dynB)
		assert.Equal(t, 1, b1.Count)
		assert.Equal(t, 2, b2.Count)
	}
}

func TestUnbindThenBindWithPerFieldBinders(t *testing.T) {
	type root struct {
		Action Dynamic
		Others []Dynamic
	}

	// start with concrete dynamic values
	r1 := &root{
		Action: &dynA{Name: "alpha"},
		Others: []Dynamic{&dynB{Count: 1}, &dynB{Count: 2}},
	}

	// unbind to map
	m, err := Unbind(r1)
	assert.NoError(t, err)

	// bind back using per-field dynamic binders
	opts := &Options{
		FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
			"root.Action": {
				"a": func(m map[string]any) (Dynamic, error) {
					name, _ := m["name"].(string)
					return &dynA{Name: name}, nil
				},
			},
			"root.Others": {
				"b": func(m map[string]any) (Dynamic, error) {
					countAny := m["count"]
					count := 0
					switch v := countAny.(type) {
					case int:
						count = v
					case int64:
						count = int(v)
					case float64:
						count = int(v)
					}
					return &dynB{Count: count}, nil
				},
			},
		},
	}

	r2 := &root{}
	err = Bind(r2, m, opts)
	assert.NoError(t, err)

	if assert.NotNil(t, r2.Action) {
		a, ok := r2.Action.(*dynA)
		assert.True(t, ok)
		assert.Equal(t, "alpha", a.Name)
	}
	if assert.Len(t, r2.Others, 2) {
		b1, _ := r2.Others[0].(*dynB)
		b2, _ := r2.Others[1].(*dynB)
		assert.Equal(t, 1, b1.Count)
		assert.Equal(t, 2, b2.Count)
	}
}

func TestUnbindConcreteThenBindWithPerFieldBinders(t *testing.T) {
	type rootConcrete struct {
		DoA dynA
		Bs  []*dynB
	}
	type rootIface struct {
		DoA Dynamic
		Bs  []Dynamic
	}

	// start with concrete types implementing Dynamic
	r1 := &rootConcrete{
		DoA: dynA{Name: "alpha"},
		Bs:  []*dynB{{Count: 1}, {Count: 2}},
	}

	m, err := Unbind(r1)
	assert.NoError(t, err)

	// bind into interface-typed fields using per-field binders
	opts := &Options{
		FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
			"rootIface.DoA": {
				"a": func(m map[string]any) (Dynamic, error) {
					name, _ := m["name"].(string)
					return &dynA{Name: name}, nil
				},
			},
			"rootIface.Bs": {
				"b": func(m map[string]any) (Dynamic, error) {
					countAny := m["count"]
					count := 0
					switch v := countAny.(type) {
					case int:
						count = v
					case int64:
						count = int(v)
					case float64:
						count = int(v)
					}
					return &dynB{Count: count}, nil
				},
			},
		},
	}

	r2 := &rootIface{}
	err = Bind(r2, m, opts)
	assert.NoError(t, err)

	if assert.NotNil(t, r2.DoA) {
		a, ok := r2.DoA.(*dynA)
		assert.True(t, ok)
		assert.Equal(t, "alpha", a.Name)
	}
	if assert.Len(t, r2.Bs, 2) {
		b1, _ := r2.Bs[0].(*dynB)
		b2, _ := r2.Bs[1].(*dynB)
		assert.Equal(t, 1, b1.Count)
		assert.Equal(t, 2, b2.Count)
	}
}

// customBindType is a custom type for testing Unmarshaler
type customBindType struct {
	Value string
}

// UnmarshalDF implements the Unmarshaler interface for *customBindType
func (c *customBindType) UnmarshalDF(data map[string]any) error {
	if val, ok := data["value"].(string); ok {
		c.Value = "custom-" + val
		return nil
	}
	return fmt.Errorf("missing 'value' in custom data")
}

func TestBindCustomUnmarshaler(t *testing.T) {
	// test with a pointer field
	t.Run("pointer field", func(t *testing.T) {
		target := &struct {
			Custom *customBindType
		}{}
		data := map[string]any{
			"custom": map[string]any{
				"value": "hello",
			},
		}
		err := Bind(target, data)
		assert.NoError(t, err)
		if assert.NotNil(t, target.Custom) {
			assert.Equal(t, "custom-hello", target.Custom.Value)
		}
	})

	// test with a value field
	t.Run("value field", func(t *testing.T) {
		target := &struct {
			Custom customBindType
		}{}
		data := map[string]any{
			"custom": map[string]any{
				"value": "world",
			},
		}
		err := Bind(target, data)
		assert.NoError(t, err)
		assert.Equal(t, "custom-world", target.Custom.Value)
	})
}

// dependentUnmarshaler is a test type that checks if another field in the parent struct has been bound before its
// UnmarshalDF method is called.
type dependentUnmarshaler struct {
	Value          string
	CheckOtherFunc func()
}

func (d *dependentUnmarshaler) UnmarshalDF(data map[string]any) error {
	// when this is called, the check function should be able to verify that the other field is already bound.
	if d.CheckOtherFunc != nil {
		d.CheckOtherFunc()
	}
	if val, ok := data["value"].(string); ok {
		d.Value = val
	}
	return nil
}

func TestBindDeferredUnmarshaler(t *testing.T) {
	target := &struct {
		OtherField string               `df:"other_field"`
		Dep        dependentUnmarshaler `df:"dep"`
	}{}

	// check function that will be called from within UnmarshalDF
	target.Dep.CheckOtherFunc = func() {
		// this assertion runs during the Bind process. it verifies that the OtherField has already been populated.
		assert.Equal(t, "was bound", target.OtherField)
	}

	data := map[string]any{
		"other_field": "was bound",
		"dep": map[string]any{
			"value": "hello",
		},
	}

	err := Bind(target, data)
	assert.NoError(t, err)
	assert.Equal(t, "was bound", target.OtherField)
	assert.Equal(t, "hello", target.Dep.Value)
}
