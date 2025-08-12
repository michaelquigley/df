package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df"
)

// User represents a simple user profile for demonstration
type User struct {
	Name    string `df:"name,required"`
	Email   string `df:"email"`
	Age     int    `df:"age"`
	Active  bool   `df:"active"`
	Profile *Profile
}

// Profile contains additional user profile information
type Profile struct {
	Bio     string `df:"bio"`
	Website string `df:"website"`
}

func main() {
	fmt.Println("df - simple data persistence example")
	fmt.Println("====================================")

	// example data that might come from JSON, API, or database
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

	fmt.Println("\n1. binding data to struct:")
	fmt.Printf("input data: %+v\n", userData)

	var user User
	if err := df.Bind(&user, userData); err != nil {
		log.Fatalf("bind failed: %v", err)
	}

	fmt.Printf("bound struct: %+v\n", user)
	fmt.Printf("profile: %+v\n", *user.Profile)

	fmt.Println("\n2. unbinding struct to data:")

	unboundData, err := df.Unbind(&user)
	if err != nil {
		log.Fatalf("unbind failed: %v", err)
	}

	fmt.Printf("unbound data: %+v\n", unboundData)

	fmt.Println("\n3. round-trip verification:")

	var user2 User
	if err := df.Bind(&user2, unboundData); err != nil {
		log.Fatalf("round-trip bind failed: %v", err)
	}

	fmt.Printf("round-trip struct: %+v\n", user2)
	fmt.Printf("round-trip profile: %+v\n", *user2.Profile)

	fmt.Println("\n4. error handling example:")

	invalidData := map[string]any{
		"email": "john@example.com",
		"age":   30,
		// Missing required "name" field
	}

	var user3 User
	if err := df.Bind(&user3, invalidData); err != nil {
		fmt.Printf("expected error: %v\n", err)
	}

	fmt.Println("\nexample completed successfully!")
}
