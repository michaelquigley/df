package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df/dd"
)

// User represents a simple user profile for demonstration
type User struct {
	Name    string `df:"+required"`
	Email   string
	Age     int
	Active  bool
	Profile *Profile
}

// Profile contains additional user profile information
type Profile struct {
	Bio     string
	Website string
}

func main() {
	fmt.Println("df - basic bind/unbind example")
	fmt.Println("==============================")
	fmt.Println("demonstrates the foundation of df: converting between maps and structs")
	fmt.Println("using both modern New[T] and traditional Bind() approaches")

	// example data that might come from JSON, API, database, or configuration files
	userData := map[string]any{
		"name":   "John Doe",
		"email":  "john@example.com",
		"age":    30,
		"active": true,
		"profile": map[string]any{
			"bio":     "Software developer",
			"website": "https://johndoe.dev",
		},
	}

	fmt.Println("\n1. binding data to struct (using New[T] API):")
	fmt.Printf("input data: %+v\n", userData)
	fmt.Printf("struct tags define field mapping and requirements\n")

	user, err := dd.New[User](userData)
	if err != nil {
		log.Fatalf("bind failed: %v", err)
	}

	fmt.Printf("✓ bound struct: %+v\n", *user)
	fmt.Printf("  nested profile: %+v\n", *user.Profile)
	fmt.Printf("  note: 'profile' field automatically snake_cased to match nested struct\n")

	fmt.Println("\n2. unbinding struct back to map (for persistence/serialization):")

	unboundData, err := dd.Unbind(user)
	if err != nil {
		log.Fatalf("unbind failed: %v", err)
	}

	fmt.Printf("✓ unbound data: %+v\n", unboundData)
	fmt.Printf("  ready for JSON/YAML serialization or database storage\n")

	fmt.Println("\n3. round-trip verification (using New[T]):")

	user2, err := dd.New[User](unboundData)
	if err != nil {
		log.Fatalf("round-trip bind failed: %v", err)
	}

	fmt.Printf("round-trip struct: %+v\n", *user2)
	fmt.Printf("round-trip profile: %+v\n", *user2.Profile)

	fmt.Println("\n4. error handling example:")

	invalidData := map[string]any{
		"email": "john@example.com",
		"age":   30,
		// missing required "name" field
	}

	_, err = dd.New[User](invalidData)
	if err != nil {
		fmt.Printf("expected error: %v\n", err)
	}

	fmt.Println("\n5. Bind approach (still available):")

	var user4 User
	if err := dd.Bind(&user4, userData); err != nil {
		log.Fatalf("bind failed: %v", err)
	}
	fmt.Printf("bind result: %+v\n", user4)

	fmt.Println("\nexample completed successfully!")
}
