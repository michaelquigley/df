package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df"
)

// User implements df.Identifiable to participate in pointer references
type User struct {
	ID   string
	Name string
	Age  int
}

func (u *User) GetId() string { return u.ID }

// Document has pointer references to Users
type Document struct {
	ID     string
	Title  string
	Author *df.Pointer[*User]
	Editor *df.Pointer[*User]
}

func (d *Document) GetId() string { return d.ID }

// Container holds all objects
type Container struct {
	Users     []*User
	Documents []*Document
}

func main() {
	fmt.Println("=== df.Pointer[T] object references example ===")
	fmt.Println("demonstrates type-safe object references that support:")
	fmt.Println("• cycles and complex relationships")
	fmt.Println("• type safety with generics")  
	fmt.Println("• two-phase bind-and-link process")
	fmt.Println("• automatic reference resolution")
	
	// data with object references using $ref (like JSON Schema references)
	data := map[string]any{
		"users": []any{
			map[string]any{
				"id":   "user1",
				"name": "Alice Johnson",
				"age":  28,
			},
			map[string]any{
				"id":   "user2",
				"name": "Bob Smith", 
				"age":  35,
			},
		},
		"documents": []any{
			map[string]any{
				"id":     "doc1",
				"title":  "Go Programming Guide",
				"author": map[string]any{"$ref": "user1"}, // reference to Alice
				"editor": map[string]any{"$ref": "user2"}, // reference to Bob
			},
			map[string]any{
				"id":     "doc2", 
				"title":  "Advanced Techniques",
				"author": map[string]any{"$ref": "user2"}, // reference to Bob
				// editor omitted (optional field)
			},
		},
	}
	
	fmt.Println("\n=== input data structure ===")
	fmt.Printf("users: 2 users with IDs user1, user2\n")
	fmt.Printf("documents: 2 documents with $ref pointers to users\n") 
	fmt.Printf("note: $ref creates typed references, not string lookups\n")

	// phase 1: df.Bind() loads data and stores $ref strings (doesn't resolve yet)
	fmt.Println("\n=== phase 1: df.Bind() - load data and $ref strings ===")
	var container Container
	if err := df.Bind(&container, data); err != nil {
		log.Fatal("bind failed:", err)
	}

	fmt.Printf("✓ loaded %d users and %d documents\n", len(container.Users), len(container.Documents))
	fmt.Printf("doc1 author ref: '%s' (resolved: %t)\n",
		container.Documents[0].Author.Ref,
		container.Documents[0].Author.IsResolved())
	fmt.Printf("pointers contain reference strings but don't point to actual objects yet\n")

	// phase 2: df.Link() resolves all pointer references to actual objects
	fmt.Println("\n=== phase 2: df.Link() - resolve $ref strings to actual objects ===")
	if err := df.Link(&container); err != nil {
		log.Fatal("link failed:", err)
	}

	doc1 := container.Documents[0]
	doc2 := container.Documents[1]

	// access resolved objects with type safety
	author1 := doc1.Author.Resolve() // *User
	editor1 := doc1.Editor.Resolve() // *User
	author2 := doc2.Author.Resolve() // *User

	fmt.Printf("✓ all references resolved successfully\n")
	fmt.Printf("doc1 author: %s (resolved: %t)\n", author1.Name, doc1.Author.IsResolved())

	fmt.Println("\n=== final object graph ===")
	fmt.Printf("'%s' by %s (age %d), edited by %s (age %d)\n", 
		doc1.Title, author1.Name, author1.Age, editor1.Name, editor1.Age)
	fmt.Printf("'%s' by %s (age %d)", doc2.Title, author2.Name, author2.Age)
	
	if doc2.Editor != nil && doc2.Editor.IsResolved() {
		fmt.Printf(", edited by %s\n", doc2.Editor.Resolve().Name)
	} else {
		fmt.Printf(", no editor\n")
	}

	// demonstrate type safety and object identity
	fmt.Println("\n=== object identity and type safety ===")
	if author2 == editor1 {
		fmt.Printf("✓ author2 and editor1 both reference the same *User object: %s\n", author2.Name)
		fmt.Printf("  df.Pointer[T] ensures object identity is preserved\n")
	}
	fmt.Printf("✓ all pointers are type-safe: *df.Pointer[*User] can only point to *User objects\n")
	fmt.Printf("✓ supports cycles, caching, and complex relationships\n")
}
