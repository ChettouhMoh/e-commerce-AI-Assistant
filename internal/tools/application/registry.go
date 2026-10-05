// Package application provides the tool registry implementation.
package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	toolsdomain "ecommerce-ai-assistant/internal/tools/domain"
)

// ToolHandler is a function that executes a tool with validated arguments.
type ToolHandler func(ctx context.Context, args map[string]any) (string, error)

// ToolDefinitionWithHandler pairs a domain.ToolDefinition with its handler.
type ToolDefinitionWithHandler struct {
	Definition toolsdomain.ToolDefinition
	Handler    ToolHandler
}

// Registry implements the ports.ToolRegistry interface with an allowlist,
// argument validation, and dispatch to registered handlers.
type Registry struct {
	definitions   map[string]ToolDefinitionWithHandler
	allowlist     map[string]struct{}
	catalogRepo   interface{}
	ordersRepo    interface{}
	knowledgeRepo interface{}
}

// NewRegistry creates a new tool registry from the given tool definitions.
// The allowlist restricts which tools can be executed; pass nil to allow all.
func NewRegistry(tools []ToolDefinitionWithHandler, allowlist []string) *Registry {
	defs := make(map[string]ToolDefinitionWithHandler, len(tools))
	allowed := make(map[string]struct{}, len(allowlist))
	for _, t := range tools {
		defs[t.Definition.Name] = t
	}
	for _, name := range allowlist {
		allowed[strings.TrimSpace(name)] = struct{}{}
	}
	return &Registry{definitions: defs, allowlist: allowed}
}

// WithCatalogRepo sets the catalog repository for use by tool handlers.
func (r *Registry) WithCatalogRepo(repo interface{}) *Registry {
	r.catalogRepo = repo
	return r
}

// WithOrdersRepo sets the orders repository for use by tool handlers.
func (r *Registry) WithOrdersRepo(repo interface{}) *Registry {
	r.ordersRepo = repo
	return r
}

// WithKnowledgeRepo sets the knowledge repository for use by tool handlers.
func (r *Registry) WithKnowledgeRepo(repo interface{}) *Registry {
	r.knowledgeRepo = repo
	return r
}

// List returns all available tool definitions.
func (r *Registry) List() []toolsdomain.ToolDefinition {
	out := make([]toolsdomain.ToolDefinition, 0, len(r.definitions))
	for _, t := range r.definitions {
		out = append(out, t.Definition)
	}
	return out
}

// Execute validates arguments, checks the allowlist, and dispatches to the handler.
func (r *Registry) Execute(ctx context.Context, call toolsdomain.ToolCallRequest) (toolsdomain.ToolResultData, error) {
	tool, ok := r.definitions[call.Tool]
	if !ok {
		return toolsdomain.ToolResultData{Content: fmt.Sprintf("unknown tool: %s", call.Tool), IsError: true}, fmt.Errorf("unknown tool: %s", call.Tool)
	}

	if len(r.allowlist) > 0 {
		if _, ok := r.allowlist[call.Tool]; !ok {
			return toolsdomain.ToolResultData{Content: "tool not allowed", IsError: true}, errors.New("tool not allowed")
		}
	}

	if err := r.ValidateArgs(tool.Definition, call.Args); err != nil {
		return toolsdomain.ToolResultData{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, fmt.Errorf("invalid arguments: %w", err)
	}

	handlerCtx := ctx
	if r.catalogRepo != nil {
		handlerCtx = context.WithValue(handlerCtx, "catalog_repo", r.catalogRepo)
	}
	if r.ordersRepo != nil {
		handlerCtx = context.WithValue(handlerCtx, "orders_repo", r.ordersRepo)
	}
	if r.knowledgeRepo != nil {
		handlerCtx = context.WithValue(handlerCtx, "knowledge_repo", r.knowledgeRepo)
	}

	result, err := tool.Handler(handlerCtx, call.Args)
	if err != nil {
		return toolsdomain.ToolResultData{Content: fmt.Sprintf("tool execution error: %v", err), IsError: true}, err
	}

	return toolsdomain.ToolResultData{Content: result, IsError: false}, nil
}

// ValidateArgs checks that all required arguments are present.
func (r *Registry) ValidateArgs(def toolsdomain.ToolDefinition, args map[string]any) error {
	if args == nil {
		args = make(map[string]any)
	}

	missing := make([]string, 0, len(def.Required))
	for _, req := range def.Required {
		if _, ok := args[req]; !ok {
			missing = append(missing, req)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required arguments: %s", strings.Join(missing, ", "))
	}

	for key, val := range args {
		if _, defined := def.Parameters[key]; !defined {
			return fmt.Errorf("unknown argument: %s", key)
		}

		paramDef, ok := def.Parameters[key]
		if !ok {
			continue
		}

		switch paramDef.Type {
		case "string":
			if _, ok := val.(string); !ok {
				return fmt.Errorf("argument %s must be a string, got %T", key, val)
			}
		case "number", "integer":
			switch val.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				// valid
			default:
				return fmt.Errorf("argument %s must be a number, got %T", key, val)
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("argument %s must be a boolean, got %T", key, val)
			}
		case "array":
			if _, ok := val.([]any); !ok {
				return fmt.Errorf("argument %s must be an array, got %T", key, val)
			}
		case "object":
			if _, ok := val.(map[string]any); !ok {
				return fmt.Errorf("argument %s must be an object, got %T", key, val)
			}
		}

		if len(paramDef.Enum) > 0 {
			strVal := fmt.Sprintf("%v", val)
			valid := false
			for _, e := range paramDef.Enum {
				if strVal == e {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("argument %s must be one of %v, got %v", key, paramDef.Enum, val)
			}
		}
	}

	return nil
}

// MustGetString extracts a string argument or returns an error.
func MustGetString(args map[string]any, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument: %s", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("argument %s must be a string, got %T", key, val)
	}
	return str, nil
}

// MustGetInt extracts an integer argument or returns an error.
func MustGetInt(args map[string]any, key string) (int, error) {
	val, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument: %s", key)
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	default:
		return 0, fmt.Errorf("argument %s must be an integer, got %T", key, val)
	}
}

// MustGetFloat extracts a float64 argument or returns an error.
func MustGetFloat(args map[string]any, key string) (float64, error) {
	val, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument: %s", key)
	}
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("argument %s must be a number, got %T", key, val)
	}
}

// MustGetBool extracts a bool argument or returns an error.
func MustGetBool(args map[string]any, key string) (bool, error) {
	val, ok := args[key]
	if !ok {
		return false, fmt.Errorf("missing required argument: %s", key)
	}
	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("argument %s must be a boolean, got %T", key, val)
	}
	return b, nil
}

// GetString extracts an optional string argument; returns empty string if absent.
func GetString(args map[string]any, key string) string {
	val, ok := args[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

// GetInt extracts an optional integer argument; returns 0 if absent or invalid.
func GetInt(args map[string]any, key string) int {
	val, ok := args[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

// GetFloat extracts an optional float64 argument; returns 0 if absent or invalid.
func GetFloat(args map[string]any, key string) float64 {
	val, ok := args[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// BuildToolDefinitions returns the toolsdomain.ToolDefinition slice for all tools.
func BuildToolDefinitions() []toolsdomain.ToolDefinition {
	return []toolsdomain.ToolDefinition{
		SearchProductsDefinition(),
		GetProductDefinition(),
		CheckInventoryDefinition(),
		GetOrderStatusDefinition(),
		GetCustomerOrdersDefinition(),
		SearchPolicyKnowledgeDefinition(),
	}
}
