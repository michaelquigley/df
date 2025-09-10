package df

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test basic embedded struct functionality
func TestEmbeddedStructBasics(t *testing.T) {
	type Person struct {
		Name string `df:"name"`
		Age  int    `df:"age"`
	}

	type Employee struct {
		Person // embedded struct
		Title  string `df:"title"`
		Salary int    `df:"salary"`
	}

	t.Run("bind with embedded struct", func(t *testing.T) {
		data := map[string]any{
			"name":   "John Doe",
			"age":    30,
			"title":  "Software Engineer",
			"salary": 75000,
		}

		var emp Employee
		err := Bind(&emp, data)

		assert.Nil(t, err)
		assert.Equal(t, "John Doe", emp.Name)
		assert.Equal(t, 30, emp.Age)
		assert.Equal(t, "Software Engineer", emp.Title)
		assert.Equal(t, 75000, emp.Salary)
	})

	t.Run("unbind with embedded struct", func(t *testing.T) {
		emp := Employee{
			Person: Person{
				Name: "Jane Smith",
				Age:  28,
			},
			Title:  "DevOps Engineer",
			Salary: 80000,
		}

		result, err := Unbind(emp)

		assert.Nil(t, err)
		expected := map[string]any{
			"name":   "Jane Smith",
			"age":    28,
			"title":  "DevOps Engineer",
			"salary": 80000,
		}
		assert.Equal(t, expected, result)
	})

	t.Run("new with embedded struct", func(t *testing.T) {
		data := map[string]any{
			"name":   "Bob Wilson",
			"age":    35,
			"title":  "Team Lead",
			"salary": 90000,
		}

		result, err := New[Employee](data)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Bob Wilson", result.Name)
		assert.Equal(t, 35, result.Age)
		assert.Equal(t, "Team Lead", result.Title)
		assert.Equal(t, 90000, result.Salary)
	})

	t.Run("merge with embedded struct", func(t *testing.T) {
		emp := Employee{
			Person: Person{
				Name: "Alice Cooper",
				Age:  40,
			},
			Title:  "Manager",
			Salary: 95000,
		}

		data := map[string]any{
			"age":   41,
			"title": "Senior Manager",
		}

		err := Merge(&emp, data)

		assert.Nil(t, err)
		assert.Equal(t, "Alice Cooper", emp.Name)
		assert.Equal(t, 41, emp.Age)
		assert.Equal(t, "Senior Manager", emp.Title)
		assert.Equal(t, 95000, emp.Salary)
	})
}

// Test multiple level embedding
func TestMultiLevelEmbedding(t *testing.T) {
	type Contact struct {
		Email string `df:"email"`
		Phone string `df:"phone"`
	}

	type Person struct {
		Name string `df:"name"`
		Age  int    `df:"age"`
	}

	type Employee struct {
		Person
		Contact
		Title string `df:"title"`
	}

	t.Run("bind with multiple embedded structs", func(t *testing.T) {
		data := map[string]any{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
			"phone": "555-1234",
			"title": "Developer",
		}

		var emp Employee
		err := Bind(&emp, data)

		assert.Nil(t, err)
		assert.Equal(t, "John Doe", emp.Name)
		assert.Equal(t, 30, emp.Age)
		assert.Equal(t, "john@example.com", emp.Email)
		assert.Equal(t, "555-1234", emp.Phone)
		assert.Equal(t, "Developer", emp.Title)
	})

	t.Run("unbind with multiple embedded structs", func(t *testing.T) {
		emp := Employee{
			Person: Person{
				Name: "Jane Smith",
				Age:  28,
			},
			Contact: Contact{
				Email: "jane@example.com",
				Phone: "555-5678",
			},
			Title: "Designer",
		}

		result, err := Unbind(emp)

		assert.Nil(t, err)
		expected := map[string]any{
			"name":  "Jane Smith",
			"age":   28,
			"email": "jane@example.com",
			"phone": "555-5678",
			"title": "Designer",
		}
		assert.Equal(t, expected, result)
	})
}

// Test pointer embedded structs
func TestPointerEmbeddedStruct(t *testing.T) {
	type Address struct {
		Street string `df:"street"`
		City   string `df:"city"`
	}

	type Person struct {
		*Address
		Name string `df:"name"`
	}

	t.Run("bind with pointer embedded struct", func(t *testing.T) {
		data := map[string]any{
			"name":   "John Doe",
			"street": "123 Main St",
			"city":   "Springfield",
		}

		var person Person
		err := Bind(&person, data)

		assert.Nil(t, err)
		assert.Equal(t, "John Doe", person.Name)
		assert.NotNil(t, person.Address)
		assert.Equal(t, "123 Main St", person.Address.Street)
		assert.Equal(t, "Springfield", person.Address.City)
	})

	t.Run("unbind with pointer embedded struct", func(t *testing.T) {
		person := Person{
			Address: &Address{
				Street: "456 Oak Ave",
				City:   "Shelbyville",
			},
			Name: "Jane Smith",
		}

		result, err := Unbind(person)

		assert.Nil(t, err)
		expected := map[string]any{
			"name":   "Jane Smith",
			"street": "456 Oak Ave",
			"city":   "Shelbyville",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("unbind with nil pointer embedded struct", func(t *testing.T) {
		person := Person{
			Address: nil,
			Name:    "Bob Wilson",
		}

		result, err := Unbind(person)

		assert.Nil(t, err)
		expected := map[string]any{
			"name": "Bob Wilson",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("bind partial data with pointer embedded struct", func(t *testing.T) {
		data := map[string]any{
			"name": "Alice",
		}

		var person Person
		err := Bind(&person, data)

		assert.Nil(t, err)
		assert.Equal(t, "Alice", person.Name)
		assert.Nil(t, person.Address) // should remain nil when no embedded fields provided
	})
}

// Test embedded struct tags
func TestEmbeddedStructTags(t *testing.T) {
	type Base struct {
		ID       string `df:"id"`
		Name     string `df:"custom_name"`
		Internal string `df:"-"`
		Secret   string `df:",+secret"`
	}

	type Extended struct {
		Base
		Value string `df:"value"`
	}

	t.Run("bind respects embedded struct tags", func(t *testing.T) {
		data := map[string]any{
			"id":          "123",
			"custom_name": "test",
			"internal":    "should be ignored",
			"secret":      "hidden",
			"value":       "extended",
		}

		var ext Extended
		err := Bind(&ext, data)

		assert.Nil(t, err)
		assert.Equal(t, "123", ext.ID)
		assert.Equal(t, "test", ext.Name)
		assert.Equal(t, "", ext.Internal) // should be empty due to df:"-"
		assert.Equal(t, "hidden", ext.Secret)
		assert.Equal(t, "extended", ext.Value)
	})

	t.Run("unbind respects embedded struct tags", func(t *testing.T) {
		ext := Extended{
			Base: Base{
				ID:       "456",
				Name:     "test2",
				Internal: "ignored",
				Secret:   "secret",
			},
			Value: "extended2",
		}

		result, err := Unbind(ext)

		assert.Nil(t, err)
		expected := map[string]any{
			"id":          "456",
			"custom_name": "test2",
			"secret":      "secret",
			"value":       "extended2",
		}
		assert.Equal(t, expected, result)
		assert.NotContains(t, result, "internal")
	})
}

// Test nested embedded structs
func TestNestedEmbeddedStructs(t *testing.T) {
	type Core struct {
		ID string `df:"id"`
	}

	type Base struct {
		Core
		Name string `df:"name"`
	}

	type Extended struct {
		Base
		Value string `df:"value"`
	}

	t.Run("bind with nested embedded structs", func(t *testing.T) {
		data := map[string]any{
			"id":    "123",
			"name":  "test",
			"value": "extended",
		}

		var ext Extended
		err := Bind(&ext, data)

		assert.Nil(t, err)
		assert.Equal(t, "123", ext.ID)
		assert.Equal(t, "test", ext.Name)
		assert.Equal(t, "extended", ext.Value)
	})

	t.Run("unbind with nested embedded structs", func(t *testing.T) {
		ext := Extended{
			Base: Base{
				Core: Core{ID: "456"},
				Name: "test2",
			},
			Value: "extended2",
		}

		result, err := Unbind(ext)

		assert.Nil(t, err)
		expected := map[string]any{
			"id":    "456",
			"name":  "test2",
			"value": "extended2",
		}
		assert.Equal(t, expected, result)
	})
}

// Test embedded struct with required fields
func TestEmbeddedStructRequired(t *testing.T) {
	type Base struct {
		Required string `df:"+required"`
		Optional string `df:"optional"`
	}

	type Extended struct {
		Base
		Value string `df:"value"`
	}

	t.Run("bind fails when embedded required field missing", func(t *testing.T) {
		data := map[string]any{
			"optional": "present",
			"value":    "extended",
		}

		var ext Extended
		err := Bind(&ext, data)

		assert.NotNil(t, err)
		assert.IsType(t, &RequiredFieldError{}, err)
	})

	t.Run("bind succeeds when embedded required field present", func(t *testing.T) {
		data := map[string]any{
			"required": "present",
			"optional": "also present",
			"value":    "extended",
		}

		var ext Extended
		err := Bind(&ext, data)

		assert.Nil(t, err)
		assert.Equal(t, "present", ext.Required)
		assert.Equal(t, "also present", ext.Optional)
		assert.Equal(t, "extended", ext.Value)
	})
}

// Test embedded struct inspect
func TestEmbeddedStructInspect(t *testing.T) {
	type Person struct {
		Name string `df:"name"`
		Age  int    `df:"age"`
	}

	type Employee struct {
		Person
		Title  string `df:"title"`
		Salary int    `df:"salary"`
	}

	emp := Employee{
		Person: Person{
			Name: "John Doe",
			Age:  30,
		},
		Title:  "Software Engineer",
		Salary: 75000,
	}

	result, err := Inspect(emp)
	assert.Nil(t, err)
	
	// Check that all fields are present and flattened
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "age")
	assert.Contains(t, result, "title")
	assert.Contains(t, result, "salary")
	assert.Contains(t, result, "John Doe")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "Software Engineer")
	assert.Contains(t, result, "75000")
}

// Test field name conflicts in embedded structs
func TestEmbeddedStructFieldConflicts(t *testing.T) {
	type Base1 struct {
		Name string `df:"name"`
	}

	type Base2 struct {
		Name string `df:"name"`
	}

	type Conflicted struct {
		Base1
		Base2
		Value string `df:"value"`
	}

	t.Run("bind with field name conflicts", func(t *testing.T) {
		data := map[string]any{
			"name":  "test",
			"value": "extended",
		}

		var conf Conflicted
		err := Bind(&conf, data)

		assert.Nil(t, err)
		// Last embedded struct field should win
		assert.Equal(t, "test", conf.Base2.Name)
		assert.Equal(t, "extended", conf.Value)
	})

	t.Run("unbind with field name conflicts", func(t *testing.T) {
		conf := Conflicted{
			Base1: Base1{Name: "first"},
			Base2: Base2{Name: "second"},
			Value: "extended",
		}

		result, err := Unbind(conf)

		assert.Nil(t, err)
		// Should contain the last value encountered
		assert.Equal(t, "second", result["name"])
		assert.Equal(t, "extended", result["value"])
	})
}

// Test edge cases and error conditions
func TestEmbeddedStructEdgeCases(t *testing.T) {
	t.Run("empty embedded struct", func(t *testing.T) {
		type Empty struct{}
		type WithEmpty struct {
			Empty
			Value string `df:"value"`
		}

		data := map[string]any{"value": "test"}
		var we WithEmpty
		err := Bind(&we, data)

		assert.Nil(t, err)
		assert.Equal(t, "test", we.Value)

		result, err := Unbind(we)
		assert.Nil(t, err)
		assert.Equal(t, map[string]any{"value": "test"}, result)
	})

	t.Run("embedded struct with unexported fields", func(t *testing.T) {
		type Base struct {
			Name     string `df:"name"`
			internal string
		}
		type Extended struct {
			Base
			Value string `df:"value"`
		}

		data := map[string]any{"name": "test", "value": "extended", "internal": "ignored"}
		var ext Extended
		err := Bind(&ext, data)

		assert.Nil(t, err)
		assert.Equal(t, "test", ext.Name)
		assert.Equal(t, "", ext.Base.internal) // unexported field not bound
		assert.Equal(t, "extended", ext.Value)
	})

	t.Run("embedded struct with complex nested types", func(t *testing.T) {
		type Address struct {
			Street string   `df:"street"`
			Cities []string `df:"cities"`
		}
		type Person struct {
			Address
			Name string `df:"name"`
		}

		data := map[string]any{
			"name":   "John",
			"street": "123 Main St",
			"cities": []string{"NYC", "LA"},
		}

		var person Person
		err := Bind(&person, data)

		assert.Nil(t, err)
		assert.Equal(t, "John", person.Name)
		assert.Equal(t, "123 Main St", person.Street)
		assert.Equal(t, []string{"NYC", "LA"}, person.Cities)

		result, err := Unbind(person)
		assert.Nil(t, err)
		expected := map[string]any{
			"name":   "John",
			"street": "123 Main St",
			"cities": []interface{}{"NYC", "LA"},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("deeply nested pointer embedded structs", func(t *testing.T) {
		type Level3 struct {
			Deep string `df:"deep"`
		}
		type Level2 struct {
			*Level3
			Mid string `df:"mid"`
		}
		type Level1 struct {
			*Level2
			Top string `df:"top"`
		}

		data := map[string]any{
			"top":  "level1",
			"mid":  "level2", 
			"deep": "level3",
		}

		var l1 Level1
		err := Bind(&l1, data)

		assert.Nil(t, err)
		assert.Equal(t, "level1", l1.Top)
		assert.NotNil(t, l1.Level2)
		assert.Equal(t, "level2", l1.Level2.Mid)
		assert.NotNil(t, l1.Level2.Level3)
		assert.Equal(t, "level3", l1.Level2.Level3.Deep)

		result, err := Unbind(l1)
		assert.Nil(t, err)
		expected := map[string]any{
			"top":  "level1",
			"mid":  "level2",
			"deep": "level3",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("embedded struct with nil pointer partial data", func(t *testing.T) {
		type Address struct {
			Street string `df:"street"`
			City   string `df:"city"`
		}
		type Person struct {
			*Address
			Name string `df:"name"`
		}

		// Only bind name, no address fields
		data := map[string]any{"name": "John"}
		var person Person
		err := Bind(&person, data)

		assert.Nil(t, err)
		assert.Equal(t, "John", person.Name)
		assert.Nil(t, person.Address) // should remain nil

		result, err := Unbind(person)
		assert.Nil(t, err)
		assert.Equal(t, map[string]any{"name": "John"}, result)
	})
}