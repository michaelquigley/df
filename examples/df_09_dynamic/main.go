package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df"
)

// EmailAction sends email notifications
type EmailAction struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool `df:"html"`
}

func (e EmailAction) Type() string { return "email" }

func (e EmailAction) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":    "email",
		"to":      e.To,
		"subject": e.Subject,
		"body":    e.Body,
		"html":    e.IsHTML,
	}, nil
}

func (e EmailAction) Execute() string {
	format := "text"
	if e.IsHTML {
		format = "HTML"
	}
	return fmt.Sprintf("sending %s email to %s: %s", format, e.To, e.Subject)
}

// SlackAction sends slack messages
type SlackAction struct {
	Channel string
	Message string
	Urgent  bool
}

func (s SlackAction) Type() string { return "slack" }

func (s SlackAction) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":    "slack",
		"channel": s.Channel,
		"message": s.Message,
		"urgent":  s.Urgent,
	}, nil
}

func (s SlackAction) Execute() string {
	urgency := ""
	if s.Urgent {
		urgency = " [URGENT]"
	}
	return fmt.Sprintf("posting to #%s%s: %s", s.Channel, urgency, s.Message)
}

// WebhookAction triggers HTTP webhooks
type WebhookAction struct {
	URL     string
	Method  string
	Payload string
}

func (w WebhookAction) Type() string { return "webhook" }

func (w WebhookAction) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":    "webhook",
		"url":     w.URL,
		"method":  w.Method,
		"payload": w.Payload,
	}, nil
}

func (w WebhookAction) Execute() string {
	return fmt.Sprintf("calling %s %s with payload: %.50s...", w.Method, w.URL, w.Payload)
}

// NotificationRule contains a polymorphic action field
type NotificationRule struct {
	Name      string
	Trigger   string
	Enabled   bool
	Action    df.Dynamic    // polymorphic field
	Fallback  df.Dynamic    // optional polymorphic field
}

// Executor interface for demonstration
type Executor interface {
	Execute() string
}

func main() {
	fmt.Println("=== df.Dynamic polymorphic types example ===")
	fmt.Println("demonstrates runtime type discrimination and flexible data binding")
	fmt.Println("using the df.Dynamic interface for polymorphic data structures")

	// step 1: register type mappings for dynamic binding
	fmt.Println("\n=== step 1: registering dynamic type binders ===")
	opts := &df.Options{
		DynamicBinders: map[string]func(map[string]any) (df.Dynamic, error){
			"email": func(m map[string]any) (df.Dynamic, error) {
				return df.New[EmailAction](m)
			},
			"slack": func(m map[string]any) (df.Dynamic, error) {
				return df.New[SlackAction](m)
			},
			"webhook": func(m map[string]any) (df.Dynamic, error) {
				return df.New[WebhookAction](m)
			},
		},
	}
	fmt.Printf("✓ registered 3 dynamic types: email, slack, webhook\n")

	// step 2: bind data with different action types
	fmt.Println("\n=== step 2: binding polymorphic notification rules ===")
	rulesData := map[string]any{
		"rules": []any{
			map[string]any{
				"name":    "user signup notification",
				"trigger": "user.signup",
				"enabled": true,
				"action": map[string]any{
					"type":    "email",
					"to":      "admin@example.com",
					"subject": "new user signup",
					"body":    "a new user has signed up",
					"html":    false,
				},
			},
			map[string]any{
				"name":    "system alert",
				"trigger": "system.error",
				"enabled": true,
				"action": map[string]any{
					"type":    "slack",
					"channel": "alerts",
					"message": "system error detected",
					"urgent":  true,
				},
				"fallback": map[string]any{
					"type":    "webhook",
					"url":     "https://api.pagerduty.com/incidents",
					"method":  "POST",
					"payload": `{"incident": {"type": "system_error", "priority": "high"}}`,
				},
			},
			map[string]any{
				"name":    "deployment notification",
				"trigger": "deploy.success",
				"enabled": false,
				"action": map[string]any{
					"type":    "webhook",
					"url":     "https://hooks.slack.com/services/...",
					"method":  "POST",
					"payload": `{"text": "deployment completed successfully"}`,
				},
			},
		},
	}

	type RulesContainer struct {
		Rules []NotificationRule
	}

	var container RulesContainer
	if err := df.Bind(&container, rulesData, opts); err != nil {
		log.Fatalf("failed to bind rules: %v", err)
	}

	fmt.Printf("✓ bound %d notification rules with polymorphic actions\n", len(container.Rules))

	// step 3: demonstrate type-safe access to concrete types
	fmt.Println("\n=== step 3: executing actions with type safety ===")
	for i, rule := range container.Rules {
		fmt.Printf("\nrule %d: %s (trigger: %s, enabled: %t)\n", i+1, rule.Name, rule.Trigger, rule.Enabled)
		fmt.Printf("  action type: %s\n", rule.Action.Type())
		
		// type-safe access to concrete implementation
		if executor, ok := rule.Action.(Executor); ok {
			if rule.Enabled {
				fmt.Printf("  → %s\n", executor.Execute())
			} else {
				fmt.Printf("  → (disabled, would execute: %s)\n", executor.Execute())
			}
		}

		// handle fallback actions
		if rule.Fallback != nil {
			fmt.Printf("  fallback type: %s\n", rule.Fallback.Type())
			if executor, ok := rule.Fallback.(Executor); ok {
				fmt.Printf("  → fallback: %s\n", executor.Execute())
			}
		}
	}

	// step 4: demonstrate unbinding with type information preserved
	fmt.Println("\n=== step 4: unbinding preserves dynamic type information ===")
	unboundData, err := df.Unbind(container)
	if err != nil {
		log.Fatalf("failed to unbind: %v", err)
	}

	// show that type information is preserved
	rules := unboundData["rules"].([]any)
	for i, ruleData := range rules {
		rule := ruleData.(map[string]any)
		action := rule["action"].(map[string]any)
		fmt.Printf("rule %d action type: %s (preserved in unbind)\n", i+1, action["type"])
	}

	// step 5: demonstrate field-specific binders (advanced feature)
	fmt.Println("\n=== step 5: field-specific dynamic binders ===")
	fmt.Println("note: field-specific binders would override global binders")
	fmt.Println("for specific struct fields, enabling per-field type mappings")

	fmt.Println("\n=== use cases for df.Dynamic ===")
	fmt.Println("✓ configuration systems with varying section schemas")
	fmt.Println("✓ plugin architectures with runtime-loaded types")
	fmt.Println("✓ API clients handling different response types")
	fmt.Println("✓ workflow engines processing different action types")
	fmt.Println("✓ extensible applications requiring runtime type flexibility")

	fmt.Println("\n=== dynamic types example completed successfully! ===")
}