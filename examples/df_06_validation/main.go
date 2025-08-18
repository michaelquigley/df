package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/michaelquigley/df"
)

// ValidationError represents a custom validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors collects multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d validation errors: %s (and %d more)", len(e), e[0].Error(), len(e)-1)
}

// UserRegistration demonstrates comprehensive user input validation
type UserRegistration struct {
	Username    string    `df:",required"`
	Email       string    `df:",required"`
	Password    string    `df:",required,secret"`
	Age         int       `df:",required"`
	Country     string    `df:",required"`
	PhoneNumber string    `df:"phone"`
	Website     string
	BirthDate   time.Time `df:"birth_date"`
	Terms       bool      `df:"accept_terms,required"`
}

func (ur *UserRegistration) UnmarshalDf(data map[string]any) error {
	var errors ValidationErrors

	// basic field binding first
	type tempUser UserRegistration
	temp := (*tempUser)(ur)
	if err := df.Bind(temp, data); err != nil {
		// if basic binding fails, return immediately
		return fmt.Errorf("basic binding failed: %v", err)
	}

	// custom validation after successful binding
	errors = append(errors, ur.validateUsername()...)
	errors = append(errors, ur.validateEmail()...)
	errors = append(errors, ur.validatePassword()...)
	errors = append(errors, ur.validateAge()...)
	errors = append(errors, ur.validateCountry()...)
	errors = append(errors, ur.validatePhoneNumber()...)
	errors = append(errors, ur.validateWebsite()...)
	errors = append(errors, ur.validateBirthDate()...)
	errors = append(errors, ur.validateTerms()...)

	// cross-field validation
	errors = append(errors, ur.validateAgeConsistency()...)

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (ur *UserRegistration) validateUsername() ValidationErrors {
	var errors ValidationErrors

	if len(ur.Username) < 3 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Value:   ur.Username,
			Message: "must be at least 3 characters long",
		})
	}

	if len(ur.Username) > 20 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Value:   ur.Username,
			Message: "must be no more than 20 characters long",
		})
	}

	// username must be alphanumeric with underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(ur.Username) {
		errors = append(errors, ValidationError{
			Field:   "username",
			Value:   ur.Username,
			Message: "must contain only letters, numbers, and underscores",
		})
	}

	// username cannot start with underscore
	if strings.HasPrefix(ur.Username, "_") {
		errors = append(errors, ValidationError{
			Field:   "username",
			Value:   ur.Username,
			Message: "cannot start with underscore",
		})
	}

	return errors
}

func (ur *UserRegistration) validateEmail() ValidationErrors {
	var errors ValidationErrors

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(ur.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Value:   ur.Email,
			Message: "must be a valid email address",
		})
	}

	// check for common email providers (business rule example)
	allowedDomains := []string{"gmail.com", "yahoo.com", "outlook.com", "example.com"}
	domain := strings.Split(ur.Email, "@")
	if len(domain) == 2 {
		found := false
		for _, allowed := range allowedDomains {
			if domain[1] == allowed {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, ValidationError{
				Field:   "email",
				Value:   ur.Email,
				Message: fmt.Sprintf("domain '%s' is not in the allowed list", domain[1]),
			})
		}
	}

	return errors
}

func (ur *UserRegistration) validatePassword() ValidationErrors {
	var errors ValidationErrors

	if len(ur.Password) < 8 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Value:   "[HIDDEN]",
			Message: "must be at least 8 characters long",
		})
	}

	// password strength requirements
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(ur.Password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(ur.Password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(ur.Password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(ur.Password)

	if !hasUpper {
		errors = append(errors, ValidationError{
			Field:   "password",
			Value:   "[HIDDEN]",
			Message: "must contain at least one uppercase letter",
		})
	}

	if !hasLower {
		errors = append(errors, ValidationError{
			Field:   "password",
			Value:   "[HIDDEN]",
			Message: "must contain at least one lowercase letter",
		})
	}

	if !hasDigit {
		errors = append(errors, ValidationError{
			Field:   "password",
			Value:   "[HIDDEN]",
			Message: "must contain at least one digit",
		})
	}

	if !hasSpecial {
		errors = append(errors, ValidationError{
			Field:   "password",
			Value:   "[HIDDEN]",
			Message: "must contain at least one special character",
		})
	}

	return errors
}

func (ur *UserRegistration) validateAge() ValidationErrors {
	var errors ValidationErrors

	if ur.Age < 13 {
		errors = append(errors, ValidationError{
			Field:   "age",
			Value:   ur.Age,
			Message: "must be at least 13 years old",
		})
	}

	if ur.Age > 120 {
		errors = append(errors, ValidationError{
			Field:   "age",
			Value:   ur.Age,
			Message: "must be a realistic age (120 or younger)",
		})
	}

	return errors
}

func (ur *UserRegistration) validateCountry() ValidationErrors {
	var errors ValidationErrors

	// ISO country code validation
	validCountries := map[string]bool{
		"US": true, "CA": true, "GB": true, "DE": true, "FR": true,
		"JP": true, "AU": true, "IN": true, "BR": true, "MX": true,
	}

	if !validCountries[ur.Country] {
		errors = append(errors, ValidationError{
			Field:   "country",
			Value:   ur.Country,
			Message: "must be a valid ISO country code",
		})
	}

	return errors
}

func (ur *UserRegistration) validatePhoneNumber() ValidationErrors {
	var errors ValidationErrors

	if ur.PhoneNumber == "" {
		return errors // phone is optional
	}

	// remove all non-digits
	digits := regexp.MustCompile(`\D`).ReplaceAllString(ur.PhoneNumber, "")

	if len(digits) < 10 {
		errors = append(errors, ValidationError{
			Field:   "phone",
			Value:   ur.PhoneNumber,
			Message: "must contain at least 10 digits",
		})
	}

	if len(digits) > 15 {
		errors = append(errors, ValidationError{
			Field:   "phone",
			Value:   ur.PhoneNumber,
			Message: "must contain no more than 15 digits",
		})
	}

	return errors
}

func (ur *UserRegistration) validateWebsite() ValidationErrors {
	var errors ValidationErrors

	if ur.Website == "" {
		return errors // website is optional
	}

	// basic URL validation
	if !strings.HasPrefix(ur.Website, "http://") && !strings.HasPrefix(ur.Website, "https://") {
		errors = append(errors, ValidationError{
			Field:   "website",
			Value:   ur.Website,
			Message: "must be a valid URL starting with http:// or https://",
		})
	}

	return errors
}

func (ur *UserRegistration) validateBirthDate() ValidationErrors {
	var errors ValidationErrors

	if ur.BirthDate.IsZero() {
		return errors // birth_date is optional
	}

	now := time.Now()
	if ur.BirthDate.After(now) {
		errors = append(errors, ValidationError{
			Field:   "birth_date",
			Value:   ur.BirthDate,
			Message: "cannot be in the future",
		})
	}

	// check if birth date is too far in the past
	oldestDate := now.AddDate(-120, 0, 0)
	if ur.BirthDate.Before(oldestDate) {
		errors = append(errors, ValidationError{
			Field:   "birth_date",
			Value:   ur.BirthDate,
			Message: "cannot be more than 120 years ago",
		})
	}

	return errors
}

func (ur *UserRegistration) validateTerms() ValidationErrors {
	var errors ValidationErrors

	if !ur.Terms {
		errors = append(errors, ValidationError{
			Field:   "accept_terms",
			Value:   ur.Terms,
			Message: "must accept terms and conditions",
		})
	}

	return errors
}

func (ur *UserRegistration) validateAgeConsistency() ValidationErrors {
	var errors ValidationErrors

	if !ur.BirthDate.IsZero() {
		// calculate age from birth date
		now := time.Now()
		calculatedAge := now.Year() - ur.BirthDate.Year()
		if now.YearDay() < ur.BirthDate.YearDay() {
			calculatedAge--
		}

		// allow 1 year difference for birthday timing
		if abs(calculatedAge-ur.Age) > 1 {
			errors = append(errors, ValidationError{
				Field:   "age",
				Value:   ur.Age,
				Message: fmt.Sprintf("age (%d) does not match birth date (calculated: %d)", ur.Age, calculatedAge),
			})
		}
	}

	return errors
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ConfigurationFile demonstrates file error handling
type ConfigurationFile struct {
	DatabaseURL string `df:"database_url,required"`
	Port        int    `df:",required"`
	Debug       bool
	Features    []string
}

// DataRecord demonstrates bulk import error handling
type DataRecord struct {
	ID    string `df:",required"`
	Name  string `df:",required"`
	Value string
}

func (dr *DataRecord) UnmarshalDf(data map[string]any) error {
	// attempt to fix common data issues
	if idValue, exists := data["id"]; exists {
		switch v := idValue.(type) {
		case int:
			data["id"] = strconv.Itoa(v)
		case float64:
			data["id"] = strconv.Itoa(int(v))
		}
	}

	// basic binding
	type tempRecord DataRecord
	temp := (*tempRecord)(dr)
	if err := df.Bind(temp, data); err != nil {
		return err
	}

	// validation
	if strings.TrimSpace(dr.Name) == "" {
		return ValidationError{
			Field:   "name",
			Value:   dr.Name,
			Message: "cannot be empty or whitespace only",
		}
	}

	// normalize whitespace
	dr.Name = strings.TrimSpace(dr.Name)

	return nil
}

func main() {
	fmt.Println("=== df error handling and validation example ===")
	fmt.Println("demonstrates comprehensive error handling patterns and validation strategies")
	fmt.Println("for production-ready applications")

	// step 1: demonstrate basic validation errors
	fmt.Println("\n=== step 1: basic validation error scenarios ===")

	basicTests := []map[string]any{
		{
			"username": "user@name", // invalid characters
			"email":    "invalid-email",
			"password": "weak",
			"age":      12,   // too young
			"country":  "XX", // invalid country
		},
		{
			// missing required fields
			"username": "validuser",
		},
		{
			"username":     "validuser",
			"email":        "user@gmail.com",
			"password":     "ValidPass123!",
			"age":          25,
			"country":      "US",
			"accept_terms": false, // must be true
		},
	}

	for i, testData := range basicTests {
		fmt.Printf("\nvalidation test %d:\n", i+1)
		var user UserRegistration
		if err := user.UnmarshalDf(testData); err != nil {
			if validationErrs, ok := err.(ValidationErrors); ok {
				fmt.Printf("  validation failed with %d errors:\n", len(validationErrs))
				for j, validationErr := range validationErrs {
					fmt.Printf("    %d. %s\n", j+1, validationErr.Error())
				}
			} else {
				fmt.Printf("  binding failed: %v\n", err)
			}
		} else {
			fmt.Printf("  ✓ validation passed\n")
		}
	}

	// step 2: demonstrate successful validation with edge cases
	fmt.Println("\n=== step 2: successful validation with edge cases ===")
	validData := map[string]any{
		"username":     "valid_user123",
		"email":        "user@gmail.com",
		"password":     "SecurePass123!",
		"age":          25,
		"country":      "US",
		"phone":        "(555) 123-4567",
		"website":      "https://example.com",
		"birth_date":   "1998-06-15T00:00:00Z",
		"accept_terms": true,
	}

	var validUser UserRegistration
	if err := validUser.UnmarshalDf(validData); err != nil {
		fmt.Printf("unexpected validation failure: %v\n", err)
	} else {
		fmt.Printf("✓ comprehensive validation passed\n")
		fmt.Printf("  user: %s (%s)\n", validUser.Username, validUser.Email)
		fmt.Printf("  age: %d, country: %s\n", validUser.Age, validUser.Country)
		if validUser.PhoneNumber != "" {
			fmt.Printf("  phone: %s\n", validUser.PhoneNumber)
		}
		if validUser.Website != "" {
			fmt.Printf("  website: %s\n", validUser.Website)
		}
	}

	// step 3: demonstrate file i/o error handling
	fmt.Println("\n=== step 3: file i/o error handling ===")

	// test with invalid JSON file
	fmt.Printf("attempting to parse invalid JSON...\n")

	// simulate file binding error
	invalidConfig := map[string]any{
		"database_url": "postgres://localhost/mydb",
		"port":         "8080", // string instead of int
	}

	var config ConfigurationFile
	if err := df.Bind(&config, invalidConfig); err != nil {
		fmt.Printf("✓ expected type conversion error: %v\n", err)
	}

	// test with missing required fields
	incompleteConfig := map[string]any{
		"debug": true,
		// missing database_url and port
	}

	var incompleteConfigFile ConfigurationFile
	if err := df.Bind(&incompleteConfigFile, incompleteConfig); err != nil {
		fmt.Printf("✓ expected required field error: %v\n", err)
	}

	// step 4: demonstrate bulk data processing with error accumulation
	fmt.Println("\n=== step 4: bulk data processing with error handling ===")

	bulkData := []map[string]any{
		{"id": "1", "name": "Valid Record", "value": "data1"},
		{"id": 2, "name": "ID Type Conversion", "value": "data2"},    // int id, should be fixed
		{"name": "Missing ID", "value": "data3"},                     // missing required id
		{"id": "4", "name": "", "value": "data4"},                    // empty name
		{"id": "5", "name": "  Whitespace Name  ", "value": "data5"}, // whitespace name, should be fixed
		{"id": "6", "name": "Valid Record 2", "value": "data6"},
	}

	fmt.Printf("processing %d records...\n", len(bulkData))

	var successCount, errorCount int
	var processingErrors []error

	for i, recordData := range bulkData {
		var record DataRecord
		if err := record.UnmarshalDf(recordData); err != nil {
			errorCount++
			processingErrors = append(processingErrors, fmt.Errorf("record %d: %v", i+1, err))
		} else {
			successCount++
			fmt.Printf("  ✓ record %d: %s (name: '%s')\n", i+1, record.ID, record.Name)
		}
	}

	fmt.Printf("\nbulk processing results:\n")
	fmt.Printf("  successful: %d\n", successCount)
	fmt.Printf("  errors: %d\n", errorCount)

	if len(processingErrors) > 0 {
		fmt.Printf("  error details:\n")
		for _, err := range processingErrors {
			fmt.Printf("    - %v\n", err)
		}
	}

	// step 5: demonstrate production error patterns
	fmt.Println("\n=== step 5: production error handling patterns ===")

	// user-friendly error messages
	fmt.Printf("user-friendly error conversion:\n")
	testUserData := map[string]any{
		"username": "a",
		"email":    "invalid",
		"password": "weak",
		"age":      5,
		"country":  "INVALID",
	}

	var testUser UserRegistration
	if err := testUser.UnmarshalDf(testUserData); err != nil {
		if validationErrs, ok := err.(ValidationErrors); ok {
			fmt.Printf("  technical errors (%d):\n", len(validationErrs))
			for _, validationErr := range validationErrs {
				fmt.Printf("    %s\n", validationErr.Error())
			}

			fmt.Printf("\n  user-friendly messages:\n")
			for _, validationErr := range validationErrs {
				userMessage := convertToUserMessage(validationErr)
				fmt.Printf("    %s\n", userMessage)
			}
		}
	}

	fmt.Println("\n=== error handling best practices demonstrated ===")
	fmt.Println("✓ comprehensive field validation with custom logic")
	fmt.Println("✓ error accumulation for better user experience")
	fmt.Println("✓ cross-field validation for data consistency")
	fmt.Println("✓ automatic data cleaning and normalization")
	fmt.Println("✓ structured error information for debugging")
	fmt.Println("✓ user-friendly error message conversion")
	fmt.Println("✓ bulk processing with error reporting")
	fmt.Println("✓ production-ready error handling patterns")

	fmt.Println("\n=== error handling and validation example completed successfully! ===")
}

func convertToUserMessage(err ValidationError) string {
	switch err.Field {
	case "username":
		if strings.Contains(err.Message, "3 characters") {
			return "Username is too short. Please choose a username with at least 3 characters."
		}
		if strings.Contains(err.Message, "20 characters") {
			return "Username is too long. Please choose a username with 20 characters or fewer."
		}
		if strings.Contains(err.Message, "letters, numbers") {
			return "Username can only contain letters, numbers, and underscores."
		}
		if strings.Contains(err.Message, "underscore") {
			return "Username cannot start with an underscore."
		}
	case "email":
		if strings.Contains(err.Message, "valid email") {
			return "Please enter a valid email address."
		}
		if strings.Contains(err.Message, "allowed list") {
			return "Please use an email from an approved domain."
		}
	case "password":
		if strings.Contains(err.Message, "8 characters") {
			return "Password must be at least 8 characters long."
		}
		if strings.Contains(err.Message, "uppercase") {
			return "Password must include at least one uppercase letter."
		}
		if strings.Contains(err.Message, "lowercase") {
			return "Password must include at least one lowercase letter."
		}
		if strings.Contains(err.Message, "digit") {
			return "Password must include at least one number."
		}
		if strings.Contains(err.Message, "special") {
			return "Password must include at least one special character (!@#$%^&*...)."
		}
	case "age":
		if strings.Contains(err.Message, "13 years") {
			return "You must be at least 13 years old to register."
		}
		if strings.Contains(err.Message, "realistic") {
			return "Please enter a valid age."
		}
		if strings.Contains(err.Message, "does not match") {
			return "The age you entered doesn't match your birth date."
		}
	case "country":
		return "Please select a valid country from the list."
	case "accept_terms":
		return "You must accept the terms and conditions to register."
	}

	return fmt.Sprintf("Please check the %s field: %s", err.Field, err.Message)
}
