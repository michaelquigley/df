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

	t.Run("typed maps with string keys", func(t *testing.T) {
		type StringKeyMaps struct {
			StringMap map[string]string `dd:"string_map"`
			IntMap    map[string]int    `dd:"int_map"`
			StructMap map[string]struct {
				Name  string
				Value int
			} `dd:"struct_map"`
		}

		data := map[string]any{
			"string_map": map[string]any{"key1": "value1", "key2": "value2"},
			"int_map":    map[string]any{"one": 1, "two": 2, "three": 3},
			"struct_map": map[string]any{
				"item1": map[string]any{"name": "first", "value": 10},
				"item2": map[string]any{"name": "second", "value": 20},
			},
		}

		var target StringKeyMaps
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.Equal(t, "value1", target.StringMap["key1"])
		assert.Equal(t, "value2", target.StringMap["key2"])
		assert.Equal(t, 1, target.IntMap["one"])
		assert.Equal(t, 2, target.IntMap["two"])
		assert.Equal(t, 3, target.IntMap["three"])
		assert.Equal(t, "first", target.StructMap["item1"].Name)
		assert.Equal(t, 10, target.StructMap["item1"].Value)
		assert.Equal(t, "second", target.StructMap["item2"].Name)
		assert.Equal(t, 20, target.StructMap["item2"].Value)
	})

	t.Run("typed maps with int keys", func(t *testing.T) {
		type IntKeyMaps struct {
			IntMap   map[int]string    `dd:"int_map"`
			Int64Map map[int64]float64 `dd:"int64_map"`
		}

		// JSON has string keys, they will be coerced to int
		data := map[string]any{
			"int_map":   map[string]any{"1": "one", "2": "two", "42": "answer"},
			"int64_map": map[string]any{"10": 1.5, "20": 2.5, "30": 3.5},
		}

		var target IntKeyMaps
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.Equal(t, "one", target.IntMap[1])
		assert.Equal(t, "two", target.IntMap[2])
		assert.Equal(t, "answer", target.IntMap[42])
		assert.Equal(t, 1.5, target.Int64Map[10])
		assert.Equal(t, 2.5, target.Int64Map[20])
		assert.Equal(t, 3.5, target.Int64Map[30])
	})

	t.Run("typed maps with uint keys", func(t *testing.T) {
		type UintKeyMaps struct {
			UintMap   map[uint]string   `dd:"uint_map"`
			Uint32Map map[uint32]bool   `dd:"uint32_map"`
		}

		data := map[string]any{
			"uint_map":   map[string]any{"1": "first", "2": "second"},
			"uint32_map": map[string]any{"10": true, "20": false},
		}

		var target UintKeyMaps
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.Equal(t, "first", target.UintMap[1])
		assert.Equal(t, "second", target.UintMap[2])
		assert.Equal(t, true, target.Uint32Map[10])
		assert.Equal(t, false, target.Uint32Map[20])
	})

	t.Run("nested typed maps", func(t *testing.T) {
		type NestedMaps struct {
			Outer map[string]map[string]int `dd:"outer"`
		}

		data := map[string]any{
			"outer": map[string]any{
				"group1": map[string]any{"a": 1, "b": 2},
				"group2": map[string]any{"c": 3, "d": 4},
			},
		}

		var target NestedMaps
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.Equal(t, 1, target.Outer["group1"]["a"])
		assert.Equal(t, 2, target.Outer["group1"]["b"])
		assert.Equal(t, 3, target.Outer["group2"]["c"])
		assert.Equal(t, 4, target.Outer["group2"]["d"])
	})

	t.Run("maps with pointer values", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		type PointerMaps struct {
			Users map[string]*User `dd:"users"`
		}

		data := map[string]any{
			"users": map[string]any{
				"user1": map[string]any{"name": "Alice", "age": 30},
				"user2": map[string]any{"name": "Bob", "age": 25},
			},
		}

		var target PointerMaps
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.NotNil(t, target.Users["user1"])
		assert.Equal(t, "Alice", target.Users["user1"].Name)
		assert.Equal(t, 30, target.Users["user1"].Age)
		assert.NotNil(t, target.Users["user2"])
		assert.Equal(t, "Bob", target.Users["user2"].Name)
		assert.Equal(t, 25, target.Users["user2"].Age)
	})

	t.Run("maps with slice values", func(t *testing.T) {
		type SliceMaps struct {
			Groups map[string][]string `dd:"groups"`
		}

		data := map[string]any{
			"groups": map[string]any{
				"colors": []any{"red", "green", "blue"},
				"fruits": []any{"apple", "banana", "orange"},
			},
		}

		var target SliceMaps
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.Equal(t, []string{"red", "green", "blue"}, target.Groups["colors"])
		assert.Equal(t, []string{"apple", "banana", "orange"}, target.Groups["fruits"])
	})

	t.Run("typed maps unbind", func(t *testing.T) {
		type TypedMaps struct {
			IntMap map[int]string    `dd:"int_map"`
			StrMap map[string]int    `dd:"str_map"`
		}

		source := TypedMaps{
			IntMap: map[int]string{1: "one", 2: "two", 42: "answer"},
			StrMap: map[string]int{"a": 1, "b": 2},
		}

		result, err := Unbind(source)
		assert.NoError(t, err)

		// keys should be converted to strings in result
		intMap := result["int_map"].(map[string]any)
		assert.Equal(t, "one", intMap["1"])
		assert.Equal(t, "two", intMap["2"])
		assert.Equal(t, "answer", intMap["42"])

		strMap := result["str_map"].(map[string]any)
		assert.Equal(t, 1, strMap["a"])
		assert.Equal(t, 2, strMap["b"])
	})

	t.Run("typed maps roundtrip", func(t *testing.T) {
		type Config struct {
			Servers map[int]struct {
				Host string
				Port int
			} `dd:"servers"`
		}

		original := Config{
			Servers: map[int]struct{ Host string; Port int }{
				1: {"localhost", 8080},
				2: {"api.example.com", 443},
			},
		}

		// unbind to map
		data, err := Unbind(original)
		assert.NoError(t, err)

		// bind back to struct
		var restored Config
		err = Bind(&restored, data)
		assert.NoError(t, err)

		assert.Equal(t, "localhost", restored.Servers[1].Host)
		assert.Equal(t, 8080, restored.Servers[1].Port)
		assert.Equal(t, "api.example.com", restored.Servers[2].Host)
		assert.Equal(t, 443, restored.Servers[2].Port)
	})

	t.Run("typed maps with bool keys", func(t *testing.T) {
		type BoolKeyMap struct {
			Flags map[bool]string `dd:"flags"`
		}

		data := map[string]any{
			"flags": map[string]any{
				"true":  "enabled",
				"false": "disabled",
			},
		}

		var target BoolKeyMap
		err := Bind(&target, data)
		assert.NoError(t, err)
		assert.Equal(t, "enabled", target.Flags[true])
		assert.Equal(t, "disabled", target.Flags[false])
	})
}

func TestMapInspect(t *testing.T) {
	type StructWithMap struct {
		Name string         `dd:"name"`
		Data map[string]any `dd:"data"`
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

	t.Run("merge typed maps with int keys", func(t *testing.T) {
		type TypedMapStruct struct {
			Name    string         `dd:"name"`
			Servers map[int]string `dd:"servers"`
			Count   int            `dd:"count"`
		}

		target := TypedMapStruct{
			Name: "original",
			Servers: map[int]string{
				1: "old-server1",
				2: "old-server2",
			},
			Count: 5,
		}

		mergeData := map[string]any{
			"servers": map[string]any{
				"1":  "new-server1",
				"3":  "new-server3",
				"10": "new-server10",
			},
			"count": 10,
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		// Name should be preserved (not in merge data)
		assert.Equal(t, "original", target.Name)

		// Count should be updated
		assert.Equal(t, 10, target.Count)

		// Servers map should be completely replaced with key coercion
		assert.Equal(t, "new-server1", target.Servers[1])
		assert.Equal(t, "new-server3", target.Servers[3])
		assert.Equal(t, "new-server10", target.Servers[10])
		_, exists := target.Servers[2]
		assert.False(t, exists, "old key should be gone after merge")
	})

	t.Run("merge typed maps with struct values", func(t *testing.T) {
		type ServerConfig struct {
			Host string
			Port int
		}
		type ClusterConfig struct {
			Name    string                   `dd:"name"`
			Servers map[string]ServerConfig `dd:"servers"`
		}

		target := ClusterConfig{
			Name: "cluster1",
			Servers: map[string]ServerConfig{
				"server1": {"old-host1", 8080},
				"server2": {"old-host2", 8080},
			},
		}

		mergeData := map[string]any{
			"servers": map[string]any{
				"server1": map[string]any{"host": "new-host1", "port": 9090},
				"server3": map[string]any{"host": "new-host3", "port": 9091},
			},
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		assert.Equal(t, "cluster1", target.Name)
		assert.Equal(t, "new-host1", target.Servers["server1"].Host)
		assert.Equal(t, 9090, target.Servers["server1"].Port)
		assert.Equal(t, "new-host3", target.Servers["server3"].Host)
		assert.Equal(t, 9091, target.Servers["server3"].Port)
		_, exists := target.Servers["server2"]
		assert.False(t, exists, "old server2 should be gone after merge")
	})

	t.Run("merge preserves typed maps when not in merge data", func(t *testing.T) {
		type TypedMapStruct struct {
			Name    string         `dd:"name"`
			Servers map[int]string `dd:"servers"`
			Count   int            `dd:"count"`
		}

		target := TypedMapStruct{
			Name: "original",
			Servers: map[int]string{
				1: "server1",
				2: "server2",
			},
			Count: 5,
		}

		mergeData := map[string]any{
			"name": "updated",
			// no "servers" field in merge data
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		// Name should be updated
		assert.Equal(t, "updated", target.Name)

		// Count should be preserved
		assert.Equal(t, 5, target.Count)

		// Servers map should be completely preserved
		assert.Equal(t, "server1", target.Servers[1])
		assert.Equal(t, "server2", target.Servers[2])
	})

	t.Run("merge typed maps with nil maps", func(t *testing.T) {
		type TypedMapStruct struct {
			Name    string         `dd:"name"`
			Servers map[int]string `dd:"servers"`
		}

		target := TypedMapStruct{
			Name:    "test",
			Servers: nil,
		}

		mergeData := map[string]any{
			"servers": map[string]any{
				"1": "server1",
				"2": "server2",
			},
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		assert.Equal(t, "test", target.Name)
		assert.NotNil(t, target.Servers)
		assert.Equal(t, "server1", target.Servers[1])
		assert.Equal(t, "server2", target.Servers[2])
	})

	t.Run("merge nested typed maps", func(t *testing.T) {
		type NestedTypedMaps struct {
			Name string                          `dd:"name"`
			Envs map[string]map[string]string `dd:"envs"`
		}

		target := NestedTypedMaps{
			Name: "config",
			Envs: map[string]map[string]string{
				"dev": {
					"db_host": "localhost",
					"api_url": "http://localhost:8080",
				},
			},
		}

		mergeData := map[string]any{
			"envs": map[string]any{
				"dev": map[string]any{
					"db_host": "dev-db.example.com",
					"api_url": "http://dev-api.example.com",
				},
				"prod": map[string]any{
					"db_host": "prod-db.example.com",
					"api_url": "https://api.example.com",
				},
			},
		}

		err := Merge(&target, mergeData)
		assert.NoError(t, err)

		assert.Equal(t, "config", target.Name)
		assert.Equal(t, "dev-db.example.com", target.Envs["dev"]["db_host"])
		assert.Equal(t, "http://dev-api.example.com", target.Envs["dev"]["api_url"])
		assert.Equal(t, "prod-db.example.com", target.Envs["prod"]["db_host"])
		assert.Equal(t, "https://api.example.com", target.Envs["prod"]["api_url"])
	})
}
