package dd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSupport(t *testing.T) {
	type StructWithMaps struct {
		Name     string
		Data     map[string]any
		Metadata map[string]interface{}
		Count    int
	}

	t.Run("bind with map[string]any", func(t *testing.T) {
		data := map[string]any{
			"name": "test",
			"data": map[string]any{
				"key1": "value1",
				"key2": 42,
				"key3": true,
				"key4": nil,
			},
			"metadata": map[string]any{
				"version": "1.0",
				"enabled": true,
			},
			"count": 10,
		}

		var target StructWithMaps
		err := Bind(&target, data)
		assert.NoError(t, err)

		assert.Equal(t, "test", target.Name)
		assert.Equal(t, 10, target.Count)
		assert.Equal(t, "value1", target.Data["key1"])
		assert.Equal(t, 42, target.Data["key2"])
		assert.Equal(t, true, target.Data["key3"])
		assert.Nil(t, target.Data["key4"])
		assert.Equal(t, "1.0", target.Metadata["version"])
		assert.Equal(t, true, target.Metadata["enabled"])
	})

	t.Run("unbind with map[string]any", func(t *testing.T) {
		source := StructWithMaps{
			Name: "test",
			Data: map[string]any{
				"key1": "value1",
				"key2": 42,
				"key3": true,
				"key4": nil,
			},
			Metadata: map[string]interface{}{
				"version": "1.0",
				"enabled": true,
			},
			Count: 10,
		}

		result, err := Unbind(source)
		assert.NoError(t, err)

		assert.Equal(t, "test", result["name"])
		assert.Equal(t, 10, result["count"])

		dataMap := result["data"].(map[string]any)
		assert.Equal(t, "value1", dataMap["key1"])
		assert.Equal(t, 42, dataMap["key2"])
		assert.Equal(t, true, dataMap["key3"])
		assert.Nil(t, dataMap["key4"])

		metadataMap := result["metadata"].(map[string]any)
		assert.Equal(t, "1.0", metadataMap["version"])
		assert.Equal(t, true, metadataMap["enabled"])
	})

	t.Run("roundtrip bind/unbind", func(t *testing.T) {
		original := map[string]any{
			"name": "roundtrip",
			"data": map[string]any{
				"nested": map[string]any{
					"deep": "value",
				},
				"array": []interface{}{1, 2, 3},
			},
			"count": 42,
		}

		var intermediate StructWithMaps
		err := Bind(&intermediate, original)
		assert.NoError(t, err)

		result, err := Unbind(intermediate)
		assert.NoError(t, err)

		assert.Equal(t, "roundtrip", result["name"])
		assert.Equal(t, 42, result["count"])

		dataMap := result["data"].(map[string]any)

		// test nested map
		nestedMap := dataMap["nested"].(map[string]any)
		assert.Equal(t, "value", nestedMap["deep"])

		// test array
		arraySlice := dataMap["array"].([]interface{})
		assert.Len(t, arraySlice, 3)
		assert.Equal(t, 1, arraySlice[0])
		assert.Equal(t, 2, arraySlice[1])
		assert.Equal(t, 3, arraySlice[2])
	})

	t.Run("empty maps", func(t *testing.T) {
		data := map[string]any{
			"name":  "empty",
			"data":  map[string]any{},
			"count": 0,
		}

		var target StructWithMaps
		err := Bind(&target, data)
		assert.NoError(t, err)

		assert.Equal(t, "empty", target.Name)
		assert.NotNil(t, target.Data)
		assert.Len(t, target.Data, 0)

		result, err := Unbind(target)
		assert.NoError(t, err)

		dataMap := result["data"].(map[string]any)
		assert.Len(t, dataMap, 0)
	})

	t.Run("nil maps", func(t *testing.T) {
		source := StructWithMaps{
			Name:  "nil",
			Data:  nil,
			Count: 5,
		}

		result, err := Unbind(source)
		assert.NoError(t, err)

		assert.Equal(t, "nil", result["name"])
		assert.Equal(t, 5, result["count"])

		dataMap := result["data"].(map[string]any)
		assert.Len(t, dataMap, 0)
	})

	t.Run("unsupported map types", func(t *testing.T) {
		type UnsupportedMap struct {
			IntMap map[int]string `df:"int_map"`
		}

		data := map[string]any{
			"int_map": map[string]any{"1": "one"},
		}

		var target UnsupportedMap
		err := Bind(&target, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only map[string]any and map[string]interface{} are supported")
	})
}

func TestMapInspect(t *testing.T) {
	type StructWithMap struct {
		Name string         `df:"name"`
		Data map[string]any `df:"data"`
	}

	t.Run("inspect map fields", func(t *testing.T) {
		source := StructWithMap{
			Name: "inspect_test",
			Data: map[string]any{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
		}

		result, err := Inspect(source)
		fmt.Println(result)
		assert.NoError(t, err)
		assert.Contains(t, result, "StructWithMap")
		assert.Contains(t, result, "name:")
		assert.Contains(t, result, "data:")
		assert.Contains(t, result, "key1")
		assert.Contains(t, result, "value1")
	})

	t.Run("inspect empty map", func(t *testing.T) {
		source := StructWithMap{
			Name: "empty",
			Data: map[string]any{},
		}

		result, err := Inspect(source)
		fmt.Println(result)
		assert.NoError(t, err)
		assert.Contains(t, result, "{}")
	})

	t.Run("inspect nil map", func(t *testing.T) {
		source := StructWithMap{
			Name: "nil",
			Data: nil,
		}

		result, err := Inspect(source)
		fmt.Println(result)
		assert.NoError(t, err)
		assert.Contains(t, result, "<nil map>")
	})
}

func TestMapMerge(t *testing.T) {
	type StructWithMaps struct {
		Name     string
		Data     map[string]any
		Metadata map[string]interface{}
		Count    int
	}

	t.Run("merge replaces entire map", func(t *testing.T) {
		target := StructWithMaps{
			Name: "original",
			Data: map[string]any{
				"existing": "value",
				"keep":     "this",
			},
			Metadata: map[string]interface{}{
				"version": "1.0",
				"old":     "data",
			},
			Count: 5,
		}

		mergeData := map[string]any{
			"data": map[string]any{
				"new_key":  "new_value",
				"existing": "updated",
			},
			"metadata": map[string]any{
				"version": "2.0",
				"new":     "metadata",
			},
			"count": 10,
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		// Name should be preserved (not in merge data)
		assert.Equal(t, "original", target.Name)

		// Count should be updated
		assert.Equal(t, 10, target.Count)

		// Data map should be completely replaced
		assert.Equal(t, "new_value", target.Data["new_key"])
		assert.Equal(t, "updated", target.Data["existing"])
		_, exists := target.Data["keep"]
		assert.False(t, exists, "old key should be gone after merge")

		// Metadata map should be completely replaced
		assert.Equal(t, "2.0", target.Metadata["version"])
		assert.Equal(t, "metadata", target.Metadata["new"])
		_, exists = target.Metadata["old"]
		assert.False(t, exists, "old metadata should be gone after merge")
	})

	t.Run("merge with nil maps", func(t *testing.T) {
		target := StructWithMaps{
			Name:     "test",
			Data:     nil,
			Metadata: nil,
			Count:    1,
		}

		mergeData := map[string]any{
			"data": map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			"metadata": map[string]any{
				"enabled": true,
			},
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		assert.Equal(t, "test", target.Name)
		assert.Equal(t, 1, target.Count)

		assert.NotNil(t, target.Data)
		assert.Equal(t, "value1", target.Data["key1"])
		assert.Equal(t, 42, target.Data["key2"])

		assert.NotNil(t, target.Metadata)
		assert.Equal(t, true, target.Metadata["enabled"])
	})

	t.Run("merge preserves maps when not in merge data", func(t *testing.T) {
		target := StructWithMaps{
			Name: "original",
			Data: map[string]any{
				"preserve": "this",
				"and":      "this too",
			},
			Metadata: map[string]interface{}{
				"keep": "metadata",
			},
			Count: 5,
		}

		mergeData := map[string]any{
			"name": "updated",
			// no "data" or "metadata" fields in merge data
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		// Name should be updated
		assert.Equal(t, "updated", target.Name)

		// Count should be preserved
		assert.Equal(t, 5, target.Count)

		// Maps should be completely preserved
		assert.Equal(t, "this", target.Data["preserve"])
		assert.Equal(t, "this too", target.Data["and"])
		assert.Equal(t, "metadata", target.Metadata["keep"])
	})

	t.Run("merge with empty maps", func(t *testing.T) {
		target := StructWithMaps{
			Name: "test",
			Data: map[string]any{
				"existing": "data",
			},
			Metadata: map[string]interface{}{
				"old": "meta",
			},
			Count: 1,
		}

		mergeData := map[string]any{
			"data":     map[string]any{}, // empty map
			"metadata": map[string]any{}, // empty map
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		assert.Equal(t, "test", target.Name)
		assert.Equal(t, 1, target.Count)
		assert.Len(t, target.Data, 0)     // should be empty
		assert.Len(t, target.Metadata, 0) // should be empty
	})

	t.Run("merge roundtrip", func(t *testing.T) {
		original := StructWithMaps{
			Name: "roundtrip",
			Data: map[string]any{
				"complex": map[string]any{
					"nested": "value",
				},
				"array": []interface{}{1, 2, 3},
			},
			Count: 42,
		}

		// Convert to map
		asMap, err := Unbind(original)
		assert.NoError(t, err)

		// Create new target and merge
		target := StructWithMaps{
			Name:  "different",
			Count: 99,
		}

		err = Merge(&target, asMap)
		assert.NoError(t, err)

		// Should have original's data but replaced name/count from merge
		assert.Equal(t, "roundtrip", target.Name) // replaced by merge
		assert.Equal(t, 42, target.Count)         // replaced by merge

		dataMap := target.Data
		nestedMap := dataMap["complex"].(map[string]any)
		assert.Equal(t, "value", nestedMap["nested"])

		arraySlice := dataMap["array"].([]interface{})
		assert.Len(t, arraySlice, 3)
		assert.Equal(t, 1, arraySlice[0])
	})
}
