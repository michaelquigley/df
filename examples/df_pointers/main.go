package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df"
)

// User implements df.Identifiable to participate in pointer references
type User struct {
	ID   string `df:"id"`
	Name string `df:"name"`
	Age  int    `df:"age"`
}

func (u *User) GetId() string { return u.ID }

// Document has pointer references to Users
type Document struct {
	ID     string             `df:"id"`
	Title  string             `df:"title"`
	Author *df.Pointer[*User] `df:"author"`
	Editor *df.Pointer[*User] `df:"editor,omitempty"`
}

func (d *Document) GetId() string { return d.ID }

// Container holds all objects
type Container struct {
	Users     []*User     `df:"users"`
	Documents []*Document `df:"documents"`
}

func main() {
	// data with pointer references using $ref
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
				// editor omitted
			},
		},
	}

	// phase 1: Bind data (loads $ref strings, doesn't resolve yet)
	var container Container
	if err := df.Bind(&container, data); err != nil {
		log.Fatal("bind failed:", err)
	}

	fmt.Println("=== after binding (before linking) ===")
	fmt.Printf("doc1 author ref: %s (resolved: %t)\n",
		container.Documents[0].Author.Ref,
		container.Documents[0].Author.IsResolved())

	// phase 2: Link resolves all pointer references
	if err := df.Link(&container); err != nil {
		log.Fatal("link failed:", err)
	}

	fmt.Println("\n=== after linking ===")
	doc1 := container.Documents[0]
	doc2 := container.Documents[1]

	// access resolved objects
	author1 := doc1.Author.Resolve()
	editor1 := doc1.Editor.Resolve()
	author2 := doc2.Author.Resolve()

	fmt.Printf("doc1: %s\n", doc1.Title)
	fmt.Printf("  author: %s (age %d)\n", author1.Name, author1.Age)
	fmt.Printf("  editor: %s (age %d)\n", editor1.Name, editor1.Age)

	fmt.Printf("doc2: %s\n", doc2.Title)
	fmt.Printf("  author: %s (age %d)\n", author2.Name, author2.Age)
	if doc2.Editor != nil && doc2.Editor.IsResolved() {
		fmt.Printf("  editor: %s\n", doc2.Editor.Resolve().Name)
	} else {
		fmt.Printf("  editor: none\n")
	}

	// demonstrate type safety - both documents reference the same Bob
	if author2 == editor1 {
		fmt.Printf("\n✓ both references to user2 point to the same object: %s\n", author2.Name)
	}
}
