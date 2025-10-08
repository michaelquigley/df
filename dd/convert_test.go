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

func TestTimeTime(t *testing.T) {
	root := &struct {
		CreatedAt time.Time
	}{}

	// test with RFC3339 string (what Unbind produces)
	data := map[string]any{
		"created_at": "2024-03-15T14:30:45Z",
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	expected := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	assert.Equal(t, expected, root.CreatedAt)
}

func TestTimeTimeWithTimeValue(t *testing.T) {
	root := &struct {
		CreatedAt time.Time
	}{}

	// test with actual time.Time value
	expected := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	data := map[string]any{
		"created_at": expected,
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, expected, root.CreatedAt)
}

func TestTimeTimeRoundTrip(t *testing.T) {
	type TestStruct struct {
		Name      string
		CreatedAt time.Time
		UpdatedAt *time.Time
	}

	updatedTime := time.Date(2024, 3, 16, 10, 20, 30, 0, time.UTC)
	original := &TestStruct{
		Name:      "test",
		CreatedAt: time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
		UpdatedAt: &updatedTime,
	}

	// unbind to map
	m, err := Unbind(original)
	assert.NoError(t, err)
	assert.Equal(t, "test", m["name"])
	assert.Equal(t, "2024-03-15T14:30:45Z", m["created_at"])
	assert.Equal(t, "2024-03-16T10:20:30Z", m["updated_at"])

	// bind back to struct
	result := &TestStruct{}
	err = Bind(result, m)
	assert.NoError(t, err)
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.CreatedAt, result.CreatedAt)
	assert.NotNil(t, result.UpdatedAt)
	assert.Equal(t, *original.UpdatedAt, *result.UpdatedAt)
}

func TestTimeTimeNew(t *testing.T) {
	type TestStruct struct {
		CreatedAt time.Time
	}

	data := map[string]any{
		"created_at": "2024-03-15T14:30:45Z",
	}

	result, err := New[TestStruct](data)
	assert.NoError(t, err)
	expected := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	assert.Equal(t, expected, result.CreatedAt)
}

func TestTimeTimeMerge(t *testing.T) {
	type TestStruct struct {
		Name      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	original := &TestStruct{
		Name:      "original",
		CreatedAt: time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
		UpdatedAt: time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
	}

	// merge only updates UpdatedAt
	data := map[string]any{
		"updated_at": "2024-03-16T10:20:30Z",
	}

	err := Merge(original, data)
	assert.NoError(t, err)
	assert.Equal(t, "original", original.Name)
	assert.Equal(t, time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC), original.CreatedAt)
	assert.Equal(t, time.Date(2024, 3, 16, 10, 20, 30, 0, time.UTC), original.UpdatedAt)
}

func TestTimeTimeRFC3339Nano(t *testing.T) {
	root := &struct {
		CreatedAt time.Time
	}{}

	// test with RFC3339Nano string (higher precision)
	data := map[string]any{
		"created_at": "2024-03-15T14:30:45.123456789Z",
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	expected := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)
	assert.Equal(t, expected, root.CreatedAt)
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
