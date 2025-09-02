---
title: Advanced Features
description: Explore Dynamic types, object references, custom converters, and advanced patterns in the df framework.
---

The df framework provides several advanced features for handling complex data structures and scenarios. This guide covers polymorphic data with Dynamic types, object references with cycle handling, custom converters, and sophisticated patterns.

## Dynamic Types (Polymorphic Data)

The `Dynamic` interface enables polymorphic data structures where the actual type is determined at runtime based on configuration.

### Dynamic Interface

```go
type Dynamic interface {
    Type() string
    ToMap() (map[string]any, error)
}
```

### Implementing Dynamic Types

Define concrete types that implement the Dynamic interface:

```go
// Email notification action
type EmailAction struct {
    Recipient string `df:"recipient"`
    Subject   string `df:"subject"`
    Body      string `df:"body"`
}

func (e EmailAction) Type() string { return "email" }
func (e EmailAction) ToMap() (map[string]any, error) {
    return map[string]any{
        "recipient": e.Recipient,
        "subject":   e.Subject,
        "body":      e.Body,
    }, nil
}

// SMS notification action
type SMSAction struct {
    PhoneNumber string `df:"phone_number"`
    Message     string `df:"message"`
}

func (s SMSAction) Type() string { return "sms" }
func (s SMSAction) ToMap() (map[string]any, error) {
    return map[string]any{
        "phone_number": s.PhoneNumber,
        "message":      s.Message,
    }, nil
}

// Webhook action
type WebhookAction struct {
    URL     string            `df:"url"`
    Method  string            `df:"method"`
    Headers map[string]string `df:"headers"`
}

func (w WebhookAction) Type() string { return "webhook" }
func (w WebhookAction) ToMap() (map[string]any, error) {
    return map[string]any{
        "url":     w.URL,
        "method":  w.Method,
        "headers": w.Headers,
    }, nil
}
```

### Using Dynamic Fields

Define structs with Dynamic fields:

```go
type Notification struct {
    ID       string    `df:"id"`
    Name     string    `df:"name"`
    Enabled  bool      `df:"enabled"`
    Action   Dynamic   `df:"action"`  // Polymorphic field
    Triggers []Dynamic `df:"triggers"` // Slice of polymorphic types
}
```

### Configuring Dynamic Binders

Configure type binders to create the correct type based on a discriminator:

```go
func main() {
    // Configuration data with type discriminator
    data := map[string]any{
        "id":      "notification-1",
        "name":    "User Welcome",
        "enabled": true,
        "action": map[string]any{
            "type":      "email",
            "recipient": "user@example.com",
            "subject":   "Welcome!",
            "body":      "Welcome to our service!",
        },
        "triggers": []any{
            map[string]any{
                "type":    "user_signup",
                "delay":   "5m",
            },
            map[string]any{
                "type":      "webhook",
                "url":       "https://api.example.com/trigger",
                "method":    "POST",
            },
        },
    }

    // Configure binders for different types
    opts := &df.Options{
        DynamicBinders: map[string]func(map[string]any) (df.Dynamic, error){
            "email": func(m map[string]any) (df.Dynamic, error) {
                action, err := df.New[EmailAction](m)
                if err != nil {
                    return nil, err
                }
                return *action, nil
            },
            "sms": func(m map[string]any) (df.Dynamic, error) {
                action, err := df.New[SMSAction](m)
                if err != nil {
                    return nil, err
                }
                return *action, nil
            },
            "webhook": func(m map[string]any) (df.Dynamic, error) {
                action, err := df.New[WebhookAction](m)
                if err != nil {
                    return nil, err
                }
                return *action, nil
            },
        },
    }

    // Bind with custom options
    notification, err := df.NewWithOptions[Notification](data, opts)
    if err != nil {
        panic(err)
    }

    // Use the polymorphic action
    switch action := notification.Action.(type) {
    case EmailAction:
        fmt.Printf("Sending email to %s: %s\n", action.Recipient, action.Subject)
    case SMSAction:
        fmt.Printf("Sending SMS to %s: %s\n", action.PhoneNumber, action.Message)
    case WebhookAction:
        fmt.Printf("Calling webhook %s %s\n", action.Method, action.URL)
    }
}
```

### Field-Specific Dynamic Binders

Configure different binders for different fields:

```go
type WorkflowStep struct {
    ID       string  `df:"id"`
    Name     string  `df:"name"`
    Action   Dynamic `df:"action"`
    Condition Dynamic `df:"condition"`
}

opts := &df.Options{
    FieldDynamicBinders: map[string]map[string]func(map[string]any) (df.Dynamic, error){
        "action": {
            "email":   createEmailAction,
            "sms":     createSMSAction,
            "webhook": createWebhookAction,
        },
        "condition": {
            "time":       createTimeCondition,
            "user_state": createUserStateCondition,
            "data_check": createDataCheckCondition,
        },
    },
}
```

## Object References

The `df.Pointer[T]` type provides object references with cycle handling and lazy resolution.

### Defining Referenceable Objects

Objects must implement the `Identifiable` interface:

```go
type Identifiable interface {
    GetId() string
}

type User struct {
    ID    string `df:"id"`
    Name  string `df:"name"`
    Email string `df:"email"`
}

func (u *User) GetId() string { return u.ID }

type Document struct {
    ID       string             `df:"id"`
    Title    string             `df:"title"`
    Content  string             `df:"content"`
    Author   *df.Pointer[*User] `df:"author"`   // Reference to User
    Reviewers []df.Pointer[*User] `df:"reviewers"` // Multiple references
}

func (d *Document) GetId() string { return d.ID }
```

### Two-Phase Binding Process

Object references require a two-phase process: bind then link.

```go
type DataContainer struct {
    Users     []User     `df:"users"`
    Documents []Document `df:"documents"`
}

func main() {
    data := map[string]any{
        "users": []any{
            map[string]any{
                "id":    "user-1",
                "name":  "John Doe",
                "email": "john@example.com",
            },
            map[string]any{
                "id":    "user-2", 
                "name":  "Jane Smith",
                "email": "jane@example.com",
            },
        },
        "documents": []any{
            map[string]any{
                "id":      "doc-1",
                "title":   "Important Document",
                "content": "This is important...",
                "author":  "$ref:user-1", // Reference by ID
                "reviewers": []any{
                    "$ref:user-1",
                    "$ref:user-2",
                },
            },
        },
    }

    // Phase 1: Bind data (references are stored as strings)
    var container DataContainer
    err := df.Bind(&container, data)
    if err != nil {
        panic(err)
    }

    // Phase 2: Link references (resolve string references to actual objects)
    err = df.Link(&container)
    if err != nil {
        panic(err)
    }

    // Now references are resolved
    doc := container.Documents[0]
    author := doc.Author.Resolve() // Returns *User
    if author != nil {
        fmt.Printf("Document author: %s\n", author.Name)
    }

    // Access reviewers
    for _, reviewerPtr := range doc.Reviewers {
        if reviewer := reviewerPtr.Resolve(); reviewer != nil {
            fmt.Printf("Reviewer: %s\n", reviewer.Name)
        }
    }
}
```

### Cycle Handling

df.Pointer handles circular references automatically:

```go
type Category struct {
    ID       string                  `df:"id"`
    Name     string                  `df:"name"`
    Parent   *df.Pointer[*Category]  `df:"parent"`
    Children []df.Pointer[*Category] `df:"children"`
}

func (c *Category) GetId() string { return c.ID }

// Data with circular references
data := map[string]any{
    "categories": []any{
        map[string]any{
            "id":       "cat-1",
            "name":     "Root Category",
            "children": []any{"$ref:cat-2", "$ref:cat-3"},
        },
        map[string]any{
            "id":     "cat-2",
            "name":   "Child Category",
            "parent": "$ref:cat-1",
        },
        map[string]any{
            "id":     "cat-3", 
            "name":   "Another Child",
            "parent": "$ref:cat-1",
        },
    },
}
```

### Caching and Performance

Pointer resolution is cached for performance:

```go
// First access resolves and caches
user := doc.Author.Resolve()

// Subsequent accesses use cached value
user2 := doc.Author.Resolve() // Same instance, no lookup
```

## Custom Converters

Implement custom type conversion logic with the `Converter` interface:

### Converter Interface

```go
type Converter interface {
    CanConvert(from, to reflect.Type) bool
    Convert(value any, to reflect.Type) (any, error)
}
```

### Example: UUID Converter

```go
import "github.com/google/uuid"

type UUIDConverter struct{}

func (c *UUIDConverter) CanConvert(from, to reflect.Type) bool {
    return to == reflect.TypeOf(uuid.UUID{})
}

func (c *UUIDConverter) Convert(value any, to reflect.Type) (any, error) {
    switch v := value.(type) {
    case string:
        return uuid.Parse(v)
    case []byte:
        return uuid.ParseBytes(v)
    default:
        return nil, fmt.Errorf("cannot convert %T to UUID", value)
    }
}

// Register converter
opts := &df.Options{
    Converters: []df.Converter{&UUIDConverter{}},
}

type User struct {
    ID   uuid.UUID `df:"id"`
    Name string    `df:"name"`
}

data := map[string]any{
    "id":   "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
}

user, err := df.NewWithOptions[User](data, opts)
```

### Example: Time Converter

```go
type FlexibleTimeConverter struct{}

func (c *FlexibleTimeConverter) CanConvert(from, to reflect.Type) bool {
    return to == reflect.TypeOf(time.Time{})
}

func (c *FlexibleTimeConverter) Convert(value any, to reflect.Type) (any, error) {
    switch v := value.(type) {
    case string:
        // Try multiple time formats
        formats := []string{
            time.RFC3339,
            time.RFC3339Nano,
            "2006-01-02 15:04:05",
            "2006-01-02",
        }
        for _, format := range formats {
            if t, err := time.Parse(format, v); err == nil {
                return t, nil
            }
        }
        return nil, fmt.Errorf("cannot parse time: %s", v)
    case int64:
        return time.Unix(v, 0), nil
    case float64:
        return time.Unix(int64(v), 0), nil
    default:
        return nil, fmt.Errorf("cannot convert %T to time.Time", value)
    }
}
```

## Custom Marshaler/Unmarshaler

Implement custom serialization with Marshaler and Unmarshaler interfaces:

### Unmarshaler Interface

```go
type Unmarshaler interface {
    UnmarshalDF(data map[string]any) error
}

type IPAddress net.IP

func (ip *IPAddress) UnmarshalDF(data map[string]any) error {
    value, ok := data["ip"]
    if !ok {
        return errors.New("ip field required")
    }
    
    str, ok := value.(string)
    if !ok {
        return errors.New("ip must be string")
    }
    
    parsed := net.ParseIP(str)
    if parsed == nil {
        return fmt.Errorf("invalid IP address: %s", str)
    }
    
    *ip = IPAddress(parsed)
    return nil
}
```

### Marshaler Interface

```go
type Marshaler interface {
    MarshalDF() (map[string]any, error)
}

func (ip IPAddress) MarshalDF() (map[string]any, error) {
    return map[string]any{
        "ip": net.IP(ip).String(),
    }, nil
}

// Usage
type ServerConfig struct {
    Name string    `df:"name"`
    IP   IPAddress `df:"ip"`
}
```

## Advanced Patterns

### Plugin Architecture with Dynamic Types

```go
type Plugin interface {
    df.Dynamic
    Execute(context map[string]any) (map[string]any, error)
}

type FilterPlugin struct {
    Field     string `df:"field"`
    Operation string `df:"operation"`
    Value     any    `df:"value"`
}

func (f FilterPlugin) Type() string { return "filter" }
func (f FilterPlugin) ToMap() (map[string]any, error) {
    return map[string]any{
        "field":     f.Field,
        "operation": f.Operation,
        "value":     f.Value,
    }, nil
}

func (f FilterPlugin) Execute(context map[string]any) (map[string]any, error) {
    // Filter implementation
    return context, nil
}

type Pipeline struct {
    Name    string   `df:"name"`
    Plugins []Plugin `df:"plugins"`
}
```

### Configuration Templates

```go
type Template struct {
    Name      string         `df:"name"`
    Variables map[string]any `df:"variables"`
    Template  string         `df:"template"`
}

func (t *Template) UnmarshalDF(data map[string]any) error {
    // Standard binding
    if err := df.Bind(t, data); err != nil {
        return err
    }
    
    // Template processing
    tmpl, err := template.New(t.Name).Parse(t.Template)
    if err != nil {
        return err
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, t.Variables); err != nil {
        return err
    }
    
    t.Template = buf.String()
    return nil
}
```

### Conditional Binding

```go
type ConditionalConfig struct {
    Environment string `df:"environment"`
    
    // Only bind these if environment is "production"
    Production *ProductionConfig `df:"production,condition=environment==production"`
    
    // Only bind these if environment is "development"  
    Development *DevelopmentConfig `df:"development,condition=environment==development"`
}

// Custom binding logic
func (c *ConditionalConfig) UnmarshalDF(data map[string]any) error {
    // Bind environment first
    if env, ok := data["environment"]; ok {
        c.Environment = env.(string)
    }
    
    // Conditional binding based on environment
    switch c.Environment {
    case "production":
        if prodData, ok := data["production"]; ok {
            prod, err := df.New[ProductionConfig](prodData)
            if err != nil {
                return err
            }
            c.Production = prod
        }
    case "development":
        if devData, ok := data["development"]; ok {
            dev, err := df.New[DevelopmentConfig](devData)
            if err != nil {
                return err
            }
            c.Development = dev
        }
    }
    
    return nil
}
```

## Best Practices

### Dynamic Types
- Use type discriminators consistently
- Keep Dynamic implementations simple  
- Validate type fields in constructors
- Handle unknown types gracefully

### Object References
- Always implement GetId() consistently
- Use meaningful ID formats
- Handle nil references safely
- Consider reference caching implications

### Custom Converters
- Implement CanConvert() accurately
- Provide detailed error messages
- Handle edge cases and nil values
- Register converters globally when possible

### Performance Considerations
- Use Pointer caching for frequently accessed references
- Minimize Dynamic type switching overhead
- Cache converter instances
- Profile binding performance for large datasets

## Next Steps

You've now mastered all aspects of the df framework:

- **[Getting Started](/guides/getting-started/)** - Review the basics
- **[Data Binding](/guides/data-binding/)** - Master struct binding and configuration
- **[Dependency Injection](/guides/dependency-injection/)** - Use containers for object management  
- **[Application Lifecycle](/guides/application-lifecycle/)** - Build complex applications with orchestration

Explore the [API Reference](/reference/df/) for detailed function documentation and check out the examples in the repository for complete working applications.