package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/michaelquigley/df"
)

// Email represents a validated email address
type Email string

func (e Email) String() string {
	return string(e)
}

// EmailConverter handles conversion of string values to/from Email type
type EmailConverter struct{}

func (c *EmailConverter) FromRaw(raw interface{}) (interface{}, error) {
	s, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("expected string for email, got %T", raw)
	}
	
	// basic email validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if matched, _ := regexp.MatchString(emailRegex, s); !matched {
		return nil, fmt.Errorf("invalid email format: %s", s)
	}
	
	return Email(s), nil
}

func (c *EmailConverter) ToRaw(value interface{}) (interface{}, error) {
	email, ok := value.(Email)
	if !ok {
		return nil, fmt.Errorf("expected Email, got %T", value)
	}
	return string(email), nil
}

// Temperature represents a temperature with a unit
type Temperature struct {
	Value float64
	Unit  string // "C" or "F"
}

// TemperatureConverter handles conversion of various formats to Temperature
type TemperatureConverter struct{}

func (c *TemperatureConverter) FromRaw(raw interface{}) (interface{}, error) {
	switch v := raw.(type) {
	case string:
		// parse strings like "23.5C" or "75F"
		if len(v) < 2 {
			return nil, fmt.Errorf("invalid temperature format: %s", v)
		}
		
		unit := v[len(v)-1:]
		if unit != "C" && unit != "F" {
			return nil, fmt.Errorf("invalid temperature unit: %s (must be C or F)", unit)
		}
		
		valueStr := v[:len(v)-1]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid temperature value: %s", valueStr)
		}
		
		return Temperature{Value: value, Unit: unit}, nil
		
	case map[string]interface{}:
		// parse objects like {"value": 23.5, "unit": "C"}
		valueRaw, ok := v["value"]
		if !ok {
			return nil, fmt.Errorf("missing 'value' field in temperature object")
		}
		
		unitRaw, ok := v["unit"]
		if !ok {
			return nil, fmt.Errorf("missing 'unit' field in temperature object")
		}
		
		var value float64
		switch val := valueRaw.(type) {
		case float64:
			value = val
		case int:
			value = float64(val)
		default:
			return nil, fmt.Errorf("invalid temperature value type: %T", valueRaw)
		}
		
		unit, ok := unitRaw.(string)
		if !ok {
			return nil, fmt.Errorf("invalid temperature unit type: %T", unitRaw)
		}
		
		if unit != "C" && unit != "F" {
			return nil, fmt.Errorf("invalid temperature unit: %s (must be C or F)", unit)
		}
		
		return Temperature{Value: value, Unit: unit}, nil
		
	default:
		return nil, fmt.Errorf("unsupported temperature format: %T", raw)
	}
}

func (c *TemperatureConverter) ToRaw(value interface{}) (interface{}, error) {
	temp, ok := value.(Temperature)
	if !ok {
		return nil, fmt.Errorf("expected Temperature, got %T", value)
	}
	
	// return as string format like "23.5C"
	return fmt.Sprintf("%.1f%s", temp.Value, temp.Unit), nil
}

// TimestampConverter handles conversion of various timestamp formats to time.Time
type TimestampConverter struct{}

func (c *TimestampConverter) FromRaw(raw interface{}) (interface{}, error) {
	switch v := raw.(type) {
	case string:
		// try various timestamp formats
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("unable to parse timestamp: %s", v)
		
	case int64:
		// assume unix timestamp
		return time.Unix(v, 0), nil
		
	case float64:
		// assume unix timestamp
		return time.Unix(int64(v), 0), nil
		
	default:
		return nil, fmt.Errorf("unsupported timestamp format: %T", raw)
	}
}

func (c *TimestampConverter) ToRaw(value interface{}) (interface{}, error) {
	t, ok := value.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expected time.Time, got %T", value)
	}
	
	// return as RFC3339 string
	return t.Format(time.RFC3339), nil
}

// User represents a user with custom field types
type User struct {
	ID          int
	Email       Email
	Name        string
	Temperature Temperature `df:"preferred_temp"`
	CreatedAt   time.Time   `df:"created_at"`
}

func main() {
	fmt.Println("=== df custom converters example ===")
	fmt.Println("demonstrates how custom converters enable type-safe data binding")
	fmt.Println("with validation and support for multiple input formats")
	
	// setup converters for our custom types
	opts := &df.Options{
		Converters: map[reflect.Type]df.Converter{
			reflect.TypeOf(Email("")):       &EmailConverter{},
			reflect.TypeOf(Temperature{}):   &TemperatureConverter{},
			reflect.TypeOf(time.Time{}):     &TimestampConverter{},
		},
	}
	
	// example 1: bind from map with various input formats
	fmt.Println("\n1. binding from map with converters:")
	data := map[string]any{
		"id":             123,
		"email":          "john.doe@example.com",
		"name":           "john doe",
		"preferred_temp": "23.5C", // string format temperature
		"created_at":     "2023-12-01T10:30:00Z", // rfc3339 timestamp
	}
	
	fmt.Printf("input data: %+v\n", data)
	
	user, err := df.New[User](data, opts)
	if err != nil {
		fmt.Printf("bind failed: %v\n", err)
		return
	}
	
	fmt.Printf("bound user: %+v\n", *user)
	fmt.Printf("  email type: %T (custom type with validation)\n", user.Email)
	fmt.Printf("  temperature: %.1f°%s (parsed from string)\n", user.Temperature.Value, user.Temperature.Unit)
	fmt.Printf("  created at: %s (parsed from rfc3339)\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
	
	// example 2: bind from json with different temperature format
	fmt.Println("\n2. binding from json with object temperature:")
	jsonData := `{
		"id": 456,
		"email": "jane.smith@example.com",
		"name": "jane smith",
		"preferred_temp": {"value": 75, "unit": "F"},
		"created_at": 1701420600
	}`
	
	fmt.Printf("json input: %s\n", jsonData)
	
	var jsonMap map[string]any
	if err := json.Unmarshal([]byte(jsonData), &jsonMap); err != nil {
		fmt.Printf("json unmarshal failed: %v\n", err)
		return
	}
	
	user2, err := df.New[User](jsonMap, opts)
	if err != nil {
		fmt.Printf("bind failed: %v\n", err)
		return
	}
	
	fmt.Printf("bound user from json: %+v\n", *user2)
	fmt.Printf("  temperature: %.1f°%s (parsed from object format)\n", user2.Temperature.Value, user2.Temperature.Unit)
	fmt.Printf("  created at: %s (parsed from unix timestamp)\n", user2.CreatedAt.Format("2006-01-02 15:04:05"))
	
	// example 3: unbind back to map
	fmt.Println("\n3. unbinding back to map:")
	resultData, err := df.Unbind(user2, opts)
	if err != nil {
		fmt.Printf("unbind failed: %v\n", err)
		return
	}
	
	fmt.Printf("unbind result: %+v\n", resultData)
	
	// example 4: validation error
	fmt.Println("\n4. validation error example:")
	invalidData := map[string]any{
		"id":    789,
		"email": "invalid-email", // missing @ symbol
		"name":  "test user",
	}
	
	_, err = df.New[User](invalidData, opts)
	if err != nil {
		fmt.Printf("expected validation error: %v\n", err)
	}
	
	// example 5: without converters (should fail for custom types)
	fmt.Println("\n5. binding without converters:")
	_, err = df.New[User](data)
	if err != nil {
		fmt.Printf("expected failure without converters: %v\n", err)
	}
	
	fmt.Println("\nexample complete!")
}