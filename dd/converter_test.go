package dd

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Email is a custom type that we want to handle with a converter
type Email string

func (e Email) String() string {
	return string(e)
}

// EmailConverter implements the Converter interface for Email types
type EmailConverter struct{}

func (c *EmailConverter) FromRaw(raw interface{}) (interface{}, error) {
	s, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", raw)
	}
	if !strings.Contains(s, "@") {
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

// UserID is a custom numeric type
type UserID int

// UserIDConverter implements the Converter interface for UserID types
type UserIDConverter struct{}

func (c *UserIDConverter) FromRaw(raw interface{}) (interface{}, error) {
	switch v := raw.(type) {
	case int:
		return UserID(v), nil
	case int64:
		return UserID(v), nil
	case float64:
		return UserID(int(v)), nil
	case string:
		// try to parse as int
		if i, err := strconv.Atoi(v); err == nil {
			return UserID(i), nil
		}
		return nil, fmt.Errorf("cannot parse UserID from string: %s", v)
	default:
		return nil, fmt.Errorf("cannot convert %T to UserID", raw)
	}
}

func (c *UserIDConverter) ToRaw(value interface{}) (interface{}, error) {
	userID, ok := value.(UserID)
	if !ok {
		return nil, fmt.Errorf("expected UserID, got %T", value)
	}
	return int(userID), nil
}

// Test struct for binding
type TestUser struct {
	ID    UserID `dd:"user_id"`
	Email Email  `dd:"email"`
	Name  string `dd:"name"`
}

func TestConverterBind(t *testing.T) {
	// setup converters
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &EmailConverter{},
			reflect.TypeOf(UserID(0)): &UserIDConverter{},
		},
	}

	// test data
	data := map[string]any{
		"user_id": 123,
		"email":   "test@example.com",
		"name":    "john doe",
	}

	// bind with converters
	var user TestUser
	err := Bind(&user, data, opts)
	assert.NoError(t, err)

	// verify results
	assert.Equal(t, UserID(123), user.ID)
	assert.Equal(t, Email("test@example.com"), user.Email)
	assert.Equal(t, "john doe", user.Name)
}

func TestConverterBindInvalidEmail(t *testing.T) {
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &EmailConverter{},
		},
	}

	data := map[string]any{
		"email": "invalid-email",
	}

	var user TestUser
	err := Bind(&user, data, opts)
	assert.Error(t, err)
	var bindingErr *BindingError
	assert.True(t, errors.As(err, &bindingErr))
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestConverterUnbind(t *testing.T) {
	// setup converters
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &EmailConverter{},
			reflect.TypeOf(UserID(0)): &UserIDConverter{},
		},
	}

	// create user
	user := TestUser{
		ID:    UserID(456),
		Email: Email("jane@example.com"),
		Name:  "jane smith",
	}

	// unbind with converters
	data, err := Unbind(user, opts)
	assert.NoError(t, err)

	// verify results
	expected := map[string]any{
		"user_id": 456,
		"email":   "jane@example.com",
		"name":    "jane smith",
	}
	assert.Equal(t, expected, data)
}

func TestConverterRoundTrip(t *testing.T) {
	// setup converters
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &EmailConverter{},
			reflect.TypeOf(UserID(0)): &UserIDConverter{},
		},
	}

	// original user
	original := TestUser{
		ID:    UserID(789),
		Email: Email("roundtrip@example.com"),
		Name:  "round trip",
	}

	// unbind to map
	data, err := Unbind(original, opts)
	assert.NoError(t, err)

	// bind back to struct
	var recovered TestUser
	err = Bind(&recovered, data, opts)
	assert.NoError(t, err)

	// verify they match
	assert.Equal(t, original, recovered)
}

// ComplexType is a custom struct type that can't be converted without a converter
type ComplexType struct {
	Value string
}

type TestComplexUser struct {
	Complex ComplexType `dd:"complex"`
	Name    string      `dd:"name"`
}

func TestConverterWithoutOptions(t *testing.T) {
	// test that complex custom types fail without converters
	data := map[string]any{
		"complex": "some string", // this can't be bound to ComplexType without a converter
		"name":    "no converter",
	}

	var user TestComplexUser
	err := Bind(&user, data)
	// this should fail because we can't convert string to ComplexType struct
	assert.Error(t, err)
	// the error should be about expecting an object but getting a string
	assert.Contains(t, err.Error(), "expected object for struct")
}

// BadEmailConverter implements Converter but returns wrong type
type BadEmailConverter struct{}

func (c *BadEmailConverter) FromRaw(raw interface{}) (interface{}, error) {
	return "wrong type", nil // should return Email, not string
}

func (c *BadEmailConverter) ToRaw(value interface{}) (interface{}, error) {
	return string(value.(Email)), nil
}

func TestConverterIncompatibleReturn(t *testing.T) {
	// converter that returns wrong type
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &BadEmailConverter{},
		},
	}

	data := map[string]any{
		"email": "test@example.com",
	}

	var user TestUser
	err := Bind(&user, data, opts)
	assert.Error(t, err)
	var bindingErr *BindingError
	assert.True(t, errors.As(err, &bindingErr))
	assert.Contains(t, err.Error(), "expected dd.Email, got string")
}

// Test with pointer fields
type TestUserWithPtrEmail struct {
	Email *Email `dd:"email"`
	Name  string `dd:"name"`
}

func TestConverterWithPointers(t *testing.T) {
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &EmailConverter{},
		},
	}

	data := map[string]any{
		"email": "ptr@example.com",
		"name":  "pointer test",
	}

	var user TestUserWithPtrEmail
	err := Bind(&user, data, opts)
	assert.NoError(t, err)

	assert.NotNil(t, user.Email)
	assert.Equal(t, Email("ptr@example.com"), *user.Email)
	assert.Equal(t, "pointer test", user.Name)

	// test unbind
	data2, err := Unbind(user, opts)
	assert.NoError(t, err)
	assert.Equal(t, "ptr@example.com", data2["email"])
}

// Test with slices
type TestUserWithEmails struct {
	Emails []Email `dd:"emails"`
	Name   string  `dd:"name"`
}

func TestConverterWithSlices(t *testing.T) {
	opts := &Options{
		Converters: map[reflect.Type]Converter{
			reflect.TypeOf(Email("")): &EmailConverter{},
		},
	}

	data := map[string]any{
		"emails": []interface{}{"first@example.com", "second@example.com"},
		"name":   "slice test",
	}

	var user TestUserWithEmails
	err := Bind(&user, data, opts)
	assert.NoError(t, err)

	assert.Len(t, user.Emails, 2)
	assert.Equal(t, Email("first@example.com"), user.Emails[0])
	assert.Equal(t, Email("second@example.com"), user.Emails[1])

	// test unbind
	data2, err := Unbind(user, opts)
	assert.NoError(t, err)
	expected := []interface{}{"first@example.com", "second@example.com"}
	assert.Equal(t, expected, data2["emails"])
}
