package main

import (
	"fmt"
	"log"
	"time"

	"github.com/michaelquigley/df/dd"
)

// step types for workflow steps
type ProcessStep struct {
	Command   string
	Arguments []string
	Timeout   time.Duration
}

func (p ProcessStep) Type() string { return "process" }
func (p ProcessStep) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":      "process",
		"command":   p.Command,
		"arguments": p.Arguments,
		"timeout":   p.Timeout.String(),
	}, nil
}

func (p ProcessStep) Execute() string {
	return fmt.Sprintf("executing: %s %v (timeout: %s)", p.Command, p.Arguments, p.Timeout)
}

type DecisionStep struct {
	Condition string
	TruePath  string
	FalsePath string
}

func (d DecisionStep) Type() string { return "decision" }
func (d DecisionStep) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":       "decision",
		"condition":  d.Condition,
		"true_path":  d.TruePath,
		"false_path": d.FalsePath,
	}, nil
}

func (d DecisionStep) Execute() string {
	return fmt.Sprintf("decision: %s → true:%s, false:%s", d.Condition, d.TruePath, d.FalsePath)
}

type NotificationStep struct {
	Recipients []string
	Subject    string
	Template   string
}

func (n NotificationStep) Type() string { return "notification" }
func (n NotificationStep) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":       "notification",
		"recipients": n.Recipients,
		"subject":    n.Subject,
		"template":   n.Template,
	}, nil
}

func (n NotificationStep) Execute() string {
	return fmt.Sprintf("notification: %s to %v using %s", n.Subject, n.Recipients, n.Template)
}

// action types for step actions (different from steps)
type LogAction struct {
	Level   string
	Message string
}

func (l LogAction) Type() string { return "log" }
func (l LogAction) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":    "log",
		"level":   l.Level,
		"message": l.Message,
	}, nil
}

func (l LogAction) Execute() string {
	return fmt.Sprintf("log[%s]: %s", l.Level, l.Message)
}

type MetricAction struct {
	Name  string
	Value float64
}

func (m MetricAction) Type() string { return "metric" }
func (m MetricAction) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":  "metric",
		"name":  m.Name,
		"value": m.Value,
	}, nil
}

func (m MetricAction) Execute() string {
	return fmt.Sprintf("metric: %s = %.2f", m.Name, m.Value)
}

type AlertAction struct {
	Severity string
	Message  string
	Channel  string
}

func (a AlertAction) Type() string { return "alert" }
func (a AlertAction) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":     "alert",
		"severity": a.Severity,
		"message":  a.Message,
		"channel":  a.Channel,
	}, nil
}

func (a AlertAction) Execute() string {
	return fmt.Sprintf("alert[%s]: %s → %s", a.Severity, a.Message, a.Channel)
}

// condition types for triggers
type TimeCondition struct {
	Schedule string
	Timezone string
}

func (t TimeCondition) Type() string { return "time" }
func (t TimeCondition) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":     "time",
		"schedule": t.Schedule,
		"timezone": t.Timezone,
	}, nil
}

func (t TimeCondition) Evaluate() string {
	return fmt.Sprintf("time condition: %s (%s)", t.Schedule, t.Timezone)
}

type EventCondition struct {
	EventType string
	Source    string
}

func (e EventCondition) Type() string { return "event" }
func (e EventCondition) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":       "event",
		"event_type": e.EventType,
		"source":     e.Source,
	}, nil
}

func (e EventCondition) Evaluate() string {
	return fmt.Sprintf("event condition: %s from %s", e.EventType, e.Source)
}

// transformer types for data processing
type JSONTransformer struct {
	JSONPath string
	Default  string
}

func (j JSONTransformer) Type() string { return "json" }
func (j JSONTransformer) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":      "json",
		"json_path": j.JSONPath,
		"default":   j.Default,
	}, nil
}

func (j JSONTransformer) Transform() string {
	return fmt.Sprintf("json transform: %s (default: %s)", j.JSONPath, j.Default)
}

type RegexTransformer struct {
	Pattern     string
	Replacement string
	Global      bool
}

func (r RegexTransformer) Type() string { return "regex" }
func (r RegexTransformer) ToMap() (map[string]any, error) {
	return map[string]any{
		"type":        "regex",
		"pattern":     r.Pattern,
		"replacement": r.Replacement,
		"global":      r.Global,
	}, nil
}

func (r RegexTransformer) Transform() string {
	scope := "first"
	if r.Global {
		scope = "global"
	}
	return fmt.Sprintf("regex transform: s/%s/%s/ (%s)", r.Pattern, r.Replacement, scope)
}

// main workflow structure
type WorkflowDefinition struct {
	Name         string
	Steps        []dd.Dynamic // workflow steps (process, decision, notification)
	Actions      []dd.Dynamic // step actions (log, metric, alert)
	Triggers     []dd.Dynamic // trigger conditions (time, event)
	Transformers []dd.Dynamic // data transformers (json, regex)
}

// interfaces for type safety
type StepExecutor interface {
	Execute() string
}

type ActionExecutor interface {
	Execute() string
}

type ConditionEvaluator interface {
	Evaluate() string
}

type DataTransformer interface {
	Transform() string
}

func main() {
	fmt.Println("=== df field-specific dynamic binders example ===")
	fmt.Println("demonstrates per-field polymorphic type control using FieldDynamicBinders")
	fmt.Println("different struct fields use different sets of dynamic type mappings")

	// step 1: set up field-specific binders
	fmt.Println("\n=== step 1: configuring field-specific binders ===")
	opts := &dd.Options{
		FieldDynamicBinders: map[string]map[string]func(map[string]any) (dd.Dynamic, error){
			// workflow steps have their own type namespace
			"WorkflowDefinition.Steps": {
				"process": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[ProcessStep](m)
				},
				"decision": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[DecisionStep](m)
				},
				"notification": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[NotificationStep](m)
				},
			},
			// step actions have a different type namespace
			"WorkflowDefinition.Actions": {
				"log": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[LogAction](m)
				},
				"metric": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[MetricAction](m)
				},
				"alert": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[AlertAction](m)
				},
			},
			// trigger conditions have their own namespace
			"WorkflowDefinition.Triggers": {
				"time": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[TimeCondition](m)
				},
				"event": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[EventCondition](m)
				},
			},
			// data transformers have their own namespace
			"WorkflowDefinition.Transformers": {
				"json": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[JSONTransformer](m)
				},
				"regex": func(m map[string]any) (dd.Dynamic, error) {
					return dd.New[RegexTransformer](m)
				},
			},
		},
	}

	fmt.Printf("✓ configured field-specific binders for 4 different field namespaces\n")
	fmt.Printf("  - steps: process, decision, notification\n")
	fmt.Printf("  - actions: log, metric, alert\n")
	fmt.Printf("  - triggers: time, event\n")
	fmt.Printf("  - transformers: json, regex\n")

	// step 2: demonstrate polymorphic binding with field-specific types
	fmt.Println("\n=== step 2: binding workflow with field-specific types ===")
	workflowData := map[string]any{
		"name": "data processing workflow",
		"steps": []any{
			map[string]any{
				"type":      "process",
				"command":   "curl",
				"arguments": []string{"-X", "GET", "https://api.example.com/data"},
				"timeout":   "30s",
			},
			map[string]any{
				"type":       "decision",
				"condition":  "response.status == 200",
				"true_path":  "process_data",
				"false_path": "handle_error",
			},
			map[string]any{
				"type":       "notification",
				"recipients": []string{"admin@example.com", "ops@example.com"},
				"subject":    "workflow completed",
				"template":   "workflow_success",
			},
		},
		"actions": []any{
			map[string]any{
				"type":    "log",
				"level":   "info",
				"message": "workflow step completed",
			},
			map[string]any{
				"type":  "metric",
				"name":  "workflow.step.duration",
				"value": 1.23,
			},
			map[string]any{
				"type":     "alert",
				"severity": "warning",
				"message":  "high processing time detected",
				"channel":  "slack",
			},
		},
		"triggers": []any{
			map[string]any{
				"type":     "time",
				"schedule": "0 */6 * * *", // every 6 hours
				"timezone": "UTC",
			},
			map[string]any{
				"type":       "event",
				"event_type": "data.received",
				"source":     "api.gateway",
			},
		},
		"transformers": []any{
			map[string]any{
				"type":      "json",
				"json_path": "$.data.items[*].id",
				"default":   "unknown",
			},
			map[string]any{
				"type":        "regex",
				"pattern":     `\d{4}-\d{2}-\d{2}`,
				"replacement": "REDACTED",
				"global":      true,
			},
		},
	}

	var workflow WorkflowDefinition
	if err := dd.Bind(&workflow, workflowData, opts); err != nil {
		log.Fatalf("failed to bind workflow: %v", err)
	}

	fmt.Printf("✓ bound workflow: %s\n", workflow.Name)
	fmt.Printf("  steps: %d, actions: %d, triggers: %d, transformers: %d\n",
		len(workflow.Steps), len(workflow.Actions), len(workflow.Triggers), len(workflow.Transformers))

	// step 3: demonstrate type safety and execution
	fmt.Println("\n=== step 3: executing workflow components ===")

	fmt.Println("\nworkflow steps:")
	for i, step := range workflow.Steps {
		fmt.Printf("  step %d: %s\n", i+1, step.Type())
		if executor, ok := step.(StepExecutor); ok {
			fmt.Printf("    → %s\n", executor.Execute())
		}
	}

	fmt.Println("\nstep actions:")
	for i, action := range workflow.Actions {
		fmt.Printf("  action %d: %s\n", i+1, action.Type())
		if executor, ok := action.(ActionExecutor); ok {
			fmt.Printf("    → %s\n", executor.Execute())
		}
	}

	fmt.Println("\ntrigger conditions:")
	for i, trigger := range workflow.Triggers {
		fmt.Printf("  trigger %d: %s\n", i+1, trigger.Type())
		if evaluator, ok := trigger.(ConditionEvaluator); ok {
			fmt.Printf("    → %s\n", evaluator.Evaluate())
		}
	}

	fmt.Println("\ndata transformers:")
	for i, transformer := range workflow.Transformers {
		fmt.Printf("  transformer %d: %s\n", i+1, transformer.Type())
		if trans, ok := transformer.(DataTransformer); ok {
			fmt.Printf("    → %s\n", trans.Transform())
		}
	}

	// step 4: demonstrate field isolation and error handling
	fmt.Println("\n=== step 4: field isolation demonstration ===")
	fmt.Println("attempting to use step types in action fields...")

	// this should fail because "process" is not registered for actions
	invalidData := map[string]any{
		"name": "invalid workflow",
		"actions": []any{
			map[string]any{
				"type":    "process", // step type in action field
				"command": "invalid",
			},
		},
	}

	var invalidWorkflow WorkflowDefinition
	if err := dd.Bind(&invalidWorkflow, invalidData, opts); err != nil {
		fmt.Printf("✓ expected error: %v\n", err)
		fmt.Printf("  field isolation working correctly - step types not available in action fields\n")
	}

	// step 5: demonstrate unbinding with field-specific marshalers
	fmt.Println("\n=== step 5: unbinding with field-specific type information ===")
	unboundData, err := dd.Unbind(workflow)
	if err != nil {
		log.Fatalf("failed to unbind workflow: %v", err)
	}

	if steps, ok := unboundData["steps"].([]any); ok {
		fmt.Printf("unbound steps (%d):\n", len(steps))
		for i, step := range steps {
			if stepMap, ok := step.(map[string]any); ok {
				fmt.Printf("  step %d type: %s\n", i+1, stepMap["type"])
			}
		}
	}

	if actions, ok := unboundData["actions"].([]any); ok {
		fmt.Printf("unbound actions (%d):\n", len(actions))
		for i, action := range actions {
			if actionMap, ok := action.(map[string]any); ok {
				fmt.Printf("  action %d type: %s\n", i+1, actionMap["type"])
			}
		}
	}

	fmt.Println("\n=== key benefits of field-specific binders ===")
	fmt.Println("✓ namespace isolation: different fields support different type sets")
	fmt.Println("✓ type safety: compile-time safety with field-specific runtime flexibility")
	fmt.Println("✓ clean architecture: separate concerns by field rather than globally")
	fmt.Println("✓ extensible design: easily add new types to specific fields")
	fmt.Println("✓ precise control: fine-grained control over polymorphic behavior")

	fmt.Println("\n=== field-specific dynamic binders example completed successfully! ===")
}
