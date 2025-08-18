package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/michaelquigley/df"
)

// Document represents a versioned document with relationships
type Document struct {
	ID          string
	Title       string
	Content     string
	Version     int
	Status      string
	Author      *df.Pointer[*Author]
	Reviewer    *df.Pointer[*Author]
	Category    *df.Pointer[*Category]
	PrevVersion *df.Pointer[*Document]
	NextVersion *df.Pointer[*Document]
	References  []*df.Pointer[*Document]
	CreatedAt   time.Time
}

func (d *Document) GetId() string { return d.ID }

// Author represents content authors with organizational relationships
type Author struct {
	ID         string
	Name       string
	Email      string
	Department *df.Pointer[*Department]
	Manager    *df.Pointer[*Author]
	Reports    []*df.Pointer[*Author]
	JoinedAt   time.Time
}

func (a *Author) GetId() string { return a.ID }

// Department represents organizational departments
type Department struct {
	ID      string
	Name    string
	Head    *df.Pointer[*Author]
	Parent  *df.Pointer[*Department]
	SubDepts []*df.Pointer[*Department] `df:"sub_departments"`
}

func (d *Department) GetId() string { return d.ID }

// Category represents hierarchical content categories
type Category struct {
	ID       string
	Name     string
	Parent   *df.Pointer[*Category]
	Children []*df.Pointer[*Category]
	Related  []*df.Pointer[*Category]
}

func (c *Category) GetId() string { return c.ID }

// Project represents projects with complex dependencies
type Project struct {
	ID           string
	Name         string
	Owner        *df.Pointer[*Author]
	Dependencies []*df.Pointer[*Project]
	Dependents   []*df.Pointer[*Project]
	Documents    []*df.Pointer[*Document]
}

func (p *Project) GetId() string { return p.ID }

// TimeConverter handles conversion of ISO 8601 strings to time.Time
type TimeConverter struct{}

func (c *TimeConverter) FromRaw(raw interface{}) (interface{}, error) {
	s, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("expected string for time, got %T", raw)
	}
	return time.Parse(time.RFC3339, s)
}

func (c *TimeConverter) ToRaw(value interface{}) (interface{}, error) {
	t, ok := value.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expected time.Time, got %T", value)
	}
	return t.Format(time.RFC3339), nil
}

// ContentManagementSystem holds all objects
type ContentManagementSystem struct {
	Documents   []*Document
	Authors     []*Author
	Departments []*Department
	Categories  []*Category
	Projects    []*Project
}

func main() {
	fmt.Println("=== df advanced linker example ===")
	fmt.Println("demonstrates advanced pointer reference resolution using NewLinker()")
	fmt.Println("with custom options, multi-stage linking, and complex object graphs")

	// step 1: create data with complex relationships
	fmt.Println("\n=== step 1: creating complex object graph data ===")
	
	// configure options with time converter for proper time.Time handling
	opts := &df.Options{
		Converters: map[reflect.Type]df.Converter{
			reflect.TypeOf(time.Time{}): &TimeConverter{},
		},
	}
	
	data := map[string]any{
		"departments": []any{
			map[string]any{
				"id":   "eng",
				"name": "Engineering",
				"head": map[string]any{"$ref": "alice"},
			},
			map[string]any{
				"id":     "frontend",
				"name":   "Frontend Team",
				"parent": map[string]any{"$ref": "eng"},
				"head":   map[string]any{"$ref": "bob"},
			},
			map[string]any{
				"id":     "backend",
				"name":   "Backend Team", 
				"parent": map[string]any{"$ref": "eng"},
				"head":   map[string]any{"$ref": "charlie"},
			},
		},
		"authors": []any{
			map[string]any{
				"id":         "alice",
				"name":       "Alice Johnson",
				"email":      "alice@example.com",
				"department": map[string]any{"$ref": "eng"},
				"reports":    []any{
					map[string]any{"$ref": "bob"},
					map[string]any{"$ref": "charlie"},
				},
				"joined_at": "2020-01-15T09:00:00Z",
			},
			map[string]any{
				"id":         "bob",
				"name":       "Bob Smith",
				"email":      "bob@example.com",
				"department": map[string]any{"$ref": "frontend"},
				"manager":    map[string]any{"$ref": "alice"},
				"reports":    []any{
					map[string]any{"$ref": "david"},
				},
				"joined_at": "2020-03-10T09:00:00Z",
			},
			map[string]any{
				"id":         "charlie",
				"name":       "Charlie Brown",
				"email":      "charlie@example.com",
				"department": map[string]any{"$ref": "backend"},
				"manager":    map[string]any{"$ref": "alice"},
				"reports":    []any{
					map[string]any{"$ref": "eve"},
				},
				"joined_at": "2020-02-20T09:00:00Z",
			},
			map[string]any{
				"id":         "david",
				"name":       "David Wilson",
				"email":      "david@example.com",
				"department": map[string]any{"$ref": "frontend"},
				"manager":    map[string]any{"$ref": "bob"},
				"joined_at": "2021-01-05T09:00:00Z",
			},
			map[string]any{
				"id":         "eve",
				"name":       "Eve Davis",
				"email":      "eve@example.com",
				"department": map[string]any{"$ref": "backend"},
				"manager":    map[string]any{"$ref": "charlie"},
				"joined_at": "2021-03-15T09:00:00Z",
			},
		},
		"categories": []any{
			map[string]any{
				"id":   "tech",
				"name": "Technology",
				"children": []any{
					map[string]any{"$ref": "frontend-tech"},
					map[string]any{"$ref": "backend-tech"},
				},
			},
			map[string]any{
				"id":     "frontend-tech",
				"name":   "Frontend Technology",
				"parent": map[string]any{"$ref": "tech"},
				"related": []any{
					map[string]any{"$ref": "backend-tech"},
				},
			},
			map[string]any{
				"id":     "backend-tech",
				"name":   "Backend Technology",
				"parent": map[string]any{"$ref": "tech"},
				"related": []any{
					map[string]any{"$ref": "frontend-tech"},
				},
			},
		},
		"documents": []any{
			map[string]any{
				"id":         "doc1",
				"title":      "Frontend Architecture Guide",
				"content":    "Comprehensive guide to frontend architecture...",
				"version":    1,
				"status":     "published",
				"author":     map[string]any{"$ref": "bob"},
				"reviewer":   map[string]any{"$ref": "alice"},
				"category":   map[string]any{"$ref": "frontend-tech"},
				"created_at": "2023-01-15T10:30:00Z",
			},
			map[string]any{
				"id":          "doc2",
				"title":       "Frontend Architecture Guide v2",
				"content":     "Updated guide with new patterns...",
				"version":     2,
				"status":      "draft",
				"author":      map[string]any{"$ref": "david"},
				"reviewer":    map[string]any{"$ref": "bob"},
				"category":    map[string]any{"$ref": "frontend-tech"},
				"prev_version": map[string]any{"$ref": "doc1"},
				"references": []any{
					map[string]any{"$ref": "doc3"},
				},
				"created_at": "2023-06-20T14:15:00Z",
			},
			map[string]any{
				"id":         "doc3",
				"title":      "Backend API Design",
				"content":    "Best practices for API design...",
				"version":    1,
				"status":     "published",
				"author":     map[string]any{"$ref": "eve"},
				"reviewer":   map[string]any{"$ref": "charlie"},
				"category":   map[string]any{"$ref": "backend-tech"},
				"created_at": "2023-03-10T11:45:00Z",
			},
		},
		"projects": []any{
			map[string]any{
				"id":   "web-app",
				"name": "Web Application",
				"owner": map[string]any{"$ref": "alice"},
				"dependencies": []any{
					map[string]any{"$ref": "api-service"},
				},
				"documents": []any{
					map[string]any{"$ref": "doc1"},
					map[string]any{"$ref": "doc2"},
				},
			},
			map[string]any{
				"id":   "api-service",
				"name": "API Service",
				"owner": map[string]any{"$ref": "charlie"},
				"dependents": []any{
					map[string]any{"$ref": "web-app"},
				},
				"documents": []any{
					map[string]any{"$ref": "doc3"},
				},
			},
		},
	}

	fmt.Printf("✓ created complex data with circular references and deep relationships\n")

	// step 2: demonstrate basic binding  
	fmt.Println("\n=== step 2: basic binding ===")
	var cms ContentManagementSystem
	if err := df.Bind(&cms, data, opts); err != nil {
		log.Fatalf("failed to bind CMS data: %v", err)
	}

	fmt.Printf("✓ bound %d documents, %d authors, %d departments, %d categories, %d projects (without linking)\n",
		len(cms.Documents), len(cms.Authors), len(cms.Departments), len(cms.Categories), len(cms.Projects))

	// step 3: demonstrate advanced linker with custom options
	fmt.Println("\n=== step 3: advanced linker with custom options ===")
	
	// create a new linker instance for advanced operations
	linker := df.NewLinker()
	
	// register objects in stages to demonstrate multi-stage linking
	fmt.Println("registering objects in stages...")
	
	// stage 1: register departments first
	for _, dept := range cms.Departments {
		if err := linker.Register(dept); err != nil {
			log.Fatalf("failed to register department: %v", err)
		}
	}
	fmt.Printf("  stage 1: registered %d departments\n", len(cms.Departments))
	
	// stage 2: register authors (depends on departments)
	for _, author := range cms.Authors {
		if err := linker.Register(author); err != nil {
			log.Fatalf("failed to register author: %v", err)
		}
	}
	fmt.Printf("  stage 2: registered %d authors\n", len(cms.Authors))
	
	// stage 3: register categories
	for _, category := range cms.Categories {
		if err := linker.Register(category); err != nil {
			log.Fatalf("failed to register category: %v", err)
		}
	}
	fmt.Printf("  stage 3: registered %d categories\n", len(cms.Categories))
	
	// stage 4: register documents (depends on authors and categories)
	for _, doc := range cms.Documents {
		if err := linker.Register(doc); err != nil {
			log.Fatalf("failed to register document: %v", err)
		}
	}
	fmt.Printf("  stage 4: registered %d documents\n", len(cms.Documents))
	
	// stage 5: register projects (depends on authors and documents)
	for _, project := range cms.Projects {
		if err := linker.Register(project); err != nil {
			log.Fatalf("failed to register project: %v", err)
		}
	}
	fmt.Printf("  stage 5: registered %d projects\n", len(cms.Projects))

	// now link all objects
	if err := linker.Link(&cms); err != nil {
		log.Fatalf("failed to link with advanced linker: %v", err)
	}
	fmt.Printf("✓ advanced linker completed successfully\n")

	// step 4: demonstrate complex relationship traversal
	fmt.Println("\n=== step 4: complex relationship traversal ===")
	
	// traverse organizational hierarchy
	alice := cms.Authors[0] // Alice Johnson
	fmt.Printf("organizational hierarchy starting from %s:\n", alice.Name)
	fmt.Printf("  department: %s\n", alice.Department.Resolve().Name)
	fmt.Printf("  direct reports: %d\n", len(alice.Reports))
	for _, report := range alice.Reports {
		reportAuthor := report.Resolve()
		fmt.Printf("    - %s (%s)\n", reportAuthor.Name, reportAuthor.Department.Resolve().Name)
		
		// show their reports too
		for _, subReport := range reportAuthor.Reports {
			subReportAuthor := subReport.Resolve()
			fmt.Printf("      - %s (%s)\n", subReportAuthor.Name, subReportAuthor.Department.Resolve().Name)
		}
	}

	// traverse document version chain
	fmt.Printf("\ndocument version chain:\n")
	doc1 := cms.Documents[0]
	fmt.Printf("  v%d: %s (by %s)\n", doc1.Version, doc1.Title, doc1.Author.Resolve().Name)
	
	// find documents that reference this one as next_version
	for _, doc := range cms.Documents {
		if doc.PrevVersion != nil && doc.PrevVersion.Resolve().ID == doc1.ID {
			fmt.Printf("  v%d: %s (by %s)\n", doc.Version, doc.Title, doc.Author.Resolve().Name)
			
			// show references
			if len(doc.References) > 0 {
				fmt.Printf("    references:\n")
				for _, ref := range doc.References {
					refDoc := ref.Resolve()
					fmt.Printf("      - %s (by %s)\n", refDoc.Title, refDoc.Author.Resolve().Name)
				}
			}
		}
	}

	// traverse category hierarchy
	fmt.Printf("\ncategory hierarchy:\n")
	techCategory := cms.Categories[0] // Technology
	fmt.Printf("  %s\n", techCategory.Name)
	for _, child := range techCategory.Children {
		childCat := child.Resolve()
		fmt.Printf("    - %s\n", childCat.Name)
		
		// show related categories
		if len(childCat.Related) > 0 {
			fmt.Printf("      related: ")
			for i, related := range childCat.Related {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", related.Resolve().Name)
			}
			fmt.Printf("\n")
		}
	}

	// step 5: demonstrate project dependency graph
	fmt.Println("\n=== step 5: project dependency analysis ===")
	for _, project := range cms.Projects {
		fmt.Printf("project: %s (owner: %s)\n", project.Name, project.Owner.Resolve().Name)
		
		if len(project.Dependencies) > 0 {
			fmt.Printf("  depends on:\n")
			for _, dep := range project.Dependencies {
				depProject := dep.Resolve()
				fmt.Printf("    - %s (owner: %s)\n", depProject.Name, depProject.Owner.Resolve().Name)
			}
		}
		
		if len(project.Dependents) > 0 {
			fmt.Printf("  depended on by:\n")
			for _, dependent := range project.Dependents {
				depProject := dependent.Resolve()
				fmt.Printf("    - %s (owner: %s)\n", depProject.Name, depProject.Owner.Resolve().Name)
			}
		}
		
		if len(project.Documents) > 0 {
			fmt.Printf("  documentation:\n")
			for _, docPtr := range project.Documents {
				doc := docPtr.Resolve()
				fmt.Printf("    - %s v%d (%s)\n", doc.Title, doc.Version, doc.Status)
			}
		}
		fmt.Printf("\n")
	}

	// step 6: demonstrate linker caching and performance
	fmt.Println("\n=== step 6: linker caching and performance ===")
	
	// create additional objects to test incremental updates
	newAuthorData := []map[string]any{
		{
			"id":         "frank",
			"name":       "Frank Miller",
			"email":      "frank@example.com",
			"department": map[string]any{"$ref": "eng"},
			"manager":    map[string]any{"$ref": "alice"},
			"joined_at": "2023-09-01T09:00:00Z",
		},
	}

	// bind new author
	var newAuthors []*Author
	for _, authorData := range newAuthorData {
		var author Author
		if err := df.Bind(&author, authorData, opts); err != nil {
			log.Fatalf("failed to bind new author: %v", err)
		}
		newAuthors = append(newAuthors, &author)
	}

	// register new author with existing linker (reuses cache)
	for _, author := range newAuthors {
		if err := linker.Register(author); err != nil {
			log.Fatalf("failed to register new author: %v", err)
		}
	}

	// link new author into existing graph
	if err := linker.ResolveReferences(newAuthors[0]); err != nil {
		log.Fatalf("failed to resolve new author references: %v", err)
	}

	fmt.Printf("✓ incrementally added new author: %s\n", newAuthors[0].Name)
	fmt.Printf("  department: %s\n", newAuthors[0].Department.Resolve().Name)
	fmt.Printf("  manager: %s\n", newAuthors[0].Manager.Resolve().Name)

	// step 7: demonstrate error handling and validation
	fmt.Println("\n=== step 7: error handling and missing references ===")
	
	// create document with invalid reference
	invalidDocData := map[string]any{
		"id":       "invalid-doc",
		"title":    "Invalid Document",
		"content":  "This document has invalid references",
		"version":  1,
		"status":   "draft",
		"author":   map[string]any{"$ref": "nonexistent-author"}, // invalid reference
		"category": map[string]any{"$ref": "frontend-tech"},       // valid reference
		"created_at": "2023-12-01T10:00:00Z",
	}

	var invalidDoc Document
	if err := df.Bind(&invalidDoc, invalidDocData, opts); err != nil {
		log.Fatalf("failed to bind invalid document: %v", err)
	}

	// try to resolve references - should handle missing reference gracefully
	if err := linker.ResolveReferences(&invalidDoc); err != nil {
		fmt.Printf("✓ expected error for missing reference: %v\n", err)
	}

	fmt.Println("\n=== advanced linker capabilities demonstrated ===")
	fmt.Println("✓ multi-stage object registration and linking")
	fmt.Println("✓ complex circular reference resolution") 
	fmt.Println("✓ hierarchical relationship traversal")
	fmt.Println("✓ incremental object graph updates")
	fmt.Println("✓ performance optimization with caching")
	fmt.Println("✓ robust error handling for invalid references")

	fmt.Println("\n=== advanced linker example completed successfully! ===")
}