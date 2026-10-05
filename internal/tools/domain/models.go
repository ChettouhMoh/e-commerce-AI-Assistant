// Package domain contains the core domain models for the tools system.
package toolsdomain

import "encoding/json"

// ToolDefinition describes a tool that the LLM can call.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ParameterDef
	Required    []string
}

// ParameterDef describes a single parameter of a tool.
type ParameterDef struct {
	Type        string
	Description string
	Enum        []string
}

// ToolArgument represents a single argument passed to a tool.
type ToolArgument struct {
	Name  string
	Value any
}

// ToolCallRequest represents a request to execute a tool.
type ToolCallRequest struct {
	CallID string
	Tool   string
	Args   map[string]any
}

// ToolResultData represents the data returned by a tool execution.
type ToolResultData struct {
	Content string
	IsError bool
}

// JSONSchema returns the JSON Schema representation of the tool definition.
func (t ToolDefinition) JSONSchema() map[string]any {
	props := make(map[string]any)
	for k, v := range t.Parameters {
		prop := map[string]any{
			"type":        v.Type,
			"description": v.Description,
		}
		if len(v.Enum) > 0 {
			prop["enum"] = v.Enum
		}
		props[k] = prop
	}
	return map[string]any{
		"type":       "object",
		"properties": props,
		"required":   t.Required,
	}
}

// MustMarshal is a convenience function that marshals a value to a JSON string.
func MustMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
