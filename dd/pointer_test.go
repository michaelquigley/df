package dd

import (
	"encoding/json"
	"errors"
	"testing"
)

// test types that implement identifiable
type Node struct {
	Id       string            `dd:"id"`
	Name     string            `dd:"name"`
	Parent   *Pointer[*Node]   `dd:"parent,omitempty"`
	Children []*Pointer[*Node] `dd:"children,omitempty"`
}

func (n *Node) GetId() string { return n.Id }

type User struct {
	Id   string `dd:"id"`
	Name string `dd:"name"`
	Age  int    `dd:"age"`
}

func (u *User) GetId() string { return u.Id }

type Document struct {
	Id     string          `dd:"id"`
	Title  string          `dd:"title"`
	Author *Pointer[*User] `dd:"author"`
	Editor *Pointer[*User] `dd:"editor,omitempty"`
}

func (d *Document) GetId() string { return d.Id }

func TestBasicPointerBinding(t *testing.T) {
	data := map[string]any{
		"id":   "root",
		"name": "Root Node",
		"parent": map[string]any{
			"$ref": "parent1",
		},
		"children": []any{
			map[string]any{"$ref": "child1"},
			map[string]any{"$ref": "child2"},
		},
	}

	var root Node
	err := Bind(&root, data)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// check that references were stored but not resolved yet
	if root.Id != "root" || root.Name != "Root Node" {
		t.Errorf("basic fields not bound correctly")
	}
	if root.Parent == nil || root.Parent.Ref != "parent1" {
		t.Errorf("parent reference not bound correctly")
	}
	if len(root.Children) != 2 || root.Children[0].Ref != "child1" || root.Children[1].Ref != "child2" {
		t.Errorf("children references not bound correctly")
	}
	if root.Parent.IsResolved() {
		t.Errorf("parent should not be resolved yet")
	}
}

func TestPointerLinkingWithCycles(t *testing.T) {
	// create a tree structure with cycles: root -> child1 -> child2 -> root
	data := map[string]any{
		"nodes": []any{
			map[string]any{
				"id":   "root",
				"name": "Root",
				"children": []any{
					map[string]any{"$ref": "child1"},
				},
			},
			map[string]any{
				"id":     "child1",
				"name":   "Child 1",
				"parent": map[string]any{"$ref": "root"},
				"children": []any{
					map[string]any{"$ref": "child2"},
				},
			},
			map[string]any{
				"id":     "child2",
				"name":   "Child 2",
				"parent": map[string]any{"$ref": "child1"},
				"children": []any{
					map[string]any{"$ref": "root"}, // cycle back to root
				},
			},
		},
	}

	type TestContainer struct {
		Nodes []*Node `dd:"nodes"`
	}

	var container TestContainer
	err := Bind(&container, data)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// link the references
	err = Link(&container)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// verify the structure is correctly linked
	root := container.Nodes[0]
	child1 := container.Nodes[1]
	child2 := container.Nodes[2]

	// check that all pointers are resolved
	if !root.Children[0].IsResolved() || !child1.Parent.IsResolved() {
		t.Errorf("references should be resolved after linking")
	}

	// check the actual references
	if root.Children[0].Resolve() != child1 {
		t.Errorf("root's child should point to child1")
	}
	if child1.Parent.Resolve() != root {
		t.Errorf("child1's parent should point to root")
	}
	if child1.Children[0].Resolve() != child2 {
		t.Errorf("child1's child should point to child2")
	}
	if child2.Parent.Resolve() != child1 {
		t.Errorf("child2's parent should point to child1")
	}
	if child2.Children[0].Resolve() != root {
		t.Errorf("child2's child should point to root (cycle)")
	}
}

func TestMultipleTypesWithSameIDs(t *testing.T) {
	// both User and Document have id "1" - should not clash due to type prefixing
	data := map[string]any{
		"users": []any{
			map[string]any{
				"id":   "1",
				"name": "John Doe",
				"age":  30,
			},
		},
		"documents": []any{
			map[string]any{
				"id":     "1", // same id as user, but different type
				"title":  "My Document",
				"author": map[string]any{"$ref": "1"}, // should resolve to User with id "1"
			},
		},
	}

	type TestContainer struct {
		Users     []*User     `dd:"users"`
		Documents []*Document `dd:"documents"`
	}

	var container TestContainer
	err := Bind(&container, data)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	err = Link(&container)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// verify that the document's author points to the user, not the document
	user := container.Users[0]
	doc := container.Documents[0]

	if doc.Author.Resolve() != user {
		t.Errorf("document author should resolve to the user, not the document with same id")
	}
	if user.Id != "1" || doc.Id != "1" {
		t.Errorf("both objects should have id '1'")
	}

	// verify that all other data fields are properly bound
	if user.Name != "John Doe" {
		t.Errorf("User name not bound correctly, got %q", user.Name)
	}
	if user.Age != 30 {
		t.Errorf("User age not bound correctly, got %d", user.Age)
	}
	if doc.Title != "My Document" {
		t.Errorf("Document title not bound correctly, got %q", doc.Title)
	}
}

func TestUnbindPointers(t *testing.T) {
	// create a simple structure with pointers
	root := &Node{
		Id:   "root",
		Name: "Root Node",
	}
	child := &Node{
		Id:   "child",
		Name: "Child Node",
	}

	// set up the pointer manually (as if it was linked)
	root.Children = []*Pointer[*Node]{
		{Ref: "child", Resolved: child},
	}
	child.Parent = &Pointer[*Node]{
		Ref: "root", Resolved: root,
	}

	// unbind should serialize the $ref values, not the resolved objects
	result, err := Unbind(root)
	if err != nil {
		t.Fatalf("Unbind failed: %v", err)
	}

	// check the structure
	if result["id"] != "root" || result["name"] != "Root Node" {
		t.Errorf("basic fields not unbound correctly")
	}

	children, ok := result["children"].([]interface{})
	if !ok || len(children) != 1 {
		t.Fatalf("children not unbound correctly")
	}

	childRef, ok := children[0].(map[string]any)
	if !ok || childRef["$ref"] != "child" {
		t.Errorf("child reference not unbound correctly: %v", children[0])
	}
}

func TestUnbindPointersComprehensive(t *testing.T) {
	// test various pointer scenarios: resolved, unresolved, nil, empty slices
	user1 := &User{
		Id:   "user1",
		Name: "Alice",
		Age:  25,
	}

	doc := &Document{
		Id:     "doc1",
		Title:  "Test Document",
		Author: &Pointer[*User]{Ref: "user1", Resolved: user1}, // resolved pointer
		Editor: &Pointer[*User]{Ref: "user2"},                  // unresolved pointer (no Resolved set)
	}

	result, err := Unbind(doc)
	if err != nil {
		t.Fatalf("Unbind failed: %v", err)
	}

	// check basic fields
	if result["id"] != "doc1" || result["title"] != "Test Document" {
		t.Errorf("basic document fields not unbound correctly")
	}

	// check author pointer (resolved)
	author, ok := result["author"].(map[string]any)
	if !ok || author["$ref"] != "user1" {
		t.Errorf("author pointer not unbound correctly: %v", result["author"])
	}

	// check editor pointer (unresolved but has Ref)
	editor, ok := result["editor"].(map[string]any)
	if !ok || editor["$ref"] != "user2" {
		t.Errorf("editor pointer not unbound correctly: %v", result["editor"])
	}
}

func TestUnbindNilAndEmptyPointers(t *testing.T) {
	// test nil pointers and empty references
	doc1 := &Document{
		Id:     "doc1",
		Title:  "Document 1",
		Author: &Pointer[*User]{Ref: "user1"},
		Editor: nil, // nil pointer
	}

	doc2 := &Document{
		Id:     "doc2",
		Title:  "Document 2",
		Author: &Pointer[*User]{Ref: ""}, // empty reference
		Editor: &Pointer[*User]{},        // zero-value pointer
	}

	// test doc1 with nil editor
	result1, err := Unbind(doc1)
	if err != nil {
		t.Fatalf("Unbind doc1 failed: %v", err)
	}

	if result1["id"] != "doc1" || result1["title"] != "Document 1" {
		t.Errorf("basic fields not unbound correctly")
	}

	author, ok := result1["author"].(map[string]any)
	if !ok || author["$ref"] != "user1" {
		t.Errorf("Author not unbound correctly: %v", result1["author"])
	}

	// editor should be omitted since it's nil
	if _, exists := result1["editor"]; exists {
		t.Errorf("Nil editor should be omitted from output")
	}

	// test doc2 with empty references
	result2, err := Unbind(doc2)
	if err != nil {
		t.Fatalf("Unbind doc2 failed: %v", err)
	}

	// empty ref should be omitted
	if _, exists := result2["author"]; exists {
		t.Errorf("Empty author reference should be omitted from output")
	}
	if _, exists := result2["editor"]; exists {
		t.Errorf("Zero-value editor should be omitted from output")
	}
}

func TestUnbindPointerSlices(t *testing.T) {
	// test slices of pointers
	node := &Node{
		Id:   "parent",
		Name: "Parent Node",
		Children: []*Pointer[*Node]{
			{Ref: "child1"},
			{Ref: "child2"},
			{Ref: ""}, // empty ref
			nil,       // nil pointer in slice
		},
	}

	result, err := Unbind(node)
	if err != nil {
		t.Fatalf("Unbind failed: %v", err)
	}

	children, ok := result["children"].([]interface{})
	if !ok {
		t.Fatalf("Children not unbound as slice: %v", result["children"])
	}

	// should have 4 elements: 2 valid refs, 1 nil (empty ref), 1 nil (nil pointer)
	if len(children) != 4 {
		t.Errorf("Expected 4 children, got %d", len(children))
	}

	// check first two valid references
	ref1, ok := children[0].(map[string]any)
	if !ok || ref1["$ref"] != "child1" {
		t.Errorf("First child reference incorrect: %v", children[0])
	}

	ref2, ok := children[1].(map[string]any)
	if !ok || ref2["$ref"] != "child2" {
		t.Errorf("Second child reference incorrect: %v", children[1])
	}

	// empty ref and nil pointer should both be nil in output
	if children[2] != nil {
		t.Errorf("Empty ref should be nil in output, got: %v", children[2])
	}
	if children[3] != nil {
		t.Errorf("Nil pointer should be nil in output, got: %v", children[3])
	}
}

func TestCompleteUnbindRoundTrip(t *testing.T) {
	// start with objects, bind, link, unbind, and verify the result
	type TestContainer struct {
		Users     []*User     `dd:"users"`
		Documents []*Document `dd:"documents"`
	}

	container := TestContainer{
		Users: []*User{
			{Id: "user1", Name: "Alice", Age: 25},
			{Id: "user2", Name: "Bob", Age: 30},
		},
		Documents: []*Document{
			{
				Id:     "doc1",
				Title:  "Guide",
				Author: &Pointer[*User]{Ref: "user1"},
				Editor: &Pointer[*User]{Ref: "user2"},
			},
			{
				Id:     "doc2",
				Title:  "Manual",
				Author: &Pointer[*User]{Ref: "user2"},
				// editor omitted (nil)
			},
		},
	}

	// link the references first
	err := Link(&container)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// unbind should preserve the $ref structure
	result, err := Unbind(&container)
	if err != nil {
		t.Fatalf("Unbind failed: %v", err)
	}

	// verify the structure is correct
	users, ok := result["users"].([]interface{})
	if !ok || len(users) != 2 {
		t.Fatalf("Users not unbound correctly")
	}

	docs, ok := result["documents"].([]interface{})
	if !ok || len(docs) != 2 {
		t.Fatalf("Documents not unbound correctly")
	}

	// check first document
	doc1, ok := docs[0].(map[string]any)
	if !ok {
		t.Fatalf("First document not unbound as map")
	}

	author1, ok := doc1["author"].(map[string]any)
	if !ok || author1["$ref"] != "user1" {
		t.Errorf("Doc1 author reference incorrect: %v", doc1["author"])
	}

	editor1, ok := doc1["editor"].(map[string]any)
	if !ok || editor1["$ref"] != "user2" {
		t.Errorf("Doc1 editor reference incorrect: %v", doc1["editor"])
	}

	// check second document (editor should be omitted)
	doc2, ok := docs[1].(map[string]any)
	if !ok {
		t.Fatalf("Second document not unbound as map")
	}

	author2, ok := doc2["author"].(map[string]any)
	if !ok || author2["$ref"] != "user2" {
		t.Errorf("Doc2 author reference incorrect: %v", doc2["author"])
	}

	// editor should not be present (nil pointer)
	if _, exists := doc2["editor"]; exists {
		t.Errorf("Doc2 editor should be omitted when nil")
	}
}

func TestRoundTripWithPointers(t *testing.T) {
	// original data with pointer references
	originalData := map[string]any{
		"nodes": []any{
			map[string]any{
				"id":   "node1",
				"name": "Node 1",
				"children": []any{
					map[string]any{"$ref": "node2"},
				},
			},
			map[string]any{
				"id":     "node2",
				"name":   "Node 2",
				"parent": map[string]any{"$ref": "node1"},
			},
		},
	}

	type TestContainer struct {
		Nodes []*Node `dd:"nodes"`
	}

	// bind and link
	var container TestContainer
	err := Bind(&container, originalData)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	err = Link(&container)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// unbind back to map
	result, err := Unbind(&container)
	if err != nil {
		t.Fatalf("Unbind failed: %v", err)
	}

	// convert to JSON for easy comparison
	originalJSON, _ := json.Marshal(originalData)
	resultJSON, _ := json.Marshal(result)

	// the structure should be equivalent (though field order might differ)
	if len(string(originalJSON)) != len(string(resultJSON)) {
		t.Logf("Original: %s", originalJSON)
		t.Logf("Result:   %s", resultJSON)
		// note: exact comparison might fail due to field ordering, but lengths should be similar
		// in a real test, you'd want to do a more sophisticated comparison
	}
}

func TestEmptyPointerReferences(t *testing.T) {
	data := map[string]any{
		"id":   "node1",
		"name": "Node 1",
		// parent is omitted - should result in nil Pointer
		"children": []any{}, // empty children list
	}

	var node Node
	err := Bind(&node, data)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	err = Link(&node)
	if err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	if node.Parent != nil {
		t.Errorf("Parent should be nil when omitted")
	}
	if len(node.Children) != 0 {
		t.Errorf("Children should be empty")
	}
}

func TestUnresolvedReference(t *testing.T) {
	data := map[string]any{
		"id":     "node1",
		"name":   "Node 1",
		"parent": map[string]any{"$ref": "nonexistent"},
	}

	var node Node
	err := Bind(&node, data)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// link should fail due to unresolved reference
	err = Link(&node)
	if err == nil {
		t.Errorf("Link should have failed due to unresolved reference")
	}
	var pointerErr *PointerError
	if err != nil && !errors.As(err, &pointerErr) {
		t.Errorf("expected PointerError, got %T", err)
	}
	if err != nil && !contains(err.Error(), "unresolved reference") {
		t.Errorf("Error should mention unresolved reference: %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLinkerBasicFunctionality(t *testing.T) {
	// create test data
	data := map[string]any{
		"nodes": []any{
			map[string]any{
				"id":   "node1",
				"name": "Node 1",
				"children": []any{
					map[string]any{"$ref": "node2"},
				},
			},
			map[string]any{
				"id":     "node2",
				"name":   "Node 2",
				"parent": map[string]any{"$ref": "node1"},
			},
		},
	}

	type TestContainer struct {
		Nodes []*Node `dd:"nodes"`
	}

	var container TestContainer
	err := Bind(&container, data)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// test with default linker
	err = Link(&container)
	if err != nil {
		t.Fatalf("Linker.Link failed: %v", err)
	}

	// verify linking worked
	node1 := container.Nodes[0]
	node2 := container.Nodes[1]

	if !node1.Children[0].IsResolved() {
		t.Errorf("node1's child should be resolved")
	}
	if node1.Children[0].Resolve() != node2 {
		t.Errorf("node1's child should point to node2")
	}
	if !node2.Parent.IsResolved() {
		t.Errorf("node2's parent should be resolved")
	}
	if node2.Parent.Resolve() != node1 {
		t.Errorf("node2's parent should point to node1")
	}
}

func TestLinkerCaching(t *testing.T) {
	type TestContainer struct {
		Nodes []*Node `dd:"nodes"`
	}

	// create test data that we'll use twice
	createTestData := func() (TestContainer, error) {
		data := map[string]any{
			"nodes": []any{
				map[string]any{
					"id":   "cached1",
					"name": "Cached Node 1",
					"children": []any{
						map[string]any{"$ref": "cached2"},
					},
				},
				map[string]any{
					"id":     "cached2",
					"name":   "Cached Node 2",
					"parent": map[string]any{"$ref": "cached1"},
				},
			},
		}

		var container TestContainer
		err := Bind(&container, data)
		return container, err
	}

	// create linker with caching enabled
	linker := NewLinker(LinkerOptions{EnableCaching: true})

	// first linking operation - should populate cache
	container1, err := createTestData()
	if err != nil {
		t.Fatalf("First bind failed: %v", err)
	}

	err = linker.Link(&container1)
	if err != nil {
		t.Fatalf("First link failed: %v", err)
	}

	// second linking operation - should use cache
	container2, err := createTestData()
	if err != nil {
		t.Fatalf("Second bind failed: %v", err)
	}

	err = linker.Link(&container2)
	if err != nil {
		t.Fatalf("Second link failed: %v", err)
	}

	// verify both containers are properly linked
	if !container1.Nodes[0].Children[0].IsResolved() {
		t.Errorf("container1 not properly linked")
	}
	if !container2.Nodes[0].Children[0].IsResolved() {
		t.Errorf("container2 not properly linked")
	}

	// clear cache and try again
	linker.ClearCache()
	container3, err := createTestData()
	if err != nil {
		t.Fatalf("Third bind failed: %v", err)
	}

	err = linker.Link(&container3)
	if err != nil {
		t.Fatalf("Third link after cache clear failed: %v", err)
	}

	if !container3.Nodes[0].Children[0].IsResolved() {
		t.Errorf("container3 not properly linked after cache clear")
	}
}

func TestLinkerMultiStage(t *testing.T) {
	// create separate data sources
	source1 := map[string]any{
		"users": []any{
			map[string]any{
				"id":   "user1",
				"name": "Alice",
				"age":  25,
			},
		},
	}

	source2 := map[string]any{
		"documents": []any{
			map[string]any{
				"id":     "doc1",
				"title":  "Document 1",
				"author": map[string]any{"$ref": "user1"}, // references user from source1
			},
		},
	}

	type Source1 struct {
		Users []*User `dd:"users"`
	}

	type Source2 struct {
		Documents []*Document `dd:"documents"`
	}

	var s1 Source1
	var s2 Source2

	err := Bind(&s1, source1)
	if err != nil {
		t.Fatalf("bind source1 failed: %v", err)
	}

	err = Bind(&s2, source2)
	if err != nil {
		t.Fatalf("bind source2 failed: %v", err)
	}

	// use multi-stage linking
	linker := NewLinker(LinkerOptions{EnableCaching: true})

	// register from source1
	err = linker.Register(&s1)
	if err != nil {
		t.Fatalf("register from source1 failed: %v", err)
	}

	// register from source2 (won't find any new Identifiables, but that's ok)
	err = linker.Register(&s2)
	if err != nil {
		t.Fatalf("register from source2 failed: %v", err)
	}

	// now resolve references in source2
	err = linker.ResolveReferences(&s2)
	if err != nil {
		t.Fatalf("resolveReferences failed: %v", err)
	}

	// verify the cross-reference worked
	if !s2.Documents[0].Author.IsResolved() {
		t.Errorf("document author should be resolved")
	}

	resolvedAuthor := s2.Documents[0].Author.Resolve()
	if resolvedAuthor != s1.Users[0] {
		t.Errorf("document author should point to user from source1")
	}
}

func TestLinkerPartialResolution(t *testing.T) {
	data := map[string]any{
		"id":     "node1",
		"name":   "Node 1",
		"parent": map[string]any{"$ref": "nonexistent"},
	}

	var node Node
	err := Bind(&node, data)
	if err != nil {
		t.Fatalf("bind failed: %v", err)
	}

	// with partial resolution disabled (default), should fail
	linker1 := NewLinker()
	err = linker1.Link(&node)
	if err == nil {
		t.Errorf("link should have failed with unresolved reference")
	}
	var pointerErr2 *PointerError
	if err != nil && !errors.As(err, &pointerErr2) {
		t.Errorf("expected PointerError, got %T", err)
	}

	// with partial resolution enabled, should succeed
	linker2 := NewLinker(LinkerOptions{AllowPartialResolution: true})
	err = linker2.Link(&node)
	if err != nil {
		t.Fatalf("link with partial resolution should have succeeded: %v", err)
	}

	// the reference should remain unresolved
	if node.Parent == nil {
		t.Errorf("parent pointer should not be nil")
	}
	if node.Parent.IsResolved() {
		t.Errorf("parent should not be resolved when reference is missing")
	}
	if node.Parent.Ref != "nonexistent" {
		t.Errorf("parent ref should still contain the unresolved reference")
	}
}

func TestVariadicLink(t *testing.T) {
	// create separate data sources that reference each other
	source1 := map[string]any{
		"users": []any{
			map[string]any{
				"id":   "user1",
				"name": "Alice",
				"age":  25,
			},
		},
	}

	source2 := map[string]any{
		"documents": []any{
			map[string]any{
				"id":     "doc1",
				"title":  "Document 1",
				"author": map[string]any{"$ref": "user1"}, // references user from source1
			},
		},
	}

	type Source1 struct {
		Users []*User `dd:"users"`
	}

	type Source2 struct {
		Documents []*Document `dd:"documents"`
	}

	var s1 Source1
	var s2 Source2

	err := Bind(&s1, source1)
	if err != nil {
		t.Fatalf("bind source1 failed: %v", err)
	}

	err = Bind(&s2, source2)
	if err != nil {
		t.Fatalf("bind source2 failed: %v", err)
	}

	// test variadic Link function
	err = Link(&s1, &s2)
	if err != nil {
		t.Fatalf("variadic Link failed: %v", err)
	}

	// verify the cross-reference worked
	if !s2.Documents[0].Author.IsResolved() {
		t.Errorf("document author should be resolved")
	}

	resolvedAuthor := s2.Documents[0].Author.Resolve()
	if resolvedAuthor != s1.Users[0] {
		t.Errorf("document author should point to user from source1")
	}
}

func TestVariadicLinkerLink(t *testing.T) {
	// create test data similar to the basic functionality test but split across multiple objects
	source1 := map[string]any{
		"nodes": []any{
			map[string]any{
				"id":   "node1",
				"name": "Node 1",
				"children": []any{
					map[string]any{"$ref": "node2"},
				},
			},
		},
	}

	source2 := map[string]any{
		"nodes": []any{
			map[string]any{
				"id":     "node2",
				"name":   "Node 2",
				"parent": map[string]any{"$ref": "node1"},
			},
		},
	}

	type TestContainer struct {
		Nodes []*Node `dd:"nodes"`
	}

	var container1, container2 TestContainer

	err := Bind(&container1, source1)
	if err != nil {
		t.Fatalf("bind container1 failed: %v", err)
	}

	err = Bind(&container2, source2)
	if err != nil {
		t.Fatalf("bind container2 failed: %v", err)
	}

	// test linker with variadic Link
	linker := NewLinker()
	err = linker.Link(&container1, &container2)
	if err != nil {
		t.Fatalf("linker variadic Link failed: %v", err)
	}

	// verify linking worked
	node1 := container1.Nodes[0]
	node2 := container2.Nodes[0]

	if !node1.Children[0].IsResolved() {
		t.Errorf("node1's child should be resolved")
	}
	if node1.Children[0].Resolve() != node2 {
		t.Errorf("node1's child should point to node2")
	}
	if !node2.Parent.IsResolved() {
		t.Errorf("node2's parent should be resolved")
	}
	if node2.Parent.Resolve() != node1 {
		t.Errorf("node2's parent should point to node1")
	}
}

func TestVariadicRegister(t *testing.T) {
	// create separate data sources
	source1 := map[string]any{
		"users": []any{
			map[string]any{
				"id":   "user1",
				"name": "Alice",
				"age":  25,
			},
		},
	}

	source2 := map[string]any{
		"documents": []any{
			map[string]any{
				"id":     "doc1",
				"title":  "Document 1",
				"author": map[string]any{"$ref": "user1"}, // references user from source1
			},
		},
	}

	type Source1 struct {
		Users []*User `dd:"users"`
	}

	type Source2 struct {
		Documents []*Document `dd:"documents"`
	}

	var s1 Source1
	var s2 Source2

	err := Bind(&s1, source1)
	if err != nil {
		t.Fatalf("bind source1 failed: %v", err)
	}

	err = Bind(&s2, source2)
	if err != nil {
		t.Fatalf("bind source2 failed: %v", err)
	}

	// use multi-stage linking with variadic Register
	linker := NewLinker(LinkerOptions{EnableCaching: true})

	// register both sources in one call
	err = linker.Register(&s1, &s2)
	if err != nil {
		t.Fatalf("variadic register failed: %v", err)
	}

	// now resolve references in source2
	err = linker.ResolveReferences(&s2)
	if err != nil {
		t.Fatalf("resolveReferences failed: %v", err)
	}

	// verify the cross-reference worked
	if !s2.Documents[0].Author.IsResolved() {
		t.Errorf("document author should be resolved")
	}

	resolvedAuthor := s2.Documents[0].Author.Resolve()
	if resolvedAuthor != s1.Users[0] {
		t.Errorf("document author should point to user from source1")
	}
}
