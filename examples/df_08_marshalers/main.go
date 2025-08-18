package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/michaelquigley/df"
)

// CustomTime handles multiple time formats and timezone normalization
type CustomTime struct {
	time.Time
	OriginalFormat string
}

func (ct *CustomTime) UnmarshalDf(data map[string]any) error {
	timeValue, ok := data["time"]
	if !ok {
		return fmt.Errorf("missing required 'time' field")
	}

	timeStr, ok := timeValue.(string)
	if !ok {
		return fmt.Errorf("time field must be a string, got %T", timeValue)
	}

	// try multiple time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"01-02-2006 15:04",
	}

	var parsedTime time.Time
	var err error
	var detectedFormat string

	for _, format := range formats {
		if parsedTime, err = time.Parse(format, timeStr); err == nil {
			detectedFormat = format
			break
		}
	}

	if err != nil {
		return fmt.Errorf("unable to parse time '%s' with any supported format", timeStr)
	}

	// normalize to UTC
	ct.Time = parsedTime.UTC()
	ct.OriginalFormat = detectedFormat

	return nil
}

func (ct CustomTime) MarshalDf() (map[string]any, error) {
	return map[string]any{
		"time":            ct.Time.Format(time.RFC3339),
		"original_format": ct.OriginalFormat,
		"timestamp":       ct.Time.Unix(),
	}, nil
}

// ContactInfo validates and normalizes contact information
type ContactInfo struct {
	Email string
	Phone string
	Name  string
}

func (ci *ContactInfo) UnmarshalDf(data map[string]any) error {
	// extract and validate email
	if emailValue, exists := data["email"]; exists {
		email, ok := emailValue.(string)
		if !ok {
			return fmt.Errorf("email must be a string, got %T", emailValue)
		}

		// validate email format
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		if matched, _ := regexp.MatchString(emailRegex, email); !matched {
			return fmt.Errorf("invalid email format: %s", email)
		}
		ci.Email = strings.ToLower(email) // normalize to lowercase
	}

	// extract and normalize phone
	if phoneValue, exists := data["phone"]; exists {
		phone, ok := phoneValue.(string)
		if !ok {
			return fmt.Errorf("phone must be a string, got %T", phoneValue)
		}

		// normalize phone (remove non-digits)
		phoneRegex := regexp.MustCompile(`[^\d]`)
		normalizedPhone := phoneRegex.ReplaceAllString(phone, "")

		if len(normalizedPhone) < 10 {
			return fmt.Errorf("phone number too short: %s", phone)
		}
		ci.Phone = normalizedPhone
	}

	// extract and validate name
	if nameValue, exists := data["name"]; exists {
		name, ok := nameValue.(string)
		if !ok {
			return fmt.Errorf("name must be a string, got %T", nameValue)
		}

		name = strings.TrimSpace(name)
		if len(name) == 0 {
			return fmt.Errorf("name cannot be empty")
		}
		ci.Name = name
	}

	// validate that at least email or phone is provided
	if ci.Email == "" && ci.Phone == "" {
		return fmt.Errorf("either email or phone must be provided")
	}

	return nil
}

func (ci ContactInfo) MarshalDf() (map[string]any, error) {
	result := map[string]any{
		"name": ci.Name,
	}

	if ci.Email != "" {
		result["email"] = ci.Email
		result["email_domain"] = strings.Split(ci.Email, "@")[1]
	}

	if ci.Phone != "" {
		result["phone"] = ci.Phone
		// format phone for display
		if len(ci.Phone) == 10 {
			result["phone_formatted"] = fmt.Sprintf("(%s) %s-%s", 
				ci.Phone[:3], ci.Phone[3:6], ci.Phone[6:])
		}
	}

	return result, nil
}

// LegacyUser handles legacy user data format conversion
type LegacyUser struct {
	ID       int
	FullName string
	Contact  ContactInfo
	Settings map[string]string
}

func (lu *LegacyUser) UnmarshalDf(data map[string]any) error {
	// handle legacy ID formats
	if idValue, exists := data["id"]; exists {
		switch v := idValue.(type) {
		case int:
			lu.ID = v
		case float64:
			lu.ID = int(v)
		case string:
			// legacy string IDs need to be converted
			if v == "" {
				return fmt.Errorf("user ID cannot be empty")
			}
			// for demo, just use length as ID
			lu.ID = len(v)
		default:
			return fmt.Errorf("unsupported ID type: %T", idValue)
		}
	}

	// handle legacy name formats (could be first_name + last_name)
	if fullName, exists := data["full_name"]; exists {
		if name, ok := fullName.(string); ok {
			lu.FullName = name
		}
	} else if firstName, hasFirst := data["first_name"]; hasFirst {
		if lastName, hasLast := data["last_name"]; hasLast {
			first, _ := firstName.(string)
			last, _ := lastName.(string)
			lu.FullName = strings.TrimSpace(first + " " + last)
		}
	}

	// handle legacy contact format (flattened vs nested)
	contactData := make(map[string]any)
	if contact, exists := data["contact"]; exists {
		// nested format
		if contactMap, ok := contact.(map[string]any); ok {
			contactData = contactMap
		}
	} else {
		// flattened format
		if email, exists := data["email"]; exists {
			contactData["email"] = email
		}
		if phone, exists := data["phone"]; exists {
			contactData["phone"] = phone
		}
		contactData["name"] = lu.FullName
	}

	// bind contact using its unmarshaler
	if len(contactData) > 0 {
		if err := lu.Contact.UnmarshalDf(contactData); err != nil {
			return fmt.Errorf("failed to bind contact: %v", err)
		}
	}

	// handle settings (could be various formats)
	lu.Settings = make(map[string]string)
	if settingsValue, exists := data["settings"]; exists {
		switch settings := settingsValue.(type) {
		case map[string]any:
			for k, v := range settings {
				if strVal, ok := v.(string); ok {
					lu.Settings[k] = strVal
				} else {
					lu.Settings[k] = fmt.Sprintf("%v", v)
				}
			}
		case map[string]string:
			lu.Settings = settings
		}
	}

	return nil
}

func (lu LegacyUser) MarshalDf() (map[string]any, error) {
	// marshal contact using its marshaler
	contactData, err := lu.Contact.MarshalDf()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contact: %v", err)
	}

	result := map[string]any{
		"id":        lu.ID,
		"full_name": lu.FullName,
		"contact":   contactData,
		"settings":  lu.Settings,
	}

	// add computed fields
	result["user_type"] = "legacy"
	result["has_email"] = lu.Contact.Email != ""
	result["has_phone"] = lu.Contact.Phone != ""

	return result, nil
}

// Container for demonstration
type UserProfile struct {
	CreatedAt    CustomTime
	User         LegacyUser
	LastModified CustomTime
}

func main() {
	fmt.Println("=== df custom marshaler/unmarshaler example ===")
	fmt.Println("demonstrates complete control over binding/unbinding process")
	fmt.Println("using Marshaler and Unmarshaler interfaces")

	// step 1: demonstrate CustomTime with multiple formats
	fmt.Println("\n=== step 1: custom time handling ===")
	timeData := []map[string]any{
		{"time": "2023-12-01T10:30:00Z"},        // rfc3339
		{"time": "2023-12-02 15:45:30"},         // custom format
		{"time": "12/03/2023"},                  // US date format
		{"time": "12-04-2023 09:15"},            // another custom format
	}

	for i, data := range timeData {
		var ct CustomTime
		if err := ct.UnmarshalDf(data); err != nil {
			fmt.Printf("failed to parse time %d: %v\n", i+1, err)
			continue
		}
		fmt.Printf("time %d: %s (format: %s)\n", i+1, ct.Time.Format("2006-01-02 15:04:05"), ct.OriginalFormat)
	}

	// step 2: demonstrate contact validation and normalization
	fmt.Println("\n=== step 2: contact validation and normalization ===")
	contactTests := []map[string]any{
		{
			"name":  "John Doe",
			"email": "JOHN.DOE@Example.Com",
			"phone": "(555) 123-4567",
		},
		{
			"name":  "Jane Smith",
			"email": "jane@test.com",
			"phone": "555.987.6543",
		},
		{
			"name":  "Invalid User",
			"email": "invalid-email",
			"phone": "123", // too short
		},
	}

	for i, data := range contactTests {
		var contact ContactInfo
		fmt.Printf("\ncontact test %d:\n", i+1)
		fmt.Printf("  input: %+v\n", data)
		
		if err := contact.UnmarshalDf(data); err != nil {
			fmt.Printf("  error: %v\n", err)
			continue
		}
		
		fmt.Printf("  normalized: name=%s, email=%s, phone=%s\n", 
			contact.Name, contact.Email, contact.Phone)

		// demonstrate marshaling
		marshaled, err := contact.MarshalDf()
		if err != nil {
			fmt.Printf("  marshal error: %v\n", err)
			continue
		}
		fmt.Printf("  marshaled: %+v\n", marshaled)
	}

	// step 3: demonstrate legacy user format handling
	fmt.Println("\n=== step 3: legacy user format handling ===")
	legacyData := []map[string]any{
		{
			"id":         "user123",
			"first_name": "Alice",
			"last_name":  "Johnson",
			"email":      "alice@example.com",
			"phone":      "555-111-2222",
			"settings": map[string]any{
				"theme":    "dark",
				"language": "en",
				"notifications": true,
			},
		},
		{
			"id":        42,
			"full_name": "Bob Wilson",
			"contact": map[string]any{
				"name":  "Bob Wilson",
				"email": "bob@test.com",
				"phone": "555-333-4444",
			},
			"settings": map[string]string{
				"theme": "light",
				"timezone": "UTC",
			},
		},
	}

	for i, data := range legacyData {
		var user LegacyUser
		fmt.Printf("\nlegacy user %d:\n", i+1)
		fmt.Printf("  input format: %+v\n", data)
		
		if err := user.UnmarshalDf(data); err != nil {
			fmt.Printf("  error: %v\n", err)
			continue
		}
		
		fmt.Printf("  normalized: ID=%d, Name=%s\n", user.ID, user.FullName)
		fmt.Printf("  contact: %s (%s)\n", user.Contact.Name, user.Contact.Email)

		// demonstrate marshaling to modern format
		marshaled, err := user.MarshalDf()
		if err != nil {
			fmt.Printf("  marshal error: %v\n", err)
			continue
		}
		fmt.Printf("  modern format: %+v\n", marshaled)
	}

	// step 4: demonstrate complete round-trip with nested marshalers
	fmt.Println("\n=== step 4: complete round-trip with nested marshalers ===")
	profileData := map[string]any{
		"created_at": map[string]any{
			"time": "2023-01-15T08:30:00Z",
		},
		"user": map[string]any{
			"id":        100,
			"full_name": "Test User",
			"email":     "test@example.com",
			"phone":     "555-1234567",
			"settings": map[string]string{
				"theme": "auto",
			},
		},
		"last_modified": map[string]any{
			"time": "2023-12-01 14:22:15",
		},
	}

	var profile UserProfile
	if err := df.Bind(&profile, profileData); err != nil {
		log.Fatalf("failed to bind profile: %v", err)
	}

	fmt.Printf("profile bound successfully:\n")
	fmt.Printf("  created: %s\n", profile.CreatedAt.Time.Format("2006-01-02 15:04:05"))
	fmt.Printf("  user: %s (ID: %d)\n", profile.User.FullName, profile.User.ID)
	fmt.Printf("  modified: %s\n", profile.LastModified.Time.Format("2006-01-02 15:04:05"))

	// unbind to show the marshaled result
	unboundData, err := df.Unbind(profile)
	if err != nil {
		log.Fatalf("failed to unbind profile: %v", err)
	}

	fmt.Printf("\nunbound profile with all marshaler transformations:\n")
	if createdData, ok := unboundData["created_at"].(map[string]any); ok {
		fmt.Printf("  created_at: %+v\n", createdData)
	}
	if userData, ok := unboundData["user"].(map[string]any); ok {
		fmt.Printf("  user: %+v\n", userData)
	}

	fmt.Println("\n=== key differences from converters ===")
	fmt.Println("• marshalers handle the complete binding/unbinding process")
	fmt.Println("• converters only handle type conversion between specific types")
	fmt.Println("• marshalers can access the entire data map for validation")
	fmt.Println("• marshalers can add computed fields during marshaling")
	fmt.Println("• marshalers are ideal for legacy format support")

	fmt.Println("\n=== custom marshaler/unmarshaler example completed successfully! ===")
}